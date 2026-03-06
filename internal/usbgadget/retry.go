package usbgadget

// import (
// 	"fmt"
// 	"time"
// )

// const (
// 	maxRetries     = 3
// 	initialBackoff = 100 * time.Millisecond
// 	maxBackoff     = 2 * time.Second
// )

// // RetryConfig holds configuration for retry logic
// type RetryConfig struct {
// 	MaxRetries     int
// 	InitialBackoff time.Duration
// 	MaxBackoff     time.Duration
// 	Multiplier     float64
// }

// // DefaultRetryConfig returns the default retry configuration
// func DefaultRetryConfig() RetryConfig {
// 	return RetryConfig{
// 		MaxRetries:     maxRetries,
// 		InitialBackoff: initialBackoff,
// 		MaxBackoff:     maxBackoff,
// 		Multiplier:     2.0,
// 	}
// }

// // RetryWithBackoff retries a function with exponential backoff
// func (u *UsbGadget) RetryWithBackoff(operation string, fn func() error) error {
// 	return u.RetryWithBackoffConfig(operation, fn, DefaultRetryConfig())
// }

// // RetryWithBackoffConfig retries a function with custom retry configuration
// func (u *UsbGadget) RetryWithBackoffConfig(operation string, fn func() error, config RetryConfig) error {
// 	var lastErr error
// 	backoff := config.InitialBackoff

// 	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
// 		if attempt > 0 {
// 			u.log.Debug().
// 				Int("attempt", attempt).
// 				Int("max_retries", config.MaxRetries).
// 				Dur("backoff", backoff).
// 				Str("operation", operation).
// 				Msg("retrying operation")

// 			time.Sleep(backoff)

// 			// Exponential backoff with cap
// 			backoff = time.Duration(float64(backoff) * config.Multiplier)
// 			if backoff > config.MaxBackoff {
// 				backoff = config.MaxBackoff
// 			}
// 		}

// 		err := fn()
// 		if err == nil {
// 			if attempt > 0 {
// 				u.log.Info().
// 					Int("attempt", attempt).
// 					Str("operation", operation).
// 					Msg("operation succeeded after retry")
// 			}
// 			return nil
// 		}

// 		lastErr = err

// 		// Check if error is retryable
// 		if !isRetryableError(err) {
// 			u.log.Debug().
// 				Err(err).
// 				Str("operation", operation).
// 				Msg("error is not retryable, giving up")
// 			return err
// 		}

// 		u.log.Warn().
// 			Err(err).
// 			Int("attempt", attempt).
// 			Int("max_retries", config.MaxRetries).
// 			Str("operation", operation).
// 			Msg("operation failed, will retry")
// 	}

// 	return fmt.Errorf("operation %s failed after %d retries: %w", operation, config.MaxRetries, lastErr)
// }

// // isRetryableError determines if an error is worth retrying
// func isRetryableError(err error) bool {
// 	if err == nil {
// 		return false
// 	}

// 	errStr := err.Error()

// 	// Retryable error patterns
// 	retryablePatterns := []string{
// 		"device or resource busy",
// 		"resource temporarily unavailable",
// 		"no such device",
// 		"operation not permitted",
// 		"text file busy",
// 	}

// 	for _, pattern := range retryablePatterns {
// 		if contains(errStr, pattern) {
// 			return true
// 		}
// 	}

// 	return false
// }

// // contains checks if a string contains a substring (case-insensitive)
// func contains(s, substr string) bool {
// 	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
// 		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
// }

// func findSubstring(s, substr string) bool {
// 	for i := 0; i <= len(s)-len(substr); i++ {
// 		if s[i:i+len(substr)] == substr {
// 			return true
// 		}
// 	}
// 	return false
// }
