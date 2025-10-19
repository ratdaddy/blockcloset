package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/ratdaddy/blockcloset/gantry/internal/config"
	"github.com/ratdaddy/blockcloset/gantry/internal/store"
)

func Init(ctx context.Context, st store.Store) error {
	cradleId := store.NewID()
	cradleAddr := config.CradleAddr

	if _, err := st.CradleServers().Upsert(ctx, cradleId, cradleAddr, time.Now()); err != nil {
		return fmt.Errorf("failed to bootstrap cradle address: %w", err)
	}

	config.CradleServerID = cradleId
	return nil
}
