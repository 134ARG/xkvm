#include "themes.h"

#include <stdbool.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

#include "remote.h"
#include "vfd.h"

#define DENSE_DOTS_PER_CELL 35
#define DENSE_TOP_ROW 6
#define DENSE_WEAR_SWAP_SEC (4 * 60 * 60)

static double dense_clamp_double(double value, double min, double max) {
    if (value < min) return min;
    if (value > max) return max;
    return value;
}

static bool dense_flip_top_bottom(void) {
    struct timespec now;
    clock_gettime(CLOCK_MONOTONIC, &now);
    return ((now.tv_sec / DENSE_WEAR_SWAP_SEC) % 2) != 0;
}

static void dense_clear_cell(int cell, bool flipped) {
    (void)flipped;
    memset(vram[cell], 0, GRID_SIZE);
}

static void dense_set_dot_xy(int cell, int col, int row_from_bottom,
                             bool flipped) {
    if (flipped) row_from_bottom = DENSE_TOP_ROW - row_from_bottom;
    vram[cell][col] |= (uint8_t)(1u << (6 - row_from_bottom));
}

static void dense_set_dot_index(int cell, int dot, bool flipped) {
    int row = dot / GRID_SIZE;
    int offset = dot % GRID_SIZE;
    int col = (GRID_SIZE - 1) - offset;
    dense_set_dot_xy(cell, col, row, flipped);
}

static void dense_render_fill(int cell, float dots, bool flipped) {
    dense_clear_cell(cell, flipped);
    if (dots < 0.0f) dots = 0.0f;
    if (dots > DENSE_DOTS_PER_CELL) dots = DENSE_DOTS_PER_CELL;

    int whole_dots = (int)(dots + 0.5f);
    for (int dot = 0; dot < whole_dots; dot++) {
        dense_set_dot_index(cell, dot, flipped);
    }
}

static float dense_percent_to_dots(double pct) {
    pct = dense_clamp_double(pct, 0.0, 100.0);
    return (float)(pct * DENSE_DOTS_PER_CELL / 100.0);
}

static float dense_temp_to_dots(double temp_c) {
    temp_c = dense_clamp_double(temp_c, 0.0, 95.0);
    return (float)(temp_c * DENSE_DOTS_PER_CELL / 95.0);
}

static float dense_interp_dots(double value, double min_value, double max_value,
                               int min_dots, int max_dots) {
    double frac = (value - min_value) / (max_value - min_value);
    double dots = min_dots + frac * (max_dots - min_dots);
    dots = dense_clamp_double(dots, 0.0, DENSE_DOTS_PER_CELL);
    return (float)dots;
}

static float dense_smooth_dots(float* displayed, float target) {
    float diff = target - *displayed;
    if (diff > 1.0f || diff < -1.0f)
        *displayed += diff * 0.3f;
    else
        *displayed = target;

    if (*displayed < 0.0f) *displayed = 0.0f;
    if (*displayed > DENSE_DOTS_PER_CELL)
        *displayed = DENSE_DOTS_PER_CELL;

    return *displayed;
}

static float dense_net_to_dots(double bytes_per_sec) {
    if (bytes_per_sec <= 0.0) return 0;
    if (bytes_per_sec < 1000.0)
        return dense_interp_dots(bytes_per_sec, 0.0, 1000.0, 0, 5);
    if (bytes_per_sec < 10000.0)
        return dense_interp_dots(bytes_per_sec, 1000.0, 10000.0, 5, 10);
    if (bytes_per_sec < 100000.0)
        return dense_interp_dots(bytes_per_sec, 10000.0, 100000.0, 10, 15);
    if (bytes_per_sec < 1000000.0)
        return dense_interp_dots(bytes_per_sec, 100000.0, 1000000.0, 15, 20);
    if (bytes_per_sec < 10000000.0)
        return dense_interp_dots(bytes_per_sec, 1000000.0, 10000000.0, 20, 25);
    if (bytes_per_sec < 50000000.0)
        return dense_interp_dots(bytes_per_sec, 10000000.0, 50000000.0, 25, 30);
    if (bytes_per_sec < 100000000.0)
        return dense_interp_dots(bytes_per_sec, 50000000.0, 100000000.0, 30, 35);
    return 35;
}

static void dense_render_uptime(int cell, int uptime_sec, bool flipped) {
    dense_clear_cell(cell, flipped);

    int hours = uptime_sec / 3600;
    if (hours < 0) hours = 0;
    bool overflow = hours > 959;
    if (overflow) hours = 959;

    int blocks_30h = hours / 30;
    int hour_in_block = hours % 30;

    for (int bit = 0; bit < GRID_SIZE; bit++) {
        if (overflow || ((blocks_30h >> bit) & 1) != 0) {
            dense_set_dot_xy(cell, (GRID_SIZE - 1) - bit, DENSE_TOP_ROW,
                             flipped);
        }
    }

    for (int dot = 0; dot < hour_in_block; dot++) {
        dense_set_dot_index(cell, dot, flipped);
    }
}

static bool dense_show_connection_wait(void) {
    if (is_remote_connected()) return false;

    static int frame = 0;
    frame++;
    vfd_display_string(0, (frame % 10) < 8 ? "CON WAIT" : "        ");
    return true;
}

void dense_map_monitor(void) {
    static bool first = true;
    static double prev_rx = 0.0;
    static double prev_tx = 0.0;
    static double rx_rate = 0.0;
    static double tx_rate = 0.0;
    static struct timespec prev_ts = {0};
    static float displayed_cpu = 0.0f;
    static float displayed_cpu_temp = 0.0f;
    static float displayed_ram = 0.0f;
    static float displayed_gpu = 0.0f;
    static float displayed_gpu_temp = 0.0f;
    static float displayed_rx = 0.0f;
    static float displayed_tx = 0.0f;

    if (dense_show_connection_wait()) {
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
            rx_rate = (snap.net_rx_bytes - prev_rx) / dt;
            tx_rate = (snap.net_tx_bytes - prev_tx) / dt;
            if (rx_rate < 0.0) rx_rate = 0.0;
            if (tx_rate < 0.0) tx_rate = 0.0;
        }
        prev_rx = snap.net_rx_bytes;
        prev_tx = snap.net_tx_bytes;
        prev_ts = now_ts;
    }

    bool flipped = dense_flip_top_bottom();
    dense_render_uptime(0, snap.uptime_sec, flipped);
    dense_render_fill(1, dense_smooth_dots(&displayed_cpu,
                                            dense_percent_to_dots(snap.cpu_util)),
                      flipped);
    dense_render_fill(2, dense_smooth_dots(&displayed_ram,
                                            dense_percent_to_dots(snap.ram_util)),
                      flipped);
    dense_render_fill(3, dense_smooth_dots(&displayed_gpu,
                                            dense_percent_to_dots(snap.gpu_util)),
                      flipped);
    dense_render_fill(4, dense_smooth_dots(&displayed_cpu_temp,
                                            dense_temp_to_dots(snap.cpu_temp)),
                      flipped);
    dense_render_fill(5, dense_smooth_dots(&displayed_gpu_temp,
                                            dense_temp_to_dots(snap.gpu_temp)),
                      flipped);
    dense_render_fill(6, dense_smooth_dots(&displayed_tx,
                                            dense_net_to_dots(tx_rate)),
                      flipped);
    dense_render_fill(7, dense_smooth_dots(&displayed_rx,
                                            dense_net_to_dots(rx_rate)),
                      flipped);

    vfd_update_all_vram();
    vfd_display_all_vram();
}
