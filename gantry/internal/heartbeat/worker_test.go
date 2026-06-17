package heartbeat_test

import (
	"context"
	"testing"
	"time"

	"github.com/ratdaddy/blockcloset/gantry/internal/heartbeat"
)

type fakeClient struct {
	called chan struct{}
}

func (f *fakeClient) Heartbeat(_ context.Context) error {
	f.called <- struct{}{}
	return nil
}

func TestWorker_CallsHeartbeatOnTick(t *testing.T) {
	called := make(chan struct{}, 4)
	fake1 := &fakeClient{called: called}
	fake2 := &fakeClient{called: called}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := heartbeat.New([]heartbeat.CradleClient{fake1, fake2}, 10*time.Millisecond)

	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	for range 4 {
		select {
		case <-called:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for heartbeat calls")
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for worker to stop")
	}
}
