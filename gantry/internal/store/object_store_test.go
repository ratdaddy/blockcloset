package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	store "github.com/ratdaddy/blockcloset/gantry/internal/store"
)

func TestObjectStore_Create(t *testing.T) {
	t.Parallel()

	type tc struct {
		name             string
		id               string
		bucketID         string
		key              string
		sizeExpected     int64
		cradleServerID   string
		skipBucket       bool
		skipCradleServer bool
		setup            func(context.Context, *testing.T, *store.ObjectStore, time.Time, *sql.DB, string, string)
		wantErr          bool
	}

	base := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)

	cases := []tc{
		{
			name:           "creates object with PENDING state",
			id:             "object-id-1",
			bucketID:       "bucket-id-1",
			key:            "test-key.txt",
			sizeExpected:   1024,
			cradleServerID: "cradle-id-1",
		},
		{
			name:             "invalid foreign keys return error",
			id:               "object-id-2",
			bucketID:         "nonexistent-bucket-id",
			key:              "test-key.txt",
			sizeExpected:     1024,
			cradleServerID:   "nonexistent-cradle-id",
			skipBucket:       true,
			skipCradleServer: true,
			wantErr:          true,
		},
		{
			name:           "duplicate bucket+key allowed (multiple PENDING records)",
			id:             "object-id-3",
			bucketID:       "bucket-id-1",
			key:            "duplicate-key.txt",
			sizeExpected:   2048,
			cradleServerID: "cradle-id-1",
			setup: func(ctx context.Context, t *testing.T, s *store.ObjectStore, createdAt time.Time, db *sql.DB, bucketID, cradleServerID string) {
				// Create first object with same bucket+key
				_, err := (*s).CreatePending(ctx, "object-id-first", bucketID, "duplicate-key.txt", 1024, cradleServerID, createdAt)
				if err != nil {
					t.Fatalf("setup: create first object: %v", err)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := openIsolatedDB(t)
			s := store.NewObjectStore(db)

			createdAt := base

			// Setup prerequisites (bucket and cradle server)
			setupPrerequisites(ctx, t, db, c.bucketID, c.cradleServerID, createdAt, c.skipBucket, c.skipCradleServer)

			// Additional test-specific setup
			if c.setup != nil {
				c.setup(ctx, t, &s, createdAt, db, c.bucketID, c.cradleServerID)
			}

			rec, err := s.CreatePending(ctx, c.id, c.bucketID, c.key, c.sizeExpected, c.cradleServerID, createdAt)

			if c.wantErr {
				if err == nil {
					t.Fatal("CreatePending: expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("CreatePending: unexpected error: %v", err)
			}

			assertObjectRecord(t, ctx, db, rec, c.id, c.bucketID, c.key, c.sizeExpected, c.cradleServerID, createdAt)
		})
	}
}

func TestObjectStore_CommitWithReplace(t *testing.T) {
	t.Parallel()

	type tc struct {
		name            string
		sizeActual      int64
		lastModifiedMs  int64
		skipSetup       bool
		preCommit       bool
		replaceExisting bool
		wantErr         error
	}

	cases := []tc{
		{
			name:           "transitions PENDING to COMMITTED",
			sizeActual:     987,
			lastModifiedMs: 1735689600000,
		},
		{
			name:           "object not found returns error",
			sizeActual:     100,
			lastModifiedMs: 12345,
			skipSetup:      true,
			wantErr:        store.ErrObjectNotPending,
		},
		{
			name:           "already committed object returns error",
			sizeActual:     987,
			lastModifiedMs: 1735689600000,
			skipSetup:      true,
			preCommit:      true,
			wantErr:        store.ErrObjectNotPending,
		},
		{
			name:            "replaces previous COMMITTED version",
			sizeActual:      512,
			lastModifiedMs:  1735689600000,
			replaceExisting: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := openIsolatedDB(t)
			s := store.NewObjectStore(db)

			createdAt := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
			bucketID := "bucket-id-commit"
			cradleServerID := "cradle-id-commit"
			objectID := "object-id-commit"

			setupPrerequisites(ctx, t, db, bucketID, cradleServerID, createdAt, false, false)

			priorObjectID := "object-id-prior"
			if c.replaceExisting {
				insertCommittedObject(ctx, t, db, priorObjectID, bucketID, "photos/sunset.jpg", cradleServerID, createdAt)
			}

			if !c.skipSetup {
				_, err := s.CreatePending(ctx, objectID, bucketID, "photos/sunset.jpg", 1024, cradleServerID, createdAt)
				if err != nil {
					t.Fatalf("setup CreatePending: %v", err)
				}
			}

			if c.preCommit {
				insertCommittedObject(ctx, t, db, objectID, bucketID, "photos/sunset.jpg", cradleServerID, createdAt)
			}

			updatedAt := time.Now()
			err := s.CommitWithReplace(ctx, objectID, c.sizeActual, c.lastModifiedMs, updatedAt)

			if c.wantErr != nil {
				if err == nil {
					t.Fatalf("Commit: expected error %v, got nil", c.wantErr)
				}
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("Commit error: got %v, want %v", err, c.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Commit: unexpected error: %v", err)
			}

			var (
				storedState        string
				storedSizeActual   sql.NullInt64
				storedLastModified sql.NullInt64
				storedUpdatedAt    int64
			)

			const q = `SELECT state, size_actual, last_modified, updated_at FROM objects WHERE object_id = ?`
			if err := db.QueryRowContext(ctx, q, objectID).Scan(&storedState, &storedSizeActual, &storedLastModified, &storedUpdatedAt); err != nil {
				t.Fatalf("query committed object: %v", err)
			}

			if storedState != "COMMITTED" {
				t.Errorf("state: got %q, want %q", storedState, "COMMITTED")
			}
			if !storedSizeActual.Valid || storedSizeActual.Int64 != c.sizeActual {
				t.Errorf("size_actual: got %v, want %d", storedSizeActual, c.sizeActual)
			}
			if !storedLastModified.Valid || storedLastModified.Int64 != c.lastModifiedMs {
				t.Errorf("last_modified: got %v, want %d", storedLastModified, c.lastModifiedMs)
			}

			wantUpdatedAt := updatedAt.UTC().Truncate(time.Microsecond)
			gotUpdatedAt := time.UnixMicro(storedUpdatedAt).UTC()
			if !gotUpdatedAt.Equal(wantUpdatedAt) {
				t.Errorf("updated_at: got %s, want %s", gotUpdatedAt, wantUpdatedAt)
			}

			if c.replaceExisting {
				var priorState string
				var priorUpdatedAt int64
				if err := db.QueryRowContext(ctx, `SELECT state, updated_at FROM objects WHERE object_id = ?`, priorObjectID).Scan(&priorState, &priorUpdatedAt); err != nil {
					t.Fatalf("query prior object: %v", err)
				}
				if priorState != "REPLACED" {
					t.Errorf("prior object state: got %q, want %q", priorState, "REPLACED")
				}
				wantPriorUpdatedAt := updatedAt.UTC().Truncate(time.Microsecond)
				gotPriorUpdatedAt := time.UnixMicro(priorUpdatedAt).UTC()
				if !gotPriorUpdatedAt.Equal(wantPriorUpdatedAt) {
					t.Errorf("prior object updated_at: got %s, want %s", gotPriorUpdatedAt, wantPriorUpdatedAt)
				}

				var committedCount int
				if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM objects WHERE bucket_id = ? AND key = ? AND state = 'COMMITTED'`, bucketID, "photos/sunset.jpg").Scan(&committedCount); err != nil {
					t.Fatalf("count committed rows: %v", err)
				}
				if committedCount != 1 {
					t.Errorf("committed rows for bucket+key: got %d, want 1", committedCount)
				}
			}
		})
	}
}

func insertCommittedObject(ctx context.Context, t *testing.T, db *sql.DB, objectID, bucketID, key, cradleServerID string, createdAt time.Time) {
	t.Helper()
	stamp := createdAt.UTC().Truncate(time.Microsecond).UnixMicro()
	_, err := db.ExecContext(ctx, `
		INSERT INTO objects (object_id, bucket_id, key, state, size_expected, size_actual, last_modified, cradle_server_id, created_at, updated_at)
		VALUES (?, ?, ?, 'COMMITTED', 1024, 1024, ?, ?, ?, ?)
	`, objectID, bucketID, key, stamp, cradleServerID, stamp, stamp)
	if err != nil {
		t.Fatalf("insertCommittedObject: %v", err)
	}
}

func setupPrerequisites(ctx context.Context, t *testing.T, db *sql.DB, bucketID, cradleServerID string, createdAt time.Time, skipBucket, skipCradleServer bool) {
	t.Helper()

	if !skipBucket {
		buckets := store.NewBucketStore(db)
		_, err := buckets.Create(ctx, bucketID, "test-bucket", createdAt)
		if err != nil {
			t.Fatalf("setup: create bucket: %v", err)
		}
	}

	if !skipCradleServer {
		cradles := store.NewCradleServerStore(db)
		_, err := cradles.Upsert(ctx, cradleServerID, "127.0.0.1:9444", createdAt)
		if err != nil {
			t.Fatalf("setup: upsert cradle server: %v", err)
		}
	}
}

func assertObjectRecord(t *testing.T, ctx context.Context, db *sql.DB, rec store.ObjectRecord, wantID, wantBucketID, wantKey string, wantSize int64, wantCradleServerID string, expectedStamp time.Time) {
	t.Helper()

	wantStamp := expectedStamp.UTC().Truncate(time.Microsecond)

	// Verify returned record fields
	if rec.ID != wantID {
		t.Fatalf("returned ID: got %q, want %q", rec.ID, wantID)
	}
	if rec.BucketID != wantBucketID {
		t.Fatalf("returned BucketID: got %q, want %q", rec.BucketID, wantBucketID)
	}
	if rec.Key != wantKey {
		t.Fatalf("returned Key: got %q, want %q", rec.Key, wantKey)
	}
	if rec.State != "PENDING" {
		t.Fatalf("returned State: got %q, want %q", rec.State, "PENDING")
	}
	if rec.SizeExpected != wantSize {
		t.Fatalf("returned SizeExpected: got %d, want %d", rec.SizeExpected, wantSize)
	}
	if rec.CradleServerID != wantCradleServerID {
		t.Fatalf("returned CradleServerID: got %q, want %q", rec.CradleServerID, wantCradleServerID)
	}
	if !rec.CreatedAt.Equal(wantStamp) {
		t.Fatalf("returned CreatedAt: got %s, want %s", rec.CreatedAt, wantStamp)
	}
	if !rec.UpdatedAt.Equal(wantStamp) {
		t.Fatalf("returned UpdatedAt: got %s, want %s", rec.UpdatedAt, wantStamp)
	}

	// Query database to verify record was actually written
	var (
		storedID             string
		storedBucketID       string
		storedKey            string
		storedState          string
		storedSizeExpected   int64
		storedCradleServerID string
		storedCreatedAt      int64
		storedUpdatedAt      int64
	)

	query := `SELECT object_id, bucket_id, key, state, size_expected, cradle_server_id, created_at, updated_at FROM objects WHERE object_id = ?`
	if err := db.QueryRowContext(ctx, query, rec.ID).Scan(&storedID, &storedBucketID, &storedKey, &storedState, &storedSizeExpected, &storedCradleServerID, &storedCreatedAt, &storedUpdatedAt); err != nil {
		t.Fatalf("fetch stored object: %v", err)
	}

	if storedID != rec.ID {
		t.Fatalf("stored object_id: got %q, want %q", storedID, rec.ID)
	}
	if storedBucketID != rec.BucketID {
		t.Fatalf("stored bucket_id: got %q, want %q", storedBucketID, rec.BucketID)
	}
	if storedKey != rec.Key {
		t.Fatalf("stored key: got %q, want %q", storedKey, rec.Key)
	}
	if storedState != "PENDING" {
		t.Fatalf("stored state: got %q, want %q", storedState, "PENDING")
	}
	if storedSizeExpected != rec.SizeExpected {
		t.Fatalf("stored size_expected: got %d, want %d", storedSizeExpected, rec.SizeExpected)
	}
	if storedCradleServerID != rec.CradleServerID {
		t.Fatalf("stored cradle_server_id: got %q, want %q", storedCradleServerID, rec.CradleServerID)
	}

	createdAt := time.UnixMicro(storedCreatedAt).UTC()
	updatedAt := time.UnixMicro(storedUpdatedAt).UTC()

	if !createdAt.Equal(wantStamp) {
		t.Fatalf("stored created_at: got %s, want %s", createdAt, wantStamp)
	}
	if !updatedAt.Equal(wantStamp) {
		t.Fatalf("stored updated_at: got %s, want %s", updatedAt, wantStamp)
	}
}
