package kvm

import (
	"context"
	"time"

	"github.com/coder/websocket/wsjson"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const (
	// WebsocketPingInterval is the interval at which the websocket client sends ping messages to the cloud
	WebsocketPingInterval = 15 * time.Second
)

var (
	metricConnectionLastPingTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "xkvm_connection_last_ping_timestamp_seconds",
			Help: "The timestamp when the last ping response was received",
		},
		[]string{"type", "source"},
	)
	metricConnectionLastPingReceivedTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "xkvm_connection_last_ping_received_timestamp_seconds",
			Help: "The timestamp when the last ping request was received",
		},
		[]string{"type", "source"},
	)
	metricConnectionLastPingDuration = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "xkvm_connection_last_ping_duration_seconds",
			Help: "The duration of the last ping response",
		},
		[]string{"type", "source"},
	)
	metricConnectionPingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "xkvm_connection_ping_duration_seconds",
			Help: "The duration of the ping response",
			Buckets: []float64{
				0.1, 0.5, 1, 10,
			},
		},
		[]string{"type", "source"},
	)
	metricConnectionTotalPingSentCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xkvm_connection_ping_sent_total",
			Help: "The total number of pings sent to the connection",
		},
		[]string{"type", "source"},
	)
	metricConnectionTotalPingReceivedCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xkvm_connection_ping_received_total",
			Help: "The total number of pings received from the connection",
		},
		[]string{"type", "source"},
	)
	metricConnectionSessionRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xkvm_connection_session_requests_total",
			Help: "The total number of session requests received",
		},
		[]string{"type", "source"},
	)
	metricConnectionSessionRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "xkvm_connection_session_request_duration_seconds",
			Help: "The duration of session requests",
			Buckets: []float64{
				0.1, 0.5, 1, 10,
			},
		},
		[]string{"type", "source"},
	)
	metricConnectionLastSessionRequestTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "xkvm_connection_last_session_request_timestamp_seconds",
			Help: "The timestamp of the last session request",
		},
		[]string{"type", "source"},
	)
	metricConnectionLastSessionRequestDuration = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "xkvm_connection_last_session_request_duration",
			Help: "The duration of the last session request",
		},
		[]string{"type", "source"},
	)
)

func handleSessionRequest(
	ctx context.Context,
	c *websocket.Conn,
	req WebRTCSessionRequest,
	isCloudConnection bool,
	source string,
	scopedLogger *zerolog.Logger,
) error {
	var sourceType = "local"

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		metricConnectionLastSessionRequestDuration.WithLabelValues(sourceType, source).Set(v)
		metricConnectionSessionRequestDuration.WithLabelValues(sourceType, source).Observe(v)
	}))
	defer timer.ObserveDuration()

	session, err := newSession(SessionConfig{
		ws:                  c,
		IsCloud:             isCloudConnection,
		LocalIP:             req.IP,
		ICEServers:          req.ICEServers,
		Logger:              scopedLogger,
		PreferredVideoCodec: req.PreferredVideoCodec,
	})
	if err != nil {
		_ = wsjson.Write(context.Background(), c, gin.H{"error": err})
		return err
	}

	sd, err := session.ExchangeOffer(req.Sd)
	if err != nil {
		_ = wsjson.Write(context.Background(), c, gin.H{"error": err})
		return err
	}
	if cs := getCurrentSession(); cs != nil {
		writeJSONRPCEvent("otherSessionConnected", nil, cs)
		peerConn := cs.peerConnection
		go func() {
			time.Sleep(1 * time.Second)
			_ = peerConn.Close()
		}()
	}

	cloudLogger.Info().Interface("session", session).Msg("new session accepted")
	cloudLogger.Trace().Interface("session", session).Msg("new session accepted")

	// Cancel any ongoing keyboard macro when session changes
	cancelKeyboardMacro()

	setCurrentSession(session)
	_ = wsjson.Write(context.Background(), c, gin.H{"type": "answer", "data": sd})
	return nil
}
