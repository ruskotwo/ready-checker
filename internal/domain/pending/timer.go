package pending

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var minDuration = 10 * time.Minute
var maxDuration = 5 * time.Hour

type Timer struct {
	logger   *slog.Logger
	mu       sync.Mutex
	channels map[string]chan bool
}

func NewTimer(logger *slog.Logger) *Timer {
	return &Timer{
		logger:   logger,
		channels: make(map[string]chan bool),
	}
}

func (t *Timer) Start(id string, duration time.Duration) (chan bool, time.Duration) {
	if duration > maxDuration {
		duration = maxDuration
	} else if duration < minDuration {
		duration = minDuration
	}

	t.logger.Info(fmt.Sprintf("Start timer %fm for %s", duration.Minutes(), id))

	t.mu.Lock()
	defer t.mu.Unlock()

	if ch, exists := t.channels[id]; exists {
		t.logger.Info(fmt.Sprintf("Restart timer for %s", id))
		ch <- false
	}

	ch := make(chan bool, 1)
	t.channels[id] = ch

	go func() {
		time.Sleep(duration)
		ch <- true

		t.mu.Lock()
		delete(t.channels, id)
		t.mu.Unlock()
		close(ch)
	}()

	return ch, duration
}

func (t *Timer) Stop(id string) {
	t.logger.Info(fmt.Sprintf("Stop timer for %s", id))

	t.mu.Lock()
	defer t.mu.Unlock()

	if ch, exists := t.channels[id]; exists {
		ch <- false
	}
}
