package usbgadget

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// cleanupStaleGadget removes any leftover USB gadget configuration from previous runs
func (u *UsbGadget) cleanupStaleGadget() error {
	// Check if the gadget directory exists
	if _, err := os.Stat(u.kvmGadgetPath); os.IsNotExist(err) {
		u.log.Debug().Msg("no stale gadget found, proceeding with fresh init")
		return nil
	}

	u.log.Info().Str("path", u.kvmGadgetPath).Msg("found stale USB gadget, cleaning up")

	// Try to unbind UDC if it's bound. Read the gadget's own UDC file so
	// stale cleanup also works after a service restart, before u.udc is set.
	if err := u.unbindStaleGadgetUDC(); err != nil {
		u.log.Warn().Err(err).Msg("failed to unbind UDC during cleanup")
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

func (u *UsbGadget) unbindStaleGadgetUDC() error {
	udcPath := filepath.Join(u.kvmGadgetPath, "UDC")
	content, err := os.ReadFile(udcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read UDC file: %w", err)
	}

	if strings.TrimSpace(string(content)) == "" {
		return nil
	}
	return os.WriteFile(udcPath, []byte("\n"), 0644)
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
