package heartbeat

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type CradleClient interface {
	Heartbeat(ctx context.Context) error
}

type Worker struct {
	clients  []CradleClient
	interval time.Duration
}

func New(clients []CradleClient, interval time.Duration) *Worker {
	return &Worker{
		clients,
		interval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	slog.Debug("starting heartbeat worker")
	w.fanOut(ctx)
	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.fanOut(ctx)
		}
	}
}

func (w *Worker) fanOut(ctx context.Context) {
	var wg sync.WaitGroup

	slog.Debug("heartbeat tick", "clients", len(w.clients))
	for _, c := range w.clients {
		wg.Go(func() { c.Heartbeat(ctx) })
	}
	wg.Wait()
}
