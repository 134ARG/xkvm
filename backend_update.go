package kvm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const (
	githubOwner       = "134ARG"
	githubRepo        = "xkvm"
	latestReleaseURL  = "https://api.github.com/repos/" + githubOwner + "/" + githubRepo + "/releases/latest"
	debInstallDir     = "/usr/lib/xkvm"
	selfUpdateScript  = "/usr/lib/xkvm/xkvm-self-update.sh"
	updateDebPath     = "/var/lib/xkvm/xkvm_update.deb"
	updateSystemdUnit = "xkvm-self-update"
)

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	HTMLURL string        `json:"html_url"`
	Body    string        `json:"body"`
	Assets  []githubAsset `json:"assets"`
}

// BackendUpdateInfo is returned to the UI for the version/update section.
type BackendUpdateInfo struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleaseNotes    string `json:"releaseNotes"`
	ReleaseURL      string `json:"releaseUrl"`
	DebURL          string `json:"debUrl"`
	CanAutoUpdate   bool   `json:"canAutoUpdate"`
}

// getLatestRelease fetches the latest GitHub release metadata.
func getLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error building release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status fetching latest release: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("error decoding latest release: %w", err)
	}

	return &release, nil
}

// detectDebInstall reports whether the backend was installed from the Debian
// package and can therefore self-update.
func detectDebInstall() bool {
	exe, err := os.Executable()
	if err != nil || !strings.HasPrefix(exe, debInstallDir) {
		return false
	}
	if err := exec.Command("dpkg-query", "-W", "-f=${Version}", "xkvm").Run(); err != nil {
		return false
	}
	return true
}

// findDebAsset returns the arm64 .deb asset download URL, if present.
func findDebAsset(release *githubRelease) string {
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, "xkvm_") && strings.HasSuffix(asset.Name, "_arm64.deb") {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

func rpcCheckBackendUpdate() (*BackendUpdateInfo, error) {
	current := GetBuiltAppVersion()

	release, err := getLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("error checking for backend update: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")

	currentVer, err := semver.NewVersion(current)
	if err != nil {
		return nil, fmt.Errorf("invalid current version %q: %w", current, err)
	}
	latestVer, err := semver.NewVersion(latest)
	if err != nil {
		return nil, fmt.Errorf("invalid latest version %q: %w", latest, err)
	}

	debURL := findDebAsset(release)

	return &BackendUpdateInfo{
		CurrentVersion:  current,
		LatestVersion:   latest,
		UpdateAvailable: latestVer.GreaterThan(currentVer),
		ReleaseNotes:    release.Body,
		ReleaseURL:      release.HTMLURL,
		DebURL:          debURL,
		CanAutoUpdate:   debURL != "" && detectDebInstall(),
	}, nil
}

// downloadFile streams a URL to dest.
func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status downloading %s: %s", url, resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("error writing %s: %w", dest, err)
	}
	return nil
}

func rpcTryUpdateBackend() error {
	if !detectDebInstall() {
		return fmt.Errorf("backend auto-update is only supported for Debian package installs")
	}

	info, err := rpcCheckBackendUpdate()
	if err != nil {
		return err
	}
	if info.DebURL == "" {
		return fmt.Errorf("no Debian package asset found in latest release")
	}

	logger.Info().Str("url", info.DebURL).Msg("downloading backend update")
	if err := downloadFile(info.DebURL, updateDebPath); err != nil {
		return err
	}

	// Run the install + service restart in a detached transient unit so it
	// survives this service being restarted mid-update.
	logger.Info().Msg("launching detached self-update")
	cmd := exec.Command("systemd-run", "--collect", "--unit="+updateSystemdUnit, selfUpdateScript, updateDebPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error launching self-update: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}
