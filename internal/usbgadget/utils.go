package usbgadget

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type ByteSlice []byte

func (s ByteSlice) MarshalJSON() ([]byte, error) {
	vals := make([]int, len(s))
	for i, v := range s {
		vals[i] = int(v)
	}
	return json.Marshal(vals)
}

func (s *ByteSlice) UnmarshalJSON(data []byte) error {
	var vals []int
	if err := json.Unmarshal(data, &vals); err != nil {
		return err
	}
	*s = make([]byte, len(vals))
	for i, v := range vals {
		if v < 0 || v > 255 {
			return fmt.Errorf("value %d out of byte range", v)
		}
		(*s)[i] = byte(v)
	}
	return nil
}

func joinPath(basePath string, paths []string) string {
	pathArr := append([]string{basePath}, paths...)
	return filepath.Join(pathArr...)
}

func (u *UsbGadget) writeWithTimeout(file *os.File, data []byte) (n int, err error) {
	return u.writeWithTimeoutDuration(file, data, hidWriteTimeout)
}

func (u *UsbGadget) writeWithTimeoutDuration(file *os.File, data []byte, timeout time.Duration) (n int, err error) {
	if err := file.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return -1, err
	}

	n, err = file.Write(data)
	if err == nil {
		return
	}

	u.log.Trace().
		Str("file", file.Name()).
		Bytes("data", data).
		Err(err).
		Msg("write failed")

	if errors.Is(err, os.ErrDeadlineExceeded) {
		return n, fmt.Errorf("write to %s timed out: %w", file.Name(), err)
	}

	return
}

func (u *UsbGadget) logWithSuppression(counterName string, every int, logger *zerolog.Logger, err error, msg string, args ...any) {
	u.logSuppressionLock.Lock()
	defer u.logSuppressionLock.Unlock()

	if _, ok := u.logSuppressionCounter[counterName]; !ok {
		u.logSuppressionCounter[counterName] = 0
	} else {
		u.logSuppressionCounter[counterName]++
	}

	l := logger.With().Int("counter", u.logSuppressionCounter[counterName]).Logger()

	if u.logSuppressionCounter[counterName]%every == 0 {
		if err != nil {
			l.Error().Err(err).Msgf(msg, args...)
		} else {
			l.Error().Msgf(msg, args...)
		}
	}
}

func (u *UsbGadget) resetLogSuppressionCounter(counterName string) {
	u.logSuppressionLock.Lock()
	defer u.logSuppressionLock.Unlock()

	if _, ok := u.logSuppressionCounter[counterName]; !ok {
		u.logSuppressionCounter[counterName] = 0
	}
}

func unlockWithLog(lock *sync.Mutex, logger *zerolog.Logger, msg string, args ...any) {
	logger.Trace().Msgf(msg, args...)
	lock.Unlock()
}
