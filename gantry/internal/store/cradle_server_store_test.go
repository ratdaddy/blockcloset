package store_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	store "github.com/ratdaddy/blockcloset/gantry/internal/store"
)

func TestCradleServerStore_Upsert(t *testing.T) {
	t.Parallel()

	type tc struct {
		name        string
		address     string
		firstID     string
		firstStamp  time.Time
		secondID    string
		secondStamp time.Time
	}

	base := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)

	cases := []tc{
		{
			name:       "creates cradle server record",
			address:    "127.0.0.1:9444",
			firstID:    "cradle-server-id-initial",
			firstStamp: base,
		},
		{
			name:        "updates existing cradle server record",
			address:     "127.0.0.1:9444",
			firstID:     "cradle-server-id-initial",
			firstStamp:  base,
			secondID:    "cradle-server-id-replaced",
			secondStamp: base.Add(10 * time.Minute),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := openIsolatedDB(t)
			s := store.NewCradleServerStore(db)

			rec, err := s.Upsert(ctx, c.firstID, c.address, c.firstStamp)
			if err != nil {
				t.Fatalf("Upsert initial: unexpected error: %v", err)
			}

			wantCreated := c.firstStamp.UTC().Truncate(time.Microsecond)
			wantUpdated := wantCreated

			if c.secondID != "" {
				rec, err = s.Upsert(ctx, c.secondID, c.address, c.secondStamp)
				if err != nil {
					t.Fatalf("Upsert update: unexpected error: %v", err)
				}

				wantUpdated = c.secondStamp.UTC().Truncate(time.Microsecond)

				if rec.ID != c.secondID {
					t.Fatalf("ID mismatch: got %s want %s", rec.ID, c.secondID)
				}
			} else if rec.ID != c.firstID {
				t.Fatalf("ID mismatch: got %s want %s", rec.ID, c.firstID)
			}

			if rec.Address != c.address {
				t.Fatalf("Address mismatch: got %s want %s", rec.Address, c.address)
			}

			if !rec.CreatedAt.Equal(wantCreated) {
				t.Fatalf("CreatedAt mismatch: got %s want %s", rec.CreatedAt, wantCreated)
			}
			if !rec.UpdatedAt.Equal(wantUpdated) {
				t.Fatalf("UpdatedAt mismatch: got %s want %s", rec.UpdatedAt, wantUpdated)
			}

			assertCradleServerRow(t, ctx, db, c.address, rec.ID, wantCreated, wantUpdated)
		})
	}
}

func assertCradleServerRow(t *testing.T, ctx context.Context, db *sql.DB, address string, wantID string, wantCreated time.Time, wantUpdated time.Time) {
	t.Helper()

	var (
		id         string
		storedAddr string
		createdAt  int64
		updatedAt  int64
	)

	query := `SELECT id, address, created_at, updated_at FROM cradle_servers WHERE address = ?`
	if err := db.QueryRowContext(ctx, query, address).Scan(&id, &storedAddr, &createdAt, &updatedAt); err != nil {
		t.Fatalf("fetch cradle server: %v", err)
	}

	if id != wantID {
		t.Fatalf("stored id mismatch: got %s want %s", id, wantID)
	}

	if storedAddr != address {
		t.Fatalf("stored address mismatch: got %s want %s", storedAddr, address)
	}

	if created := time.UnixMicro(createdAt).UTC(); !created.Equal(wantCreated) {
		t.Fatalf("stored created_at mismatch: got %s want %s", created, wantCreated)
	}

	if updated := time.UnixMicro(updatedAt).UTC(); !updated.Equal(wantUpdated) {
		t.Fatalf("stored updated_at mismatch: got %s want %s", updated, wantUpdated)
	}
}
