package logging

import (
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/rs/zerolog"
)

func TestConcurrentLoggerCreation(t *testing.T) {
	logger := NewLogger(zerolog.New(io.Discard))

	const goroutines = 64
	const iterations = 200

	var wg sync.WaitGroup
	start := make(chan struct{})

	for worker := 0; worker < goroutines; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			<-start

			for i := 0; i < iterations; i++ {
				scope := fmt.Sprintf("ice-%d-%d", worker, i)
				if logger.getLogger(scope) == nil {
					t.Errorf("logger for scope %q is nil", scope)
				}
			}
		}(worker)
	}

	close(start)
	wg.Wait()
}
