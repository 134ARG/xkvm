package kvm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/134ARG/xkvm/internal/diagnostics"
	"github.com/134ARG/xkvm/internal/hidrpc"
	"github.com/134ARG/xkvm/internal/logging"
	"github.com/134ARG/xkvm/internal/usbgadget"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"
)

type Session struct {
	peerConnection           *webrtc.PeerConnection
	VideoTrack               *webrtc.TrackLocalStaticSample
	ControlChannel           *webrtc.DataChannel
	RPCChannel               *webrtc.DataChannel
	HidChannel               *webrtc.DataChannel
	shouldUmountVirtualMedia bool

	rpcQueue chan webrtc.DataChannelMessage

	hidRPCAvailable          bool
	lastKeepAliveArrivalTime time.Time  // Track when last keep-alive packet arrived
	lastTimerResetTime       time.Time  // Track when auto-release timer was last reset
	keepAliveJitterLock      sync.Mutex // Protect jitter compensation timing state
	hidQueueLock             sync.Mutex
	hidQueue                 []chan hidQueueMessage

	keysDownStateQueue chan usbgadget.KeysDownState

	// done is closed exactly once (guarded by closeOnce) when the session is
	// torn down. Queue consumers and senders select on it so we never close the
	// queue channels themselves — this eliminates send-on-closed and
	// double-close panics during session replacement/teardown.
	done      chan struct{}
	closeOnce sync.Once
}

// teardown signals all queue consumers/senders to stop. It is idempotent and
// safe to call from multiple goroutines (e.g. repeated ICE state callbacks).
func (s *Session) teardown() {
	s.closeOnce.Do(func() {
		close(s.done)
	})
}

var (
	actionSessions      int = 0
	activeSessionsMutex     = &sync.Mutex{}

	activeSessionVideoCodec   int32 = -1
	activeSessionVideoCodecMu sync.Mutex
)

func incrActiveSessions() int {
	activeSessionsMutex.Lock()
	defer activeSessionsMutex.Unlock()

	actionSessions++
	return actionSessions
}

func decrActiveSessions() int {
	activeSessionsMutex.Lock()
	defer activeSessionsMutex.Unlock()

	actionSessions--
	return actionSessions
}

func getActiveSessions() int {
	activeSessionsMutex.Lock()
	defer activeSessionsMutex.Unlock()

	return actionSessions
}

// GetDiagnosticsInfo returns WebRTC diagnostic info for the diagnostics package.
func (s *Session) GetDiagnosticsInfo() diagnostics.SessionInfo {
	info := diagnostics.SessionInfo{
		HasCurrentSession: true,
	}

	if s.peerConnection != nil {
		pc := s.peerConnection
		info.ICEConnectionState = pc.ICEConnectionState().String()
		info.SignalingState = pc.SignalingState().String()
		info.ConnectionState = pc.ConnectionState().String()

		var channels []diagnostics.DataChannelInfo
		if s.ControlChannel != nil {
			channels = append(channels, diagnostics.DataChannelInfo{
				Label: s.ControlChannel.Label(),
				State: s.ControlChannel.ReadyState().String(),
			})
		}
		if s.RPCChannel != nil {
			channels = append(channels, diagnostics.DataChannelInfo{
				Label: s.RPCChannel.Label(),
				State: s.RPCChannel.ReadyState().String(),
			})
		}
		if s.HidChannel != nil {
			channels = append(channels, diagnostics.DataChannelInfo{
				Label: s.HidChannel.Label(),
				State: s.HidChannel.ReadyState().String(),
			})
		}
		info.DataChannels = channels
	}

	return info
}

func (s *Session) resetKeepAliveTime() {
	s.keepAliveJitterLock.Lock()
	defer s.keepAliveJitterLock.Unlock()
	s.lastKeepAliveArrivalTime = time.Time{} // Reset keep-alive timing tracking
	s.lastTimerResetTime = time.Time{}       // Reset auto-release timer tracking
}

type hidQueueMessage struct {
	webrtc.DataChannelMessage
	channel string
}

type SessionConfig struct {
	ICEServers          []string
	LocalIP             string
	IsCloud             bool
	ws                  *websocket.Conn
	Logger              *zerolog.Logger
	PreferredVideoCodec *int32
}

func (s *Session) ExchangeOffer(offerStr string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(offerStr)
	if err != nil {
		return "", err
	}
	offer := webrtc.SessionDescription{}
	err = json.Unmarshal(b, &offer)
	if err != nil {
		return "", err
	}
	// Set the remote SessionDescription
	if err = s.peerConnection.SetRemoteDescription(offer); err != nil {
		return "", err
	}

	// Create answer
	answer, err := s.peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err = s.peerConnection.SetLocalDescription(answer); err != nil {
		return "", err
	}

	localDescription, err := json.Marshal(s.peerConnection.LocalDescription())
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(localDescription), nil
}

func (s *Session) initQueues() {
	s.hidQueueLock.Lock()
	defer s.hidQueueLock.Unlock()

	s.hidQueue = make([]chan hidQueueMessage, 0)
	for i := 0; i < 4; i++ {
		q := make(chan hidQueueMessage, 256)
		s.hidQueue = append(s.hidQueue, q)
	}
}

func (s *Session) handleQueues(index int) {
	for {
		select {
		case <-s.done:
			return
		case msg := <-s.hidQueue[index]:
			onHidMessage(msg, s)
		}
	}
}

const keysDownStateQueueSize = 64

func (s *Session) initKeysDownStateQueue() {
	// serialise outbound key state reports so unreliable links can't stall input handling
	s.keysDownStateQueue = make(chan usbgadget.KeysDownState, keysDownStateQueueSize)
	go s.handleKeysDownStateQueue()
}

func (s *Session) handleKeysDownStateQueue() {
	for {
		select {
		case <-s.done:
			return
		case state := <-s.keysDownStateQueue:
			s.reportHidRPCKeysDownState(state)
		}
	}
}

func (s *Session) enqueueKeysDownState(state usbgadget.KeysDownState) {
	if s == nil || s.keysDownStateQueue == nil {
		return
	}

	select {
	case s.keysDownStateQueue <- state:
	case <-s.done:
	default:
		hidRPCLogger.Warn().Msg("dropping keys down state update; queue full")
	}
}

func getOnHidMessageHandler(session *Session, scopedLogger *zerolog.Logger, channel string) func(msg webrtc.DataChannelMessage) {
	return func(msg webrtc.DataChannelMessage) {
		l := scopedLogger.With().
			Str("channel", channel).
			Int("length", len(msg.Data)).
			Logger()
		// only log data if the log level is debug or lower
		if scopedLogger.GetLevel() > zerolog.DebugLevel {
			l = l.With().Str("data", string(msg.Data)).Logger()
		}

		if msg.IsString {
			l.Warn().Msg("received string data in HID RPC message handler")
			return
		}

		if len(msg.Data) < 1 {
			l.Warn().Msg("received empty data in HID RPC message handler")
			return
		}

		l.Trace().Msg("received data in HID RPC message handler")

		// Enqueue to ensure ordered processing
		queueIndex := hidrpc.GetQueueIndex(hidrpc.MessageType(msg.Data[0]))
		if queueIndex >= len(session.hidQueue) || queueIndex < 0 {
			l.Warn().Int("queueIndex", queueIndex).Msg("received data in HID RPC message handler, but queue index not found")
			queueIndex = 3
		}

		queue := session.hidQueue[queueIndex]
		if queue != nil {
			// Select on done so a late message during teardown can't block or
			// panic (the queue channels are never closed).
			select {
			case queue <- hidQueueMessage{
				DataChannelMessage: msg,
				channel:            channel,
			}:
			case <-session.done:
			}
		} else {
			l.Warn().Int("queueIndex", queueIndex).Msg("received data in HID RPC message handler, but queue is nil")
			return
		}
	}
}

func newSession(config SessionConfig) (*Session, error) {
	// Get codec from global config before it's shadowed
	globalConfig := getGlobalConfig()
	videoCodec := globalConfig.VideoCodec
	if config.PreferredVideoCodec != nil {
		if *config.PreferredVideoCodec != 0 && *config.PreferredVideoCodec != 1 {
			return nil, fmt.Errorf("invalid preferred video codec: %d", *config.PreferredVideoCodec)
		}
		videoCodec = *config.PreferredVideoCodec
	}

	webrtcSettingEngine := webrtc.SettingEngine{
		LoggerFactory: logging.GetPionDefaultLoggerFactory(),
	}
	iceServer := webrtc.ICEServer{}

	var scopedLogger *zerolog.Logger
	if config.Logger != nil {
		l := config.Logger.With().Str("component", "webrtc").Logger()
		scopedLogger = &l
	} else {
		scopedLogger = webrtcLogger
	}

	if config.IsCloud {
		if config.ICEServers == nil {
			scopedLogger.Info().Msg("ICE Servers not provided by cloud")
		} else {
			iceServer.URLs = config.ICEServers
			scopedLogger.Info().Interface("iceServers", iceServer.URLs).Msg("Using ICE Servers provided by cloud")
		}

		if config.LocalIP == "" || net.ParseIP(config.LocalIP) == nil {
			scopedLogger.Info().Str("localIP", config.LocalIP).Msg("Local IP address not provided or invalid, won't set ICE address rewrite rules")
		} else {
			err := webrtcSettingEngine.SetICEAddressRewriteRules(
				webrtc.ICEAddressRewriteRule{
					External:        []string{config.LocalIP},
					AsCandidateType: webrtc.ICECandidateTypeSrflx,
					Mode:            webrtc.ICEAddressRewriteAppend,
				},
			)
			if err != nil {
				scopedLogger.Warn().Err(err).Msg("Failed to set ICE address rewrite rules")
			} else {
				scopedLogger.Info().Str("localIP", config.LocalIP).Msg("Setting ICE address rewrite rules")
			}
		}
	}

	activeSessionVideoCodecMu.Lock()
	if activeSessionVideoCodec != videoCodec {
		if err := nativeInstance.VideoSetEncoder(videoCodec); err != nil {
			activeSessionVideoCodecMu.Unlock()
			scopedLogger.Warn().Err(err).Int32("codec", videoCodec).Msg("Failed to set video encoder for session")
			return nil, err
		}
		activeSessionVideoCodec = videoCodec
	}
	activeSessionVideoCodecMu.Unlock()

	api := webrtc.NewAPI(webrtc.WithSettingEngine(webrtcSettingEngine))
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{iceServer},
	})
	if err != nil {
		scopedLogger.Warn().Err(err).Msg("Failed to create PeerConnection")
		return nil, err
	}

	session := &Session{peerConnection: peerConnection}
	session.done = make(chan struct{})
	session.rpcQueue = make(chan webrtc.DataChannelMessage, 256)
	session.initQueues()
	session.initKeysDownStateQueue()

	go func() {
		for {
			select {
			case <-session.done:
				return
			case msg := <-session.rpcQueue:
				// TODO: only use goroutine if the task is asynchronous
				go onRPCMessage(msg, session)
			}
		}
	}()

	for i := 0; i < len(session.hidQueue); i++ {
		go session.handleQueues(i)
	}

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		defer func() {
			if r := recover(); r != nil {
				scopedLogger.Error().Interface("error", r).Msg("Recovered from panic in DataChannel handler")
			}
		}()

		scopedLogger.Info().Str("label", d.Label()).Uint16("id", *d.ID()).Msg("New DataChannel")

		switch d.Label() {
		case "hidrpc":
			session.HidChannel = d
			d.OnMessage(getOnHidMessageHandler(session, scopedLogger, "hidrpc"))
		// we won't send anything over the unreliable channels
		case "hidrpc-unreliable-ordered":
			d.OnMessage(getOnHidMessageHandler(session, scopedLogger, "hidrpc-unreliable-ordered"))
		case "hidrpc-unreliable-nonordered":
			d.OnMessage(getOnHidMessageHandler(session, scopedLogger, "hidrpc-unreliable-nonordered"))
		case "rpc":
			session.RPCChannel = d
			d.OnMessage(func(msg webrtc.DataChannelMessage) {
				// Enqueue to ensure ordered processing. Select on done so a
				// late message during teardown can't block or panic.
				select {
				case session.rpcQueue <- msg:
				case <-session.done:
				}
			})
			// Wait for channel to be open before sending initial state
			d.OnOpen(func() {
				// triggerOTAStateUpdate(otaState.ToRPCState())
				triggerVideoStateUpdate()
				triggerUSBStateUpdate(currentUSBState())
				notifyFailsafeMode(session)
			})
		case "terminal":
			handleTerminalChannel(d)
		case "serial":
			handleSerialChannel(d)
		default:
			if strings.HasPrefix(d.Label(), uploadIdPrefix) {
				go handleUploadChannel(d)
			}
		}
	})

	// Determine codec based on global config
	var mimeType string
	if videoCodec == 1 {
		mimeType = webrtc.MimeTypeH265
		scopedLogger.Info().Msg("Creating VideoTrack with H.265 codec")
	} else {
		mimeType = webrtc.MimeTypeH264
		scopedLogger.Info().Msg("Creating VideoTrack with H.264 codec")
	}

	session.VideoTrack, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: mimeType}, "video", "kvm")
	if err != nil {
		scopedLogger.Warn().Err(err).Msg("Failed to create VideoTrack")
		return nil, err
	}

	rtpSender, err := peerConnection.AddTrack(session.VideoTrack)
	if err != nil {
		scopedLogger.Warn().Err(err).Msg("Failed to add VideoTrack to PeerConnection")
		return nil, err
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()
	var isConnected bool

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		scopedLogger.Info().Interface("candidate", candidate).Msg("WebRTC peerConnection has a new ICE candidate")
		if candidate != nil && config.ws != nil {
			err := wsjson.Write(context.Background(), config.ws, gin.H{"type": "new-ice-candidate", "data": candidate.ToJSON()})
			if err != nil {
				scopedLogger.Warn().Err(err).Msg("failed to write new-ice-candidate to WebRTC signaling channel")
			}
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		scopedLogger.Info().Str("connectionState", connectionState.String()).Msg("ICE Connection State has changed")
		if connectionState == webrtc.ICEConnectionStateConnected {
			if !isConnected {
				isConnected = true
				onActiveSessionsChanged()
				if incrActiveSessions() == 1 {
					onFirstSessionConnected()
				}
			}
		}
		//state changes on closing browser tab disconnected->failed, we need to manually close it
		if connectionState == webrtc.ICEConnectionStateFailed {
			scopedLogger.Debug().Msg("ICE Connection State is failed, closing peerConnection")
			_ = peerConnection.Close()
		}
		if connectionState == webrtc.ICEConnectionStateClosed {
			scopedLogger.Debug().Msg("ICE Connection State is closed, unmounting virtual media")
			// Only clear the global pointer if we're still the active session.
			// CompareAndSwap avoids clobbering a newer session that already
			// replaced us.
			if currentSessionPtr.CompareAndSwap(session, nil) {
				// Cancel any ongoing keyboard report multi when session closes
				cancelKeyboardMacro()
			}

			// Stop all queue consumers/senders exactly once. We deliberately do
			// not close the queue channels (senders select on done instead),
			// so this is panic-free even if pion fires Closed more than once.
			session.teardown()

			if session.shouldUmountVirtualMedia {
				if err := rpcUnmountImage(); err != nil {
					scopedLogger.Warn().Err(err).Msg("unmount image failed on connection close")
				}
			}
			if isConnected {
				isConnected = false
				onActiveSessionsChanged()
				if decrActiveSessions() == 0 {
					scopedLogger.Info().Msg("last session disconnected, stopping video stream")
					onLastSessionDisconnected()
				}
			}
		}
	})
	return session, nil
}

func onActiveSessionsChanged() {
	notifyFailsafeMode(getCurrentSession())
}

func onFirstSessionConnected() {
	notifyFailsafeMode(getCurrentSession())
	_ = nativeInstance.VideoStart()
	stopVideoSleepModeTicker()
}

func onLastSessionDisconnected() {
	_ = nativeInstance.VideoStop()
	startVideoSleepModeTicker()
}
