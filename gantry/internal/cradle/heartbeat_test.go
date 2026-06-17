package cradle

import (
	"context"
	"testing"
)

func TestClientHeartbeat(t *testing.T) {
	client, svc := newTestClient(t)

	if err := client.Heartbeat(context.Background()); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}

	if got := svc.HeartbeatCalls(); got != 1 {
		t.Fatalf("HeartbeatCalls: got %d, want 1", got)
	}
}
