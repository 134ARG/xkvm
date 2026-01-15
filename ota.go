package kvm

import (
	"fmt"

	"github.com/134ARG/xkvm/internal/ota"
	"github.com/Masterminds/semver/v3"
)

var builtAppVersion = "0.1.0+dev"

// GetBuiltAppVersion returns the built-in app version
func GetBuiltAppVersion() string {
	return builtAppVersion
}

// GetLocalVersion returns the local version of the system and app
func GetLocalVersion() (systemVersion *semver.Version, appVersion *semver.Version, err error) {
	appVersion, err = semver.NewVersion(builtAppVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid built-in app version: %w", err)
	}

	// systemVersionBytes, err := os.ReadFile("/version")
	// if err != nil {
	// 	return nil, appVersion, fmt.Errorf("error reading system version: %w", err)
	// }

	// systemVersion, err = semver.NewVersion(strings.TrimSpace(string(systemVersionBytes)))
	// if err != nil {
	// 	return nil, appVersion, fmt.Errorf("invalid system version: %w", err)
	// }

	return nil, appVersion, nil
}

func rpcGetLocalVersion() (*ota.LocalMetadata, error) {
	systemVersion, appVersion, err := GetLocalVersion()
	if err != nil {
		return nil, fmt.Errorf("error getting local version: %w", err)
	}

	systemVersionStr := ""
	if systemVersion != nil {
		systemVersionStr = systemVersion.String()
	}

	appVersionStr := ""
	if appVersion != nil {
		appVersionStr = appVersion.String()
	}

	return &ota.LocalMetadata{
		AppVersion:    appVersionStr,
		SystemVersion: systemVersionStr,
	}, nil
}
