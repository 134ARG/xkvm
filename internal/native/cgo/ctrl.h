#ifndef VIDEO_DAEMON_CTRL_H
#define VIDEO_DAEMON_CTRL_H

#include <stdbool.h>
#include <stdint.h>
#include <sys/types.h>

typedef struct
{
    bool ready;
    uint8_t streaming;
    const char *error;
    u_int16_t width;
    u_int16_t height;
    double frame_per_second;
} xkvm_video_state_t;

typedef void (xkvm_video_state_handler_t)(xkvm_video_state_t *state);
typedef void (xkvm_log_handler_t)(int level, const char *filename, const char *funcname, int line, const char *message);
typedef void (xkvm_rpc_handler_t)(const char *method, const char *params);
typedef void (xkvm_video_handler_t)(const uint8_t *frame, ssize_t len);
typedef void (xkvm_indev_handler_t)(int code);

void xkvm_set_log_handler(xkvm_log_handler_t *handler);
void xkvm_set_video_handler(xkvm_video_handler_t *handler);
void xkvm_set_indev_handler(xkvm_indev_handler_t *handler);
void xkvm_set_rpc_handler(xkvm_rpc_handler_t *handler);
void xkvm_call_rpc_handler(const char *method, const char *params);
void xkvm_set_video_state_handler(xkvm_video_state_handler_t *handler);
void xkvm_crash();

void xkvm_ui_set_var(const char *name, const char *value);
const char *xkvm_ui_get_var(const char *name);

void xkvm_ui_init(u_int16_t rotation);
void xkvm_ui_tick();


void xkvm_ui_set_rotation(u_int16_t rotation);
const char *xkvm_ui_get_current_screen();
void xkvm_ui_load_screen(const char *obj_name);
int xkvm_ui_set_text(const char *obj_name, const char *text);
void xkvm_ui_set_image(const char *obj_name, const char *image_name);
void xkvm_ui_add_state(const char *obj_name, const char *state_name);
void xkvm_ui_clear_state(const char *obj_name, const char *state_name);
void xkvm_ui_fade_in(const char *obj_name, u_int32_t duration);
void xkvm_ui_fade_out(const char *obj_name, u_int32_t duration);
void xkvm_ui_set_opacity(const char *obj_name, u_int8_t opacity);
int xkvm_ui_add_flag(const char *obj_name, const char *flag_name);
int xkvm_ui_clear_flag(const char *obj_name, const char *flag_name);

const char *xkvm_ui_get_lvgl_version();

const char *xkvm_ui_event_code_to_name(int code);

int xkvm_video_init(float quality_factor);
void xkvm_video_shutdown();
void xkvm_video_start();
void xkvm_video_stop();
uint8_t xkvm_video_get_streaming_status();
int xkvm_video_set_quality_factor(float quality_factor);
float xkvm_video_get_quality_factor();
void xkvm_video_set_encoder(int32_t encoder);
int32_t xkvm_video_get_encoder();
int xkvm_video_set_edid(const char *edid_hex);
char *xkvm_video_get_edid_hex();
char *xkvm_video_log_status();
xkvm_video_state_t *xkvm_video_get_status();

void video_report_format(bool ready, const char *error, u_int16_t width, u_int16_t height, double frame_per_second);
void video_send_format_report();
int video_send_frame(const uint8_t *frame, ssize_t len);



#endif //VIDEO_DAEMON_CTRL_H
