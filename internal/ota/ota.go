package ota

import (
	"net/http"
)

// HttpClient is the interface for the HTTP client
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// UpdateParams represents the parameters for the update
type UpdateParams struct {
	DeviceID          string            `json:"deviceID"`
	Components        map[string]string `json:"components"`
	IncludePreRelease bool              `json:"includePreRelease"`
	ResetConfig       bool              `json:"resetConfig"`
	// RequestID is a unique identifier for the update request
	// When it's set, detailed trace logs will be enabled (if the log level is Trace)
	RequestID string
}
