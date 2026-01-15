package ota

// LocalMetadata represents the local metadata of the system
type LocalMetadata struct {
	AppVersion    string `json:"appVersion"`
	SystemVersion string `json:"systemVersion"`
}
