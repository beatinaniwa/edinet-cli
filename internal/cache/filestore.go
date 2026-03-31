package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileStore is a file-based Cache implementation.
// Each key maps to a file under baseDir. TTL is checked via file mtime.
// Writes are atomic (temp file + rename).
type FileStore struct {
	baseDir string
}

// NewFileStore creates a new FileStore rooted at baseDir.
func NewFileStore(baseDir string) (*FileStore, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &FileStore{baseDir: baseDir}, nil
}

// Get retrieves cached data for the given key.
// Returns ErrCacheMiss if the key does not exist or has expired.
// Returns other errors for read failures (corruption, permissions).
func (s *FileStore) Get(key string, maxAge time.Duration) ([]byte, error) {
	path := filepath.Join(s.baseDir, filepath.FromSlash(key))

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("cache stat error for %q: %w", key, err)
	}

	// Check if it's a directory (corrupt state)
	if info.IsDir() {
		return nil, fmt.Errorf("cache corruption: %q is a directory", key)
	}

	// Check TTL using mtime
	if maxAge > 0 && time.Since(info.ModTime()) > maxAge {
		return nil, ErrCacheMiss
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cache read error for %q: %w", key, err)
	}

	return data, nil
}

// Set stores data for the given key atomically (temp file + rename).
func (s *FileStore) Set(key string, data []byte) error {
	path := filepath.Join(s.baseDir, filepath.FromSlash(key))
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory for %q: %w", key, err)
	}

	// Write to temp file first
	tmp, err := os.CreateTemp(dir, ".cache-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file for %q: %w", key, err)
	}

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("failed to write cache for %q: %w", key, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("failed to close cache temp file for %q: %w", key, err)
	}

	// Atomic rename
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("failed to rename cache file for %q: %w", key, err)
	}

	return nil
}

// Clear removes all cached data.
func (s *FileStore) Clear() error {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(s.baseDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

// ClearKey removes a single cached key.
func (s *FileStore) ClearKey(key string) error {
	path := filepath.Join(s.baseDir, filepath.FromSlash(key))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
