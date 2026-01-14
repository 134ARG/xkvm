package usbgadget

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const lockFilePath = "/var/lock/jetkvm-usb.lock"
const lockTimeout = 10 * time.Second

// cleanupStaleGadget removes any leftover USB gadget configuration from previous runs
func (u *UsbGadget) cleanupStaleGadget() error {
	// Check if the gadget directory exists
	if _, err := os.Stat(u.kvmGadgetPath); os.IsNotExist(err) {
		u.log.Debug().Msg("no stale gadget found, proceeding with fresh init")
		return nil
	}

	u.log.Info().Str("path", u.kvmGadgetPath).Msg("found stale USB gadget, cleaning up")

	// Log directory contents for debugging
	if entries, err := os.ReadDir(u.kvmGadgetPath); err == nil {
		u.log.Debug().Int("entries", len(entries)).Msg("gadget directory contents")
		for _, entry := range entries {
			u.log.Trace().Str("name", entry.Name()).Bool("is_dir", entry.IsDir()).Msg("entry")
		}
	}

	// Try to unbind UDC if it's bound
	if err := u.unbindUDCIfBound(); err != nil {
		u.log.Warn().Err(err).Msg("failed to unbind UDC during cleanup")
		// Continue anyway, as the UDC might not be bound
	}

	// Remove all symlinks in the config directory
	if err := u.removeConfigSymlinks(); err != nil {
		u.log.Warn().Err(err).Msg("failed to remove config symlinks")
	}

	// Remove function directories
	if err := u.removeFunctionDirs(); err != nil {
		u.log.Warn().Err(err).Msg("failed to remove function directories")
	}

	// Remove config directories
	configsPath := filepath.Join(u.kvmGadgetPath, "configs")
	if err := os.RemoveAll(configsPath); err != nil && !os.IsNotExist(err) {
		u.log.Warn().Err(err).Msg("failed to remove configs directory")
	}

	// Remove strings directories
	stringsPath := filepath.Join(u.kvmGadgetPath, "strings")
	if err := os.RemoveAll(stringsPath); err != nil && !os.IsNotExist(err) {
		u.log.Warn().Err(err).Msg("failed to remove strings directory")
	}

	// Finally, remove the gadget directory itself
	if err := os.Remove(u.kvmGadgetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove gadget directory: %w", err)
	}

	u.log.Info().Msg("stale USB gadget cleaned up successfully")
	return nil
}

// unbindUDCIfBound attempts to unbind the UDC if it's currently bound
func (u *UsbGadget) unbindUDCIfBound() error {
	if u.udc == "" {
		return nil
	}

	// Check if UDC file exists and has content
	udcFilePath := filepath.Join(u.kvmGadgetPath, "UDC")
	content, err := os.ReadFile(udcFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // UDC file doesn't exist, nothing to unbind
		}
		return fmt.Errorf("failed to read UDC file: %w", err)
	}

	// If UDC is bound (file has content), unbind it
	if len(strings.TrimSpace(string(content))) > 0 {
		u.log.Info().Str("udc", u.udc).Msg("unbinding UDC during cleanup")
		// Write empty string to unbind
		if err := os.WriteFile(udcFilePath, []byte("\n"), 0644); err != nil {
			return fmt.Errorf("failed to unbind UDC: %w", err)
		}
		// Give it a moment to unbind
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// removeConfigSymlinks removes all symlinks in the config directory
func (u *UsbGadget) removeConfigSymlinks() error {
	if _, err := os.Stat(u.configC1Path); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(u.configC1Path)
	if err != nil {
		return fmt.Errorf("failed to read config directory: %w", err)
	}

	for _, entry := range entries {
		// Skip strings directory
		if entry.Name() == "strings" {
			continue
		}

		entryPath := filepath.Join(u.configC1Path, entry.Name())

		// Check if it's a symlink
		info, err := os.Lstat(entryPath)
		if err != nil {
			u.log.Warn().Err(err).Str("path", entryPath).Msg("failed to stat entry")
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			u.log.Debug().Str("path", entryPath).Msg("removing symlink")
			if err := os.Remove(entryPath); err != nil {
				u.log.Warn().Err(err).Str("path", entryPath).Msg("failed to remove symlink")
			}
		}
	}

	return nil
}

// removeFunctionDirs removes function directories (functions/*)
func (u *UsbGadget) removeFunctionDirs() error {
	functionsPath := filepath.Join(u.kvmGadgetPath, "functions")
	if _, err := os.Stat(functionsPath); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(functionsPath)
	if err != nil {
		return fmt.Errorf("failed to read functions directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(functionsPath, entry.Name())
		u.log.Debug().Str("path", entryPath).Msg("removing function directory")

		if err := os.RemoveAll(entryPath); err != nil {
			u.log.Warn().Err(err).Str("path", entryPath).Msg("failed to remove function directory")
		}
	}

	return nil
}

// AcquireLock acquires a file lock to prevent concurrent USB gadget operations
func (u *UsbGadget) AcquireLock() (*os.File, error) {
	// Ensure lock directory exists
	lockDir := filepath.Dir(lockFilePath)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Open or create lock file
	lockFile, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to acquire lock with timeout
	deadline := time.Now().Add(lockTimeout)
	for {
		err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			u.log.Debug().Msg("acquired USB gadget lock")
			return lockFile, nil
		}

		if time.Now().After(deadline) {
			lockFile.Close()
			return nil, fmt.Errorf("timeout acquiring USB gadget lock after %v", lockTimeout)
		}

		// Wait a bit before retrying
		time.Sleep(100 * time.Millisecond)
	}
}

// ReleaseLock releases the file lock
func (u *UsbGadget) ReleaseLock(lockFile *os.File) error {
	if lockFile == nil {
		return nil
	}

	u.log.Debug().Msg("releasing USB gadget lock")

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN); err != nil {
		lockFile.Close()
		return fmt.Errorf("failed to unlock: %w", err)
	}

	return lockFile.Close()
}

// IsInitialized checks if the USB gadget is properly initialized
func (u *UsbGadget) IsInitialized() bool {
	if u == nil {
		return false
	}

	// Check if gadget directory exists
	if _, err := os.Stat(u.kvmGadgetPath); os.IsNotExist(err) {
		return false
	}

	// Check if UDC is set
	if u.udc == "" {
		return false
	}

	return true
}

// ValidateState checks if the USB gadget is in a valid state
func (u *UsbGadget) ValidateState() error {
	if !u.IsInitialized() {
		return fmt.Errorf("USB gadget not initialized")
	}

	// Check if gadget directory exists
	if _, err := os.Stat(u.kvmGadgetPath); err != nil {
		return fmt.Errorf("gadget directory missing: %w", err)
	}

	// Check if config directory exists
	if _, err := os.Stat(u.configC1Path); err != nil {
		return fmt.Errorf("config directory missing: %w", err)
	}

	return nil
}

// Cleanup performs a full cleanup of the USB gadget
func (u *UsbGadget) Cleanup() error {
	u.log.Info().Msg("performing USB gadget cleanup")

	// Close resources first
	if err := u.Close(); err != nil {
		u.log.Warn().Err(err).Msg("error during Close()")
	}

	// Acquire lock for cleanup
	lockFile, err := u.AcquireLock()
	if err != nil {
		u.log.Warn().Err(err).Msg("failed to acquire lock for cleanup, proceeding anyway")
	} else {
		defer u.ReleaseLock(lockFile)
	}

	// Perform cleanup
	return u.cleanupStaleGadget()
}
