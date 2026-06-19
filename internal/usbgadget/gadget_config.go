package usbgadget

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
)

// getOrderedConfigItems returns the config items sorted by their .order field.
// configfs records function symlink creation order, so a stable order here gives
// a deterministic, correct gadget layout.
func (u *UsbGadget) getOrderedConfigItems() orderedGadgetConfigItems {
	items := make(orderedGadgetConfigItems, 0, len(u.configMap))
	for key, item := range u.configMap {
		items = append(items, gadgetConfigItemWithKey{key, item})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].item.order < items[j].item.order
	})

	return items
}

// writeGadget builds the entire USB gadget tree in configfs from scratch.
//
// The set of functions and their order is statically known, and callers always
// invoke this on a freshly cleaned-up tree (see cleanupStaleGadget), so there is
// no need for idempotent diffing or dependency resolution: we simply create
// directories, write attributes, write report descriptors, create the enable
// symlinks in order, and finally bind by writing the UDC.
//
// configfs orders functions in a configuration by the order in which their
// symlinks are created, so iterating getOrderedConfigItems() (sorted by .order)
// gives a stable, correct enumeration order.
func (u *UsbGadget) writeGadget() error {
	if u.udc == "" {
		return u.logError("no udc available, cannot write gadget", nil)
	}

	if err := u.ensureConfigFSMounted(); err != nil {
		return u.logError("failed to mount configfs", err)
	}

	if err := os.MkdirAll(u.kvmGadgetPath, 0755); err != nil {
		return u.logError("failed to create gadget path", err)
	}

	for _, val := range u.getOrderedConfigItems() {
		if !u.isGadgetConfigItemEnabled(val.key) {
			continue
		}
		if err := u.writeGadgetItem(val.item); err != nil {
			return u.logError(fmt.Sprintf("failed to write gadget item %s", val.key), err)
		}
	}

	// Bind the gadget to the UDC. This must happen after everything else is in
	// place, including the function symlinks.
	udcPath := path.Join(u.kvmGadgetPath, "UDC")
	if err := os.WriteFile(udcPath, []byte(u.udc), 0644); err != nil {
		return u.logError("failed to bind UDC", err)
	}

	u.log.Info().Str("udc", u.udc).Msg("USB gadget written and bound")
	return nil
}

// writeGadgetItem writes a single config item: its function-side attributes and
// report descriptor, its config-side attributes, and (for plain functions) the
// enable symlink into configs/c.1.
func (u *UsbGadget) writeGadgetItem(item gadgetConfigItem) error {
	// Function-side path: gadget root when item.path is nil (e.g. base device
	// descriptors), otherwise functions/<device> or strings/0x409.
	itemPath := joinPath(u.kvmGadgetPath, item.path)
	if err := os.MkdirAll(itemPath, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", itemPath, err)
	}

	if err := writeAttrs(itemPath, item.attrs); err != nil {
		return err
	}

	if item.reportDesc != nil {
		reportDescPath := path.Join(itemPath, "report_desc")
		if err := os.WriteFile(reportDescPath, item.reportDesc, 0644); err != nil {
			return fmt.Errorf("write report_desc %s: %w", reportDescPath, err)
		}
	}

	// Config-side attributes (e.g. MaxPower, configuration string) live under
	// configs/c.1[/<configPath>].
	if len(item.configAttrs) > 0 {
		configItemPath := joinPath(u.configC1Path, item.configPath)
		if err := os.MkdirAll(configItemPath, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", configItemPath, err)
		}
		if err := writeAttrs(configItemPath, item.configAttrs); err != nil {
			return err
		}
	}

	// A plain function (configPath set, no configAttrs) is enabled by symlinking
	// it into the configuration. configfs records the creation order.
	if item.configPath != nil && item.configAttrs == nil {
		linkPath := joinPath(u.configC1Path, item.configPath)
		if err := os.Symlink(itemPath, linkPath); err != nil && !os.IsExist(err) {
			return fmt.Errorf("symlink %s -> %s: %w", linkPath, itemPath, err)
		}
	}

	return nil
}

func writeAttrs(basePath string, attrs gadgetAttributes) error {
	for key, val := range attrs {
		attrPath := filepath.Join(basePath, key)
		if err := os.WriteFile(attrPath, []byte(val), 0644); err != nil {
			return fmt.Errorf("write attr %s: %w", attrPath, err)
		}
	}
	return nil
}

// ensureConfigFSMounted mounts configfs if the usb_gadget directory is not
// already present.
func (u *UsbGadget) ensureConfigFSMounted() error {
	if _, err := os.Stat(gadgetPath); err == nil {
		return nil
	}
	return mountConfigFS(configFSPath)
}
