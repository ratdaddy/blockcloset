package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// Writer handles writing object data to disk.
type Writer struct {
	File      *os.File
	TempPath  string
	FinalPath string
}

// NewWriter creates a new Writer for the given bucket and object ID.
func NewWriter(objectsRoot, bucket, objectID string) (*Writer, error) {
	bucketDir := filepath.Join(objectsRoot, bucket)
	if err := os.MkdirAll(bucketDir, 0755); err != nil {
		return nil, err
	}

	tempPath := filepath.Join(bucketDir, fmt.Sprintf(".%s.part", objectID))
	finalPath := filepath.Join(bucketDir, objectID)

	f, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		return nil, err
	}

	return &Writer{
		File:      f,
		TempPath:  tempPath,
		FinalPath: finalPath,
	}, nil
}

// Write writes data to the temp file.
func (w *Writer) Write(p []byte) (int, error) {
	return w.File.Write(p)
}

// Commit atomically moves the temp file to the final path.
func (w *Writer) Commit() error {
	// Sync file data to disk
	if err := w.File.Sync(); err != nil {
		return err
	}

	// Close the file
	if err := w.File.Close(); err != nil {
		return err
	}

	// Atomically rename temp file to final path
	if err := os.Rename(w.TempPath, w.FinalPath); err != nil {
		return err
	}

	// Sync parent directory to persist the rename operation
	parentDir := filepath.Dir(w.FinalPath)
	dir, err := os.Open(parentDir)
	if err != nil {
		return err
	}
	defer dir.Close()

	if err := dir.Sync(); err != nil {
		return err
	}

	return nil
}

// Abort closes the file and removes the temp file.
func (w *Writer) Abort() error {
	// Close the file if it's still open (ignore error if already closed)
	if w.File != nil {
		w.File.Close()
	}

	// Remove the temp file if it exists
	if err := os.Remove(w.TempPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
