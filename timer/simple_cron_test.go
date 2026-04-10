package timer

import (
	"testing"
	"time"
)

func TestSimpleCronStopCanExitCheckerGoroutine(t *testing.T) {
	sc := NewSimpleCron(time.Minute)

	done := make(chan struct{})
	go func() {
		sc.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop() timeout, checker goroutine may not exit")
	}
}

func TestSimpleCronStopCanBeCalledMultiTimes(t *testing.T) {
	sc := NewSimpleCron(time.Minute)
	sc.Stop()
	sc.Stop()
}

