package bootstrap

import (
	"context"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/ratdaddy/blockcloset/gantry/internal/config"
	"github.com/ratdaddy/blockcloset/gantry/internal/testutil"
)

func TestInitSeedsCradleServer(t *testing.T) {
	t.Setenv("GANTRY_CRADLE_ADDR", "127.0.0.1:9444")
	config.CradleServerID = ""
	config.Init()

	ctx := context.Background()
	cradle := testutil.NewFakeCradleStore()
	st := testutil.NewFakeStore(testutil.WithCradles(cradle))

	if err := Init(ctx, st); err != nil {
		t.Fatalf("Init: unexpected error: %v", err)
	}

	if cradle.CallCount() != 1 {
		t.Fatalf("Upsert calls: got %d want 1", cradle.CallCount())
	}

	calls := cradle.Calls()
	call := calls[0]

	if call.Address != config.CradleAddr {
		t.Fatalf("Upsert address: got %q want %q", call.Address, config.CradleAddr)
	}

	if _, err := ulid.Parse(call.ID); err != nil {
		t.Fatalf("Upsert id not ULID: %v", err)
	}

	if config.CradleServerID != call.ID {
		t.Fatalf("config CradleServerID not updated: got %q want %q", config.CradleServerID, call.ID)
	}

	if call.Stamp.IsZero() {
		t.Fatal("Upsert timestamp not set")
	}
	if !call.Stamp.Equal(call.Stamp.UTC()) {
		t.Fatalf("Upsert timestamp not UTC: %s", call.Stamp)
	}
}
