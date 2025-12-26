//go:build linux

// TODO: use a generator to generate the cgo code for the native functions
// there's too much boilerplate code to write manually

package native

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/rs/zerolog"
)

/*
#cgo LDFLAGS: -Lcgo/lib -ljknative
#cgo CFLAGS: -Icgo/include
#cgo linux,arm64 LDFLAGS: -Lcgo/lib -ljknative -lrockit -lrockchip_mpp
#cgo linux,amd64 LDFLAGS: -Lcgo/lib -ljknative
#include "ctrl.h"
#include <stdlib.h>

typedef const char cchar_t;
typedef const uint8_t cuint8_t;

extern void jetkvm_go_log_handler(int level, cchar_t *filename, cchar_t *funcname, int line, cchar_t *message);
static inline void jetkvm_cgo_setup_log_handler() {
    jetkvm_set_log_handler(&jetkvm_go_log_handler);
}

extern void jetkvm_go_video_state_handler(jetkvm_video_state_t *state);
static inline void jetkvm_cgo_setup_video_state_handler() {
    jetkvm_set_video_state_handler(&jetkvm_go_video_state_handler);
}

extern void jetkvm_go_video_handler(cuint8_t *frame, ssize_t len);
static inline void jetkvm_cgo_setup_video_handler() {
    jetkvm_set_video_handler(&jetkvm_go_video_handler);
}

extern void jetkvm_go_indev_handler(int code);
static inline void jetkvm_cgo_setup_indev_handler() {
    jetkvm_set_indev_handler(&jetkvm_go_indev_handler);
}

extern void jetkvm_go_rpc_handler(cchar_t *method, cchar_t *params);
static inline void jetkvm_cgo_setup_rpc_handler() {
    jetkvm_set_rpc_handler(&jetkvm_go_rpc_handler);
}
*/
import "C"

var (
	cgoLock sync.Mutex
)

//export jetkvm_go_video_state_handler
func jetkvm_go_video_state_handler(state *C.jetkvm_video_state_t) {
	videoState := VideoState{
		Ready:          bool(state.ready),
		Streaming:      VideoStreamingStatus(state.streaming),
		Error:          C.GoString(state.error),
		Width:          int(state.width),
		Height:         int(state.height),
		FramePerSecond: float64(state.frame_per_second),
	}
	videoStateChan <- videoState
}

//export jetkvm_go_log_handler
func jetkvm_go_log_handler(level C.int, filename *C.cchar_t, funcname *C.cchar_t, line C.int, message *C.cchar_t) {
	logMessage := nativeLogMessage{
		Level:    zerolog.Level(level),
		Message:  C.GoString(message),
		File:     C.GoString(filename),
		FuncName: C.GoString(funcname),
		Line:     int(line),
	}

	logChan <- logMessage
}

//export jetkvm_go_video_handler
func jetkvm_go_video_handler(frame *C.cuint8_t, len C.ssize_t) {
	videoFrameChan <- C.GoBytes(unsafe.Pointer(frame), C.int(len))
}

//export jetkvm_go_indev_handler
func jetkvm_go_indev_handler(code C.int) {
	indevEventChan <- int(code)
}

//export jetkvm_go_rpc_handler
func jetkvm_go_rpc_handler(method *C.cchar_t, params *C.cchar_t) {
	rpcEventChan <- C.GoString(method)
}

var eventCodeToNameMap = map[int]string{}

func uiEventCodeToName(code int) string {
	// Headless operation - no LVGL event codes
	return "UNKNOWN"
}

func setUpNativeHandlers() {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	C.jetkvm_cgo_setup_log_handler()
	C.jetkvm_cgo_setup_video_state_handler()
	C.jetkvm_cgo_setup_video_handler()
	C.jetkvm_cgo_setup_indev_handler()
	C.jetkvm_cgo_setup_rpc_handler()
}

func uiInit(rotation uint16) {
	// No-op for headless operation
	nativeLogger.Info().Msg("UI initialization skipped - headless operation")
}

func uiTick() {
	// No-op for headless operation
}

func videoInit(factor float64) error {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	factorC := C.float(factor)

	ret := C.jetkvm_video_init(factorC)
	if ret != 0 {
		return fmt.Errorf("failed to initialize video: %d", ret)
	}
	return nil
}

func videoShutdown() {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	C.jetkvm_video_shutdown()
}

func videoStart() {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	C.jetkvm_video_start()
}

func videoStop() {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	C.jetkvm_video_stop()
}

func videoGetStreamingStatus() VideoStreamingStatus {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	isStreaming := C.jetkvm_video_get_streaming_status()

	return VideoStreamingStatus(isStreaming)
}

func videoLogStatus() string {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	logStatus := C.jetkvm_video_log_status()
	defer C.free(unsafe.Pointer(logStatus))

	return C.GoString(logStatus)
}

func uiSetVar(name string, value string) {
	// No-op for headless operation
}

func uiGetVar(name string) string {
	// No-op for headless operation
	return ""
}

func uiSwitchToScreen(screen string) {
	// No-op for headless operation
}

func uiGetCurrentScreen() string {
	// No-op for headless operation
	return ""
}

func uiObjAddState(objName string, state string) (bool, error) {
	// No-op for headless operation
	return true, nil
}

func uiObjClearState(objName string, state string) (bool, error) {
	// No-op for headless operation
	return true, nil
}

func uiGetLVGLVersion() string {
	// No LVGL for headless operation
	return "N/A (headless)"
}

// TODO: use Enum instead of string but it's not a hot path and performance is not a concern now
func uiObjAddFlag(objName string, flag string) (bool, error) {
	// No-op for headless operation
	return true, nil
}

func uiObjClearFlag(objName string, flag string) (bool, error) {
	// No-op for headless operation
	return true, nil
}

func uiObjHide(objName string) (bool, error) {
	return uiObjAddFlag(objName, "LV_OBJ_FLAG_HIDDEN")
}

func uiObjShow(objName string) (bool, error) {
	return uiObjClearFlag(objName, "LV_OBJ_FLAG_HIDDEN")
}

func uiObjSetOpacity(objName string, opacity int) (bool, error) {
	// No-op for headless operation
	return true, nil
}

func uiObjFadeIn(objName string, duration uint32) (bool, error) {
	// No-op for headless operation
	return true, nil
}

func uiObjFadeOut(objName string, duration uint32) (bool, error) {
	// No-op for headless operation
	return true, nil
}

func uiLabelSetText(objName string, text string) (bool, error) {
	// No-op for headless operation
	return true, nil
}

func uiImgSetSrc(objName string, src string) (bool, error) {
	// No-op for headless operation
	return true, nil
}

func uiDispSetRotation(rotation uint16) (bool, error) {
	// No-op for headless operation
	nativeLogger.Info().Uint16("rotation", rotation).Msg("rotation setting skipped - headless operation")
	return true, nil
}

func videoGetStreamQualityFactor() (float64, error) {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	factor := C.jetkvm_video_get_quality_factor()
	return float64(factor), nil
}

func videoSetStreamQualityFactor(factor float64) error {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	C.jetkvm_video_set_quality_factor(C.float(factor))
	return nil
}

func videoGetEDID() (string, error) {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	edidCStr := C.jetkvm_video_get_edid_hex()
	return C.GoString(edidCStr), nil
}

func videoSetEDID(edid string) error {
	cgoLock.Lock()
	defer cgoLock.Unlock()

	edidCStr := C.CString(edid)
	defer C.free(unsafe.Pointer(edidCStr))
	C.jetkvm_video_set_edid(edidCStr)
	return nil
}

// DO NOT USE THIS FUNCTION IN PRODUCTION
// This is only for testing purposes
func crash() {
	C.jetkvm_crash()
}
