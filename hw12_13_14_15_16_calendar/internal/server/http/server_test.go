package internalhttp

import (
	"context"
	"testing"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/logger"
)

func TestServerGracefulShutdown(t *testing.T) {
	logg := logger.New("debug")
	srv := NewServer(logg, nil, "127.0.0.1", 18080)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			t.Logf("server start err: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)
	cancel() // trigger shutdown

	time.Sleep(200 * time.Millisecond)
}
