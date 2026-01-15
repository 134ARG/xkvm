#ifndef XKVM_NATIVE_MAIN_H
#define XKVM_NATIVE_MAIN_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <errno.h>
#include "ctrl.h"

void xkvm_c_log_handler(int level, const char *filename, const char *funcname, int line, const char *message);
void xkvm_video_handler(const uint8_t *frame, ssize_t len);
void xkvm_video_state_handler(xkvm_video_state_t *state);
void xkvm_indev_handler(int code);
void xkvm_rpc_handler(const char *method, const char *params);


// typedef void (xkvm_video_state_handler_t)(xkvm_video_state_t *state);
// typedef void (xkvm_log_handler_t)(int level, const char *filename, const char *funcname, int line, const char *message);
// typedef void (xkvm_rpc_handler_t)(const char *method, const char *params);
// typedef void (xkvm_video_handler_t)(const uint8_t *frame, ssize_t len);
// typedef void (xkvm_indev_handler_t)(int code);

#endif