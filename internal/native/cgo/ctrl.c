#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/un.h>
#include <sys/socket.h>
#include <errno.h>
#include <unistd.h>
#include <pthread.h>
#include <stdint.h>
#include <fcntl.h>
#include "video.h"
#include "edid.h"
#include "ctrl.h"
#include "log.h"
#include "log_handler.h"

xkvm_video_state_t state;
xkvm_video_state_handler_t *video_state_handler = NULL;
xkvm_rpc_handler_t *rpc_handler = NULL;
xkvm_video_handler_t *video_handler = NULL;


void xkvm_set_log_handler(xkvm_log_handler_t *handler) {
    log_set_handler(handler);
}

void xkvm_set_video_handler(xkvm_video_handler_t *handler) {
    video_handler = handler;
}

static xkvm_indev_handler_t *xkvm_indev_handler = NULL;

void xkvm_set_indev_handler(xkvm_indev_handler_t *handler) {
    xkvm_indev_handler = handler;
    // Note: No LVGL integration for headless operation
}

void xkvm_set_rpc_handler(xkvm_rpc_handler_t *handler) {
    rpc_handler = handler;
}

void xkvm_call_rpc_handler(const char *method, const char *params) {
    if (rpc_handler != NULL) {
        (*rpc_handler)(method, params);
    }
}

const char *xkvm_ui_event_code_to_name(int code) {
    // Headless operation - no LVGL event codes
    return "UNKNOWN";
}

void video_report_format(bool ready, const char *error, u_int16_t width, u_int16_t height, double frame_per_second)
{
    state.streaming = video_get_streaming_status();
    state.ready = ready;
    state.error = error;
    state.width = width;
    state.height = height;
    state.frame_per_second = frame_per_second;
    if (video_state_handler != NULL) {
        (*video_state_handler)(&state);
    }
}

void video_send_format_report() {
    state.streaming = video_get_streaming_status();
    if (video_state_handler != NULL) {
        (*video_state_handler)(&state);
    }
}

int video_send_frame(const uint8_t *frame, ssize_t len)
{
    if (video_handler != NULL) {
        (*video_handler)(frame, len);
    } else {
        log_error("video handler is not set");
    }
    return 0;
}

/**
 * @brief Convert a hexadecimal string to an array of uint8_t bytes
 *
 * @param hex_str The input hexadecimal string
 * @param bytes The output byte array (must be pre-allocated)
 * @param max_len The maximum number of bytes that can be stored in the output array
 * @return int The number of bytes converted, or -1 on error
 */
int hex_to_bytes(const char *hex_str, uint8_t *bytes, size_t max_len)
{
    size_t hex_len = strnlen(hex_str, 4096);
    if (hex_len % 2 != 0 || hex_len / 2 > max_len)
    {
        return -1; // Invalid input length or insufficient output buffer
    }

    for (size_t i = 0; i < hex_len; i += 2)
    {
        char byte_str[3] = {hex_str[i], hex_str[i + 1], '\0'};
        char *end_ptr;
        long value = strtol(byte_str, &end_ptr, 16);

        if (*end_ptr != '\0' || value < 0 || value > 255)
        {
            return -1; // Invalid hexadecimal value
        }

        bytes[i / 2] = (uint8_t)value;
    }

    return hex_len / 2;
}

/**
 * @brief Convert an array of uint8_t bytes to a hexadecimal string, user must free the returned string
 *
 * @param bytes The input byte array
 * @param len The number of bytes in the input array
 * @return char* The output hexadecimal string (dynamically allocated, must be freed by the caller), or NULL on error
 */
const char *bytes_to_hex(const uint8_t *bytes, size_t len)
{
    if (bytes == NULL || len == 0)
    {
        return NULL;
    }

    char *hex_str = malloc(2 * len + 1); // Each byte becomes 2 hex chars, plus null terminator
    if (hex_str == NULL)
    {
        return NULL; // Memory allocation failed
    }

    for (size_t i = 0; i < len; i++)
    {
        snprintf(hex_str + (2 * i), 3, "%02x", bytes[i]);
    }

    hex_str[2 * len] = '\0'; // Ensure null termination
    return hex_str;
}

// LVGL functions removed for headless operation
void xkvm_ui_set_var(const char *name, const char *value) {
    // No-op for headless operation
}

const char *xkvm_ui_get_var(const char *name) {
    // No-op for headless operation
    return NULL;
}

void xkvm_ui_init(u_int16_t rotation) {
    // No-op for headless operation
}

void xkvm_ui_tick() {
    // No-op for headless operation
}

void xkvm_set_video_state_handler(xkvm_video_state_handler_t *handler) {
    video_state_handler = handler;
}

void xkvm_ui_set_rotation(u_int16_t rotation) {
    // No-op for headless operation
}

const char *xkvm_ui_get_current_screen() {
    // No-op for headless operation
    return NULL;
}

void xkvm_ui_load_screen(const char *obj_name) {
    // No-op for headless operation
}

int xkvm_ui_set_text(const char *obj_name, const char *text) {
    // No-op for headless operation
    return -1;
}

void xkvm_ui_set_image(const char *obj_name, const char *image_name) {
    // No-op for headless operation
}

void xkvm_ui_add_state(const char *obj_name, const char *state_name) {
    // No-op for headless operation
}

void xkvm_ui_clear_state(const char *obj_name, const char *state_name) {
    // No-op for headless operation
}

int xkvm_ui_add_flag(const char *obj_name, const char *flag_name) {
    // No-op for headless operation
    return -1;
}

int xkvm_ui_clear_flag(const char *obj_name, const char *flag_name) {
    // No-op for headless operation
    return -1;
}

void xkvm_ui_fade_in(const char *obj_name, u_int32_t duration) {
    // No-op for headless operation
}

void xkvm_ui_fade_out(const char *obj_name, u_int32_t duration) {
    // No-op for headless operation
}

void xkvm_ui_set_opacity(const char *obj_name, u_int8_t opacity) {
    // No-op for headless operation
}

const char *xkvm_ui_get_lvgl_version() {
    // No LVGL for headless operation
    return "N/A (headless)";
}

void xkvm_video_start() {
    video_start_streaming();
}

void xkvm_video_stop() {
    video_stop_streaming();
}

uint8_t xkvm_video_get_streaming_status() {
    return video_get_streaming_status();
}

int xkvm_video_set_quality_factor(float quality_factor) {
    // Validate bitrate range (1000-20000 kbps)
    if (quality_factor < 1000 || quality_factor > 20000) {
        fprintf(stderr, "[NATIVE] xkvm_video_set_quality_factor: Invalid bitrate %.0f, must be between 1000-20000 kbps\n", quality_factor);
        return -1;
    }
    fprintf(stderr, "[NATIVE] xkvm_video_set_quality_factor: Calling video_set_quality_factor with %.0f kbps\n", quality_factor);
    video_set_quality_factor(quality_factor);
    return 0;
}

float xkvm_video_get_quality_factor() {
    return video_get_quality_factor();
}

void xkvm_video_set_encoder(int32_t encoder) {
    fprintf(stderr, "[NATIVE] xkvm_video_set_encoder: Calling video_set_encoder with %d\n", encoder);
    video_set_encoder(encoder);
}

int32_t xkvm_video_get_encoder() {
    return video_get_encoder();
}

int xkvm_video_set_edid(const char *edid_hex) {
    uint8_t edid[256];
    int edid_len = hex_to_bytes(edid_hex, edid, 256);
    if (edid_len < 0) {
        return -1;
    }
    return set_edid(edid, edid_len);
}

char *xkvm_video_get_edid_hex() {
    uint8_t edid[256];
    int edid_len = get_edid(edid, 256);
    if (edid_len < 0) {
        return NULL;
    }
    return (char *)bytes_to_hex(edid, edid_len);
}

xkvm_video_state_t *xkvm_video_get_status() {
    return &state;
}

char *xkvm_video_log_status() {
    return (char *)videoc_log_status();
}

int xkvm_video_init(float factor) {
    return video_init(factor);
}

void xkvm_video_shutdown() {
    video_shutdown();
}

void xkvm_crash() {
    // let's call a function that will crash the program
    int* p = 0;
    *p = 0;
}
