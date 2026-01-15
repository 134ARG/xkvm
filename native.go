package kvm

import (
	"os"
	"sync"
	"time"

	"github.com/134ARG/xkvm/internal/diagnostics"
	"github.com/134ARG/xkvm/internal/native"
	"github.com/Masterminds/semver/v3"
	"github.com/pion/webrtc/v4/pkg/media"
)

var (
	nativeInstance native.NativeInterface
	nativeCmdLock  = sync.Mutex{}
)

func initNative(systemVersion *semver.Version, appVersion *semver.Version) {
	if failsafeModeActive {
		nativeInstance = &native.EmptyNativeInterface{}
		nativeLogger.Warn().Msg("failsafe mode active, using empty native interface")
		return
	}

	nativeLogger.Info().Msg("initializing native proxy")
	var err error
	nativeInstance, err = native.NewNativeProxy(native.NativeOptions{
		SystemVersion:        systemVersion,
		AppVersion:           appVersion,
		DefaultQualityFactor: config.VideoQualityFactor,
		MaxRestartAttempts:   config.NativeMaxRestart,
		OnNativeRestart: func() {
			// configureDisplayOnNativeRestart()
		},
		OnVideoStateChange: func(state native.VideoState) {
			lastVideoState = state
			triggerVideoStateUpdate()
		},
		OnIndevEvent: func(event string) {
			nativeLogger.Trace().Str("event", event).Msg("indev event received")
		},
		OnRpcEvent: func(event string) {
			nativeCmdLock.Lock()
			defer nativeCmdLock.Unlock()

			nativeLogger.Trace().Str("event", event).Msg("rpc event received")
			switch event {
			case "resetConfig":
				nativeLogger.Info().Msg("Reset configuration request via native rpc event")
				err := rpcResetConfig()
				if err != nil {
					nativeLogger.Warn().Err(err).Msg("error resetting config")
				}
			case "toggleDHCPClient":
				nativeLogger.Warn().Msg("DHCP client toggle disabled - network management is read-only")
				// Don't call rpcToggleDHCPClient() - it's now disabled
			default:
				nativeLogger.Warn().Str("event", event).Msg("unknown rpc event received")
			}
		},
		OnVideoFrameReceived: func(frame []byte, duration time.Duration) {
			if currentSession != nil {
				err := currentSession.VideoTrack.WriteSample(media.Sample{Data: frame, Duration: duration})
				if err != nil {
					nativeLogger.Warn().Err(err).Msg("error writing sample")
				}
			}
		},
		GetSessionInfo: func() diagnostics.SessionInfo {
			info := diagnostics.SessionInfo{
				ActiveSessions:    getActiveSessions(),
				HasCurrentSession: currentSession != nil,
			}
			if currentSession != nil {
				sessionInfo := currentSession.GetDiagnosticsInfo()
				info.ICEConnectionState = sessionInfo.ICEConnectionState
				info.SignalingState = sessionInfo.SignalingState
				info.ConnectionState = sessionInfo.ConnectionState
				info.DataChannels = sessionInfo.DataChannels
			}
			return info
		},
	})
	if err != nil {
		nativeLogger.Fatal().Err(err).Msg("failed to create native proxy")
	}

	if err := nativeInstance.Start(); err != nil {
		nativeLogger.Fatal().Err(err).Msg("failed to start native proxy")
	}
	go func() {
		if err := nativeInstance.VideoSetEDID(config.EdidString); err != nil {
			nativeLogger.Warn().Err(err).Msg("error setting EDID")
		}
	}()

	if os.Getenv("XKVM_CRASH_TESTING") == "1" {
		nativeInstance.DoNotUseThisIsForCrashTestingOnly()
	}
}
