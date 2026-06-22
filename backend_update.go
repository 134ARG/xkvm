package kvm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
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

// downloadStallTimeout fails a download only when no data arrives for this long.
// A slow-but-moving transfer is never interrupted.
const downloadStallTimeout = 30 * time.Second

// progressWriter resets the stall timer on every write and reports cumulative
// bytes (throttled) for UI progress.
type progressWriter struct {
	total      int64
	written    int64
	lastEmit   time.Time
	keepAlive  func()
	onProgress func(downloaded, total int64)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.written += int64(n)
	p.keepAlive()
	if time.Since(p.lastEmit) >= 250*time.Millisecond {
		p.lastEmit = time.Now()
		p.onProgress(p.written, p.total)
	}
	return n, nil
}

// updateCancel cancels the in-flight download, if any. Only valid during the
// download phase; cleared once the install is launched (point of no return).
var (
	updateMu       sync.Mutex
	updateCancel   context.CancelFunc
	updateCanceled bool
)

// downloadFile streams a URL to dest and returns the number of bytes written.
// onProgress is called periodically with the bytes downloaded so far and the
// total size (-1 if unknown). It fails if the transfer stalls, not if it is
// slow, and aborts if parent is canceled.
func downloadFile(parent context.Context, url, dest string, onProgress func(downloaded, total int64)) (int64, error) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	stall := time.AfterFunc(downloadStallTimeout, cancel)
	defer stall.Stop()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("error building download request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status downloading %s: %s", url, resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return 0, fmt.Errorf("error creating %s: %w", dest, err)
	}
	defer out.Close()

	progress := &progressWriter{
		total:      resp.ContentLength,
		keepAlive:  func() { stall.Reset(downloadStallTimeout) },
		onProgress: onProgress,
	}
	n, err := io.Copy(out, io.TeeReader(resp.Body, progress))
	if err != nil {
		if ctx.Err() != nil {
			return n, fmt.Errorf("download stalled (no data for %s)", downloadStallTimeout)
		}
		return n, fmt.Errorf("error writing %s: %w", dest, err)
	}
	return n, nil
}

// rpcTryUpdateBackend validates synchronously, then runs the download and
// install in the background, reporting outcome via events.
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

	go runBackendUpdate(info.DebURL, getCurrentSession())
	return nil
}

func runBackendUpdate(debURL string, session *Session) {
	emit := func(event string, params any) { writeJSONRPCEvent(event, params, session) }
	emitProgress := func(downloaded, total int64) {
		emit("backendUpdateProgress", map[string]int64{"downloaded": downloaded, "total": total})
	}
	fail := func(err error) {
		logger.Error().Err(err).Msg("backend update failed")
		emit("backendUpdateFailed", map[string]string{"error": err.Error()})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updateMu.Lock()
	updateCancel = cancel
	updateCanceled = false
	updateMu.Unlock()
	defer func() {
		updateMu.Lock()
		updateCancel = nil
		updateMu.Unlock()
	}()

	logger.Info().Str("url", debURL).Msg("downloading backend update")
	emitProgress(0, -1)
	n, err := downloadFile(ctx, debURL, updateDebPath, emitProgress)
	if err != nil {
		updateMu.Lock()
		canceled := updateCanceled
		updateMu.Unlock()
		if canceled {
			logger.Info().Msg("backend update canceled")
			emit("backendUpdateCanceled", nil)
			return
		}
		fail(err)
		return
	}
	emitProgress(n, n)
	logger.Info().Int64("bytes", n).Str("path", updateDebPath).Msg("backend update downloaded")

	// Past the point of no return: the install is launched detached, so cancel
	// no longer applies.
	updateMu.Lock()
	updateCancel = nil
	updateMu.Unlock()

	// Run the install + service restart in a detached transient unit so it
	// survives this service being restarted mid-update.
	logger.Info().Msg("launching detached self-update")
	cmd := exec.Command("systemd-run", "--collect", "--unit="+updateSystemdUnit, selfUpdateScript, updateDebPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		fail(fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out))))
		return
	}
	emit("backendUpdateInstalling", nil)
}

// rpcCancelBackendUpdate aborts an in-progress download. It is a no-op once the
// install has been launched.
func rpcCancelBackendUpdate() error {
	updateMu.Lock()
	defer updateMu.Unlock()
	if updateCancel != nil {
		updateCanceled = true
		updateCancel()
	}
	return nil
}
