package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	t.Parallel()

	type tc struct {
		name     string
		bucket   string
		objectID string
		setup    func(t *testing.T, objectsRoot string) // optional setup before calling NewWriter
		wantErr  bool
	}

	cases := []tc{
		{
			name:     "creates temp file",
			bucket:   "photos",
			objectID: "obj-456",
			wantErr:  false,
		},
		{
			name:     "rejects existing temp",
			bucket:   "docs",
			objectID: "obj-789",
			setup: func(t *testing.T, objectsRoot string) {
				// Pre-create bucket directory and temp file
				bucketDir := filepath.Join(objectsRoot, "docs")
				if err := os.MkdirAll(bucketDir, 0755); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				tempPath := filepath.Join(bucketDir, ".obj-789.part")
				if err := os.WriteFile(tempPath, []byte("existing"), 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			objectsRoot := t.TempDir()

			if c.setup != nil {
				c.setup(t, objectsRoot)
			}

			_, err := NewWriter(objectsRoot, c.bucket, c.objectID)

			if c.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify bucket directory was created
			if !c.wantErr {
				bucketDir := filepath.Join(objectsRoot, c.bucket)
				if _, err := os.Stat(bucketDir); os.IsNotExist(err) {
					t.Fatalf("bucket directory not created: %s", bucketDir)
				}

				// Verify temp file was created
				expectedTempPath := filepath.Join(bucketDir, fmt.Sprintf(".%s.part", c.objectID))
				info, err := os.Stat(expectedTempPath)
				if err != nil {
					t.Fatalf("temp file not created: %v", err)
				}
				if info.IsDir() {
					t.Fatal("temp path is a directory, expected file")
				}

				// Verify no other files (especially not the final file)
				entries, err := os.ReadDir(bucketDir)
				if err != nil {
					t.Fatalf("failed to read bucket directory: %v", err)
				}
				for _, entry := range entries {
					if !strings.HasPrefix(entry.Name(), ".") {
						t.Fatalf("unexpected non-temp file in bucket directory: %s", entry.Name())
					}
				}
			}
		})
	}
}

func TestWriter_Write(t *testing.T) {
	t.Parallel()

	objectsRoot := t.TempDir()
	bucket := "test-bucket"
	objectID := "test-obj"

	w, err := NewWriter(objectsRoot, bucket, objectID)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}

	n, err := w.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != 11 {
		t.Fatalf("bytes written: got %d, want 11", n)
	}

	// Verify data was written to temp file
	content, err := os.ReadFile(w.TempPath)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	if string(content) != "hello world" {
		t.Fatalf("temp file content: got %q, want %q", string(content), "hello world")
	}
}

func TestWriter_Commit(t *testing.T) {
	t.Parallel()

	objectsRoot := t.TempDir()
	bucket := "test-bucket"
	objectID := "test-obj"

	w, err := NewWriter(objectsRoot, bucket, objectID)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}

	// Write some data
	_, err = w.Write([]byte("committed data"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Commit the file
	err = w.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify final file exists and has correct content
	content, err := os.ReadFile(w.FinalPath)
	if err != nil {
		t.Fatalf("failed to read final file: %v", err)
	}

	if string(content) != "committed data" {
		t.Fatalf("final file content: got %q, want %q", string(content), "committed data")
	}

	// Verify temp file no longer exists
	if _, err := os.Stat(w.TempPath); err == nil {
		t.Fatal("temp file still exists after commit")
	} else if !os.IsNotExist(err) {
		t.Fatalf("unexpected error checking temp file: %v", err)
	}
}

func TestWriter_Abort(t *testing.T) {
	t.Parallel()

	type tc struct {
		name    string
		setup   func(t *testing.T, w *Writer) // optional setup after writer creation
		verify  func(t *testing.T, w *Writer) // verify specific behavior
		wantErr bool
	}

	cases := []tc{
		{
			name: "removes temp file",
			setup: func(t *testing.T, w *Writer) {
				// Write some data to the temp file
				_, err := w.Write([]byte("test data"))
				if err != nil {
					t.Fatalf("Write failed: %v", err)
				}
			},
			verify: func(t *testing.T, w *Writer) {
				// Verify temp file was removed
				if _, err := os.Stat(w.TempPath); err == nil {
					t.Fatal("temp file still exists after abort")
				} else if !os.IsNotExist(err) {
					t.Fatalf("unexpected error checking temp file: %v", err)
				}
			},
			wantErr: false,
		},
		{
			name: "idempotent",
			setup: func(t *testing.T, w *Writer) {
				// Call Abort once
				if err := w.Abort(); err != nil {
					t.Fatalf("first Abort failed: %v", err)
				}
			},
			verify: func(t *testing.T, w *Writer) {
				// Verify second Abort doesn't error
			},
			wantErr: false,
		},
		{
			name: "closes file handle",
			setup: func(t *testing.T, w *Writer) {
				// Write some data
				_, err := w.Write([]byte("data"))
				if err != nil {
					t.Fatalf("Write failed: %v", err)
				}
			},
			verify: func(t *testing.T, w *Writer) {
				// Try to write after abort - should fail because file is closed
				_, err := w.Write([]byte("more data"))
				if err == nil {
					t.Fatal("Write succeeded after Abort, expected error")
				}
			},
			wantErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			objectsRoot := t.TempDir()
			bucket := "test-bucket"
			objectID := "test-obj"

			w, err := NewWriter(objectsRoot, bucket, objectID)
			if err != nil {
				t.Fatalf("NewWriter failed: %v", err)
			}

			if c.setup != nil {
				c.setup(t, w)
			}

			err = w.Abort()

			if c.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if c.verify != nil {
				c.verify(t, w)
			}
		})
	}
}
