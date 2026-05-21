#include "themes.h"

#include <stdbool.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

#include "font_3x5.h"
#include "remote.h"
#include "vfd.h"

#define NUMERIC_CANVAS_WIDTH 7
#define NUMERIC_CANVAS_HEIGHT 5
#define NUMERIC_ROLL_FRAMES 5
#define NUMERIC_TARGET_HYSTERESIS 0.75
#define NUMERIC_TARGET_HOLD_FRAMES 10
#define NUMERIC_NET_EMA_ALPHA 0.05
#define NUMERIC_NET_TARGET_HYSTERESIS 1.5
#define NUMERIC_NET_TARGET_HOLD_FRAMES 40
#define NUMERIC_NET_COMMIT_FRAMES 12
#define NUMERIC_NET_MODE_UP_BPS 1250000.0
#define NUMERIC_NET_MODE_DOWN_BPS 750000.0
#define NUMERIC_ALERT_BLINK_PERIOD_MS 500
#define NUMERIC_ALERT_BLANK_MS 100

typedef struct {
    bool initialized;
    int digits[2];
    int next_digits[2];
    int directions[2];
    int frames[2];
} numeric_roll_state_t;

typedef struct {
    bool initialized;
    int target;
    int hold_frames;
} numeric_target_state_t;

typedef struct {
    bool initialized;
    double filtered_rate;
    bool fractional;
    int target;
    int direction;
    int hold_frames;
    bool candidate_active;
    bool candidate_fractional;
    int candidate_target;
    int candidate_frames;
} numeric_net_state_t;

static int numeric_clamp_int(int value, int min, int max) {
    if (value < min) return min;
    if (value > max) return max;
    return value;
}

static int numeric_round_clamped(double value, int min, int max) {
    if (value < 0.0)
        return numeric_clamp_int((int)(value - 0.5), min, max);
    return numeric_clamp_int((int)(value + 0.5), min, max);
}

static void numeric_set_pixel(int cell, int x, int y) {
    if (cell < 0 || cell >= NUM_GRID) return;
    if (x < 0 || x >= NUMERIC_CANVAS_WIDTH) return;
    if (y < 0 || y >= NUMERIC_CANVAS_HEIGHT) return;

    int col = y;
    int row_from_bottom = x;
    vram[cell][col] |= (uint8_t)(1u << (6 - row_from_bottom));
}

static void numeric_draw_symbol_at(int cell, int symbol, int x_offset,
                                   int y_offset) {
    const uint8_t* glyph = font_3x5_dot;
    if (symbol != FONT_3X5_SYMBOL_DOT) {
        int digit = numeric_clamp_int(symbol, 0, 15);
        glyph = font_3x5_hex[digit];
    }

    for (int y = 0; y < FONT_3X5_HEIGHT; y++) {
        uint8_t row = glyph[y];
        for (int x = 0; x < FONT_3X5_WIDTH; x++) {
            if ((row & (1u << (FONT_3X5_WIDTH - 1 - x))) != 0) {
                numeric_set_pixel(cell, x_offset + x, y + y_offset);
            }
        }
    }
}

static void numeric_advance_roll_frames(numeric_roll_state_t* state) {
    for (int i = 0; i < 2; i++) {
        if (state->directions[i] == 0) continue;
        state->frames[i]++;
        if (state->frames[i] >= NUMERIC_ROLL_FRAMES) {
            state->digits[i] = state->next_digits[i];
            state->directions[i] = 0;
            state->frames[i] = 0;
        }
    }
}

static void numeric_start_base_roll(numeric_roll_state_t* state, int digit_idx,
                                    int direction, int base) {
    state->directions[digit_idx] = direction;
    state->next_digits[digit_idx] =
        (state->digits[digit_idx] + direction + base) % base;
    state->frames[digit_idx] = 1;
}

static int numeric_base_state_value(const numeric_roll_state_t* state,
                                    int base) {
    return state->digits[0] * base + state->digits[1];
}

static void numeric_update_base_roll(numeric_roll_state_t* state,
                                     const int target_symbols[2], int base) {
    if (!state->initialized) {
        state->initialized = true;
        state->digits[0] = target_symbols[0];
        state->digits[1] = target_symbols[1];
        return;
    }

    numeric_advance_roll_frames(state);
    int target = target_symbols[0] * base + target_symbols[1];
    int current = numeric_base_state_value(state, base);
    int direction = target > current ? 1 : target < current ? -1 : 0;
    if (direction == 0) return;

    for (int i = 0; i < 2; i++) {
        if (state->directions[i] == 0 &&
            state->digits[i] != target_symbols[i]) {
            numeric_start_base_roll(state, i, direction, base);
        }
    }
}

static void numeric_update_decimal_roll(numeric_roll_state_t* state,
                                        int target) {
    target = numeric_clamp_int(target, 0, 99);
    int symbols[2] = {target / 10, target % 10};
    numeric_update_base_roll(state, symbols, 10);
}

static int numeric_symbol_rank(const int symbols[2]) {
    if (symbols[0] == FONT_3X5_SYMBOL_DOT)
        return symbols[1];
    return (symbols[0] * 10 + symbols[1]) * 10;
}

static int numeric_next_symbol(int current, int direction, int target) {
    if (direction > 0) {
        if (current == FONT_3X5_SYMBOL_DOT) return 0;
        if (current >= 9) return 0;
        return current + 1;
    }

    if (current == 0)
        return target == FONT_3X5_SYMBOL_DOT ? FONT_3X5_SYMBOL_DOT : 9;
    if (current == FONT_3X5_SYMBOL_DOT)
        return FONT_3X5_SYMBOL_DOT;
    return current - 1;
}

static void numeric_start_symbol_roll(numeric_roll_state_t* state,
                                      int symbol_idx, int direction,
                                      int target) {
    state->directions[symbol_idx] = direction;
    state->next_digits[symbol_idx] =
        numeric_next_symbol(state->digits[symbol_idx], direction, target);
    state->frames[symbol_idx] = 1;
}

static void numeric_update_symbol_roll(numeric_roll_state_t* state,
                                       const int target_symbols[2],
                                       int direction_hint) {
    if (!state->initialized) {
        state->initialized = true;
        state->digits[0] = target_symbols[0];
        state->digits[1] = target_symbols[1];
        return;
    }

    numeric_advance_roll_frames(state);
    int direction = direction_hint;
    if (direction == 0) {
        int target_rank = numeric_symbol_rank(target_symbols);
        int current_rank = numeric_symbol_rank(state->digits);
        direction = target_rank > current_rank ? 1
                    : target_rank < current_rank ? -1
                                                 : 0;
    }
    if (direction == 0) return;

    for (int i = 0; i < 2; i++) {
        if (state->directions[i] == 0 &&
            state->digits[i] != target_symbols[i]) {
            numeric_start_symbol_roll(state, i, direction, target_symbols[i]);
        }
    }
}

static void numeric_render_digit_roll(int cell,
                                      const numeric_roll_state_t* state,
                                      int digit_idx) {
    int x_offset = digit_idx == 0 ? 0 : 4;
    int direction = state->directions[digit_idx];
    int frame = state->frames[digit_idx];

    if (direction == 0 || frame <= 0) {
        numeric_draw_symbol_at(cell, state->digits[digit_idx], x_offset, 0);
        return;
    }

    int current_y = direction > 0 ? -frame : frame;
    int next_y = direction > 0 ? FONT_3X5_HEIGHT - frame
                               : frame - FONT_3X5_HEIGHT;
    numeric_draw_symbol_at(cell, state->digits[digit_idx], x_offset,
                           current_y);
    numeric_draw_symbol_at(cell, state->next_digits[digit_idx], x_offset,
                           next_y);
}

static void numeric_render_roll(int cell, const numeric_roll_state_t* state) {
    memset(vram[cell], 0, GRID_SIZE);
    numeric_render_digit_roll(cell, state, 0);
    numeric_render_digit_roll(cell, state, 1);
}

static void numeric_render_uptime_remainder(int cell, int remainder_hours) {
    remainder_hours = numeric_clamp_int(remainder_hours, 0, 2);
    if (remainder_hours == 1) {
        numeric_set_pixel(cell, 3, 2);
    } else if (remainder_hours == 2) {
        numeric_set_pixel(cell, 3, 1);
        numeric_set_pixel(cell, 3, 3);
    }
}

static bool numeric_alert_blank_now(const struct timespec* now_ts,
                                    int failed_units) {
    if (failed_units <= 0) return false;

    int64_t now_ms = (int64_t)now_ts->tv_sec * 1000 +
                     now_ts->tv_nsec / 1000000;
    return (now_ms % NUMERIC_ALERT_BLINK_PERIOD_MS) <
           NUMERIC_ALERT_BLANK_MS;
}

static int numeric_stabilize_target(numeric_target_state_t* state, double raw) {
    if (raw < 0.0) raw = 0.0;
    if (raw > 99.0) raw = 99.0;

    if (!state->initialized) {
        state->initialized = true;
        state->target = numeric_round_clamped(raw, 0, 99);
        return state->target;
    }

    if (state->hold_frames > 0) {
        state->hold_frames--;
        return state->target;
    }

    double diff = raw - state->target;
    if (diff >= NUMERIC_TARGET_HYSTERESIS ||
        diff <= -NUMERIC_TARGET_HYSTERESIS) {
        int next = numeric_round_clamped(raw, 0, 99);
        if (next != state->target) {
            state->target = next;
            state->hold_frames = NUMERIC_TARGET_HOLD_FRAMES;
        }
    }

    return state->target;
}

static void numeric_net_symbols_from_value(bool fractional, int value,
                                           int symbols[2]) {
    if (fractional) {
        symbols[0] = FONT_3X5_SYMBOL_DOT;
        symbols[1] = numeric_clamp_int(value, 0, 9);
    } else {
        value = numeric_clamp_int(value, 1, 99);
        symbols[0] = value / 10;
        symbols[1] = value % 10;
    }
}

typedef struct {
    double value;
    int min_value;
    int max_value;
} numeric_net_value_t;

static numeric_net_value_t numeric_net_value_from_rate(double bytes_per_sec,
                                                       bool fractional) {
    numeric_net_value_t out;
    if (fractional) {
        out.value = bytes_per_sec / 100000.0;
        out.min_value = 0;
        out.max_value = 9;
    } else {
        out.value = bytes_per_sec / 1000000.0;
        out.min_value = 1;
        out.max_value = 99;
    }
    return out;
}

static int numeric_net_rank(bool fractional, int value) {
    return fractional ? value : value * 10;
}

static void numeric_set_net_candidate(numeric_net_state_t* state,
                                      bool fractional, int target) {
    if (state->candidate_active &&
        state->candidate_fractional == fractional &&
        state->candidate_target == target) {
        state->candidate_frames++;
        return;
    }

    state->candidate_active = true;
    state->candidate_fractional = fractional;
    state->candidate_target = target;
    state->candidate_frames = 1;
}

static void numeric_commit_net_candidate(numeric_net_state_t* state) {
    int target_rank = numeric_net_rank(state->fractional, state->target);
    int candidate_rank =
        numeric_net_rank(state->candidate_fractional, state->candidate_target);
    state->direction = candidate_rank > target_rank ? 1
                       : candidate_rank < target_rank ? -1
                                                      : 0;
    state->fractional = state->candidate_fractional;
    state->target = state->candidate_target;
    state->hold_frames = NUMERIC_NET_TARGET_HOLD_FRAMES;
    state->candidate_active = false;
}

static void numeric_stabilize_net(numeric_net_state_t* state,
                                  double bytes_per_sec, int symbols[2]) {
    if (bytes_per_sec < 0.0) bytes_per_sec = 0.0;

    if (!state->initialized) {
        state->initialized = true;
        state->filtered_rate = bytes_per_sec;
        state->fractional = bytes_per_sec < 1000000.0;

        numeric_net_value_t net_value =
            numeric_net_value_from_rate(state->filtered_rate, state->fractional);
        state->target = numeric_round_clamped(
            net_value.value, net_value.min_value, net_value.max_value);
        numeric_net_symbols_from_value(state->fractional, state->target,
                                       symbols);
        return;
    }

    state->filtered_rate +=
        (bytes_per_sec - state->filtered_rate) * NUMERIC_NET_EMA_ALPHA;

    bool next_fractional = state->fractional;
    if (state->fractional && state->filtered_rate >= NUMERIC_NET_MODE_UP_BPS)
        next_fractional = false;
    else if (!state->fractional &&
             state->filtered_rate < NUMERIC_NET_MODE_DOWN_BPS)
        next_fractional = true;

    numeric_net_value_t net_value =
        numeric_net_value_from_rate(state->filtered_rate, next_fractional);

    if (state->hold_frames > 0) {
        state->hold_frames--;
    } else {
        double diff = net_value.value - state->target;
        bool mode_changed = next_fractional != state->fractional;
        if (diff >= NUMERIC_NET_TARGET_HYSTERESIS ||
            diff <= -NUMERIC_NET_TARGET_HYSTERESIS || mode_changed) {
            int next = numeric_round_clamped(
                net_value.value, net_value.min_value, net_value.max_value);
            if (next != state->target || mode_changed) {
                numeric_set_net_candidate(state, next_fractional, next);
                if (state->candidate_frames >= NUMERIC_NET_COMMIT_FRAMES) {
                    numeric_commit_net_candidate(state);
                }
            }
        } else {
            state->candidate_active = false;
        }
    }

    numeric_net_symbols_from_value(state->fractional, state->target, symbols);
}

static bool numeric_show_connection_wait(void) {
    if (is_remote_connected()) return false;

    static int frame = 0;
    frame++;
    vfd_display_string(0, (frame % 10) < 8 ? "CON WAIT" : "        ");
    return true;
}

void numeric_map_monitor(void) {
    static bool first = true;
    static double prev_rx = 0.0;
    static double prev_tx = 0.0;
    static double rx_rate = 0.0;
    static double tx_rate = 0.0;
    static struct timespec prev_ts = {0};
    static numeric_roll_state_t states[NUM_GRID];
    static numeric_target_state_t target_states[NUM_GRID];
    static numeric_net_state_t net_states[2];

    if (numeric_show_connection_wait()) {
        first = true;
        return;
    }

    metrics_snapshot_t snap;
    get_remote_metrics_full(&snap);

    struct timespec now_ts;
    clock_gettime(CLOCK_MONOTONIC, &now_ts);

    if (first) {
        first = false;
        prev_rx = snap.net_rx_bytes;
        prev_tx = snap.net_tx_bytes;
        prev_ts = now_ts;
    } else if (snap.net_rx_bytes != prev_rx || snap.net_tx_bytes != prev_tx) {
        double dt = (now_ts.tv_sec - prev_ts.tv_sec) +
                    (now_ts.tv_nsec - prev_ts.tv_nsec) / 1e9;
        if (dt > 0.05) {
            rx_rate = snap.net_rx_bytes - prev_rx;
            tx_rate = snap.net_tx_bytes - prev_tx;
            rx_rate = rx_rate < 0.0 ? 0.0 : rx_rate / dt;
            tx_rate = tx_rate < 0.0 ? 0.0 : tx_rate / dt;
        }
        prev_rx = snap.net_rx_bytes;
        prev_tx = snap.net_tx_bytes;
        prev_ts = now_ts;
    }

    int uptime_hours = snap.uptime_sec / 3600;
    uptime_hours = numeric_clamp_int(uptime_hours, 0, 767);
    int uptime_bucket = uptime_hours / 3;
    int uptime_symbols[2] = {uptime_bucket / 16, uptime_bucket % 16};
    numeric_update_base_roll(&states[0], uptime_symbols, 16);
    numeric_render_roll(0, &states[0]);
    numeric_render_uptime_remainder(0, uptime_hours % 3);

    double raw_values[5] = {
        snap.cpu_util,
        snap.ram_util,
        snap.gpu_util,
        snap.cpu_temp,
        snap.gpu_temp,
    };

    for (int cell = 1; cell < 6; cell++) {
        int value = numeric_stabilize_target(&target_states[cell],
                                             raw_values[cell - 1]);
        numeric_update_decimal_roll(&states[cell], value);
        numeric_render_roll(cell, &states[cell]);
    }

    if (numeric_alert_blank_now(&now_ts, snap.failed_units)) {
        memset(vram[0], 0, GRID_SIZE);
    }

    int net_symbols[2];
    numeric_stabilize_net(&net_states[0], tx_rate, net_symbols);
    numeric_update_symbol_roll(&states[6], net_symbols,
                               net_states[0].direction);
    numeric_render_roll(6, &states[6]);

    numeric_stabilize_net(&net_states[1], rx_rate, net_symbols);
    numeric_update_symbol_roll(&states[7], net_symbols,
                               net_states[1].direction);
    numeric_render_roll(7, &states[7]);

    vfd_update_all_vram();
    vfd_display_all_vram();
}
