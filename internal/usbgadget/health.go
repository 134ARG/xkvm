package usbgadget

import (
	"context"
	"time"
)

const (
	healthCheckInterval = 30 * time.Second
	healthCheckTimeout  = 5 * time.Second
)

// HealthStatus represents the health status of the USB gadget
type HealthStatus struct {
	Healthy      bool      `json:"healthy"`
	LastCheck    time.Time `json:"last_check"`
	ErrorMessage string    `json:"error_message,omitempty"`
	UDCBound     bool      `json:"udc_bound"`
	GadgetExists bool      `json:"gadget_exists"`
	ConfigExists bool      `json:"config_exists"`
}

// StartHealthCheck starts a background health check routine
func (u *UsbGadget) StartHealthCheck(ctx context.Context) {
	go u.healthCheckLoop(ctx)
}

// healthCheckLoop runs periodic health checks
func (u *UsbGadget) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	u.log.Info().Dur("interval", healthCheckInterval).Msg("starting USB gadget health check loop")

	for {
		select {
		case <-ctx.Done():
			u.log.Info().Msg("stopping USB gadget health check loop")
			return
		case <-ticker.C:
			u.performHealthCheck()
		}
	}
}

// performHealthCheck performs a health check and attempts recovery if needed
func (u *UsbGadget) performHealthCheck() {
	status := u.GetHealthStatus()

	if status.Healthy {
		u.log.Debug().Msg("USB gadget health check passed")
		return
	}

	u.log.Warn().
		Str("error", status.ErrorMessage).
		Bool("udc_bound", status.UDCBound).
		Bool("gadget_exists", status.GadgetExists).
		Bool("config_exists", status.ConfigExists).
		Msg("USB gadget health check failed, attempting recovery")

	// Attempt recovery
	if err := u.attemptRecovery(); err != nil {
		u.log.Error().Err(err).Msg("failed to recover USB gadget")
	} else {
		u.log.Info().Msg("successfully recovered USB gadget")
	}
}

// GetHealthStatus returns the current health status of the USB gadget
func (u *UsbGadget) GetHealthStatus() HealthStatus {
	status := HealthStatus{
		Healthy:   true,
		LastCheck: time.Now(),
	}

	// Check if gadget is initialized
	if !u.IsInitialized() {
		status.Healthy = false
		status.ErrorMessage = "USB gadget not initialized"
		return status
	}

	// Check if gadget directory exists
	if err := u.ValidateState(); err != nil {
		status.Healthy = false
		status.ErrorMessage = err.Error()
		status.GadgetExists = false
		return status
	}
	status.GadgetExists = true
	status.ConfigExists = true

	// Check if UDC is bound
	bound, err := u.IsUDCBound()
	if err != nil {
		status.Healthy = false
		status.ErrorMessage = "failed to check UDC bind state: " + err.Error()
		return status
	}
	status.UDCBound = bound

	// If UDC is not bound, it's not healthy
	if !bound {
		status.Healthy = false
		status.ErrorMessage = "UDC not bound"
	}

	return status
}

// attemptRecovery attempts to recover the USB gadget from an unhealthy state
func (u *UsbGadget) attemptRecovery() error {
	u.log.Info().Msg("attempting USB gadget recovery")

	// Try to acquire lock for recovery
	lockFile, err := u.AcquireLock()
	if err != nil {
		u.log.Warn().Err(err).Msg("failed to acquire lock for recovery, proceeding anyway")
	} else {
		defer u.ReleaseLock(lockFile)
	}

	// Check if we just need to rebind the UDC
	bound, err := u.IsUDCBound()
	if err == nil && !bound {
		u.log.Info().Msg("attempting to rebind UDC")
		if err := u.BindUDC(); err != nil {
			u.log.Warn().Err(err).Msg("failed to rebind UDC, will try full recovery")
		} else {
			u.log.Info().Msg("successfully rebound UDC")
			return nil
		}
	}

	// Full recovery: cleanup and reinitialize
	u.log.Info().Msg("performing full USB gadget recovery")

	// Cleanup stale state
	if err := u.cleanupStaleGadget(); err != nil {
		u.log.Warn().Err(err).Msg("cleanup failed during recovery")
	}

	// Reinitialize
	u.configLock.Lock()
	defer u.configLock.Unlock()

	return u.RetryWithBackoff("recovery_init", func() error {
		return u.configureUsbGadget(false)
	})
}

// CheckHealth performs an immediate health check and returns the status
func (u *UsbGadget) CheckHealth() HealthStatus {
	return u.GetHealthStatus()
}
