package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileStore_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	data := []byte("hello world")
	if err := store.Set("test/key.json", data); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := store.Get("test/key.json", time.Hour)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("Get() = %q, want %q", string(got), "hello world")
	}
}

func TestFileStore_Get_Miss(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewFileStore(dir)

	_, err := store.Get("nonexistent/key", time.Hour)
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("Get() error = %v, want ErrCacheMiss", err)
	}
}

func TestFileStore_Get_Expired(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewFileStore(dir)

	if err := store.Set("expired/key", []byte("old data")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Set the file's mtime to 2 hours ago
	path := filepath.Join(dir, "expired", "key")
	past := time.Now().Add(-2 * time.Hour)
	_ = os.Chtimes(path, past, past)

	_, err := store.Get("expired/key", time.Hour)
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("Get() error = %v, want ErrCacheMiss (expired)", err)
	}
}

func TestFileStore_Get_NotExpired(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewFileStore(dir)

	if err := store.Set("fresh/key", []byte("fresh data")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := store.Get("fresh/key", time.Hour)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(got) != "fresh data" {
		t.Errorf("Get() = %q, want %q", string(got), "fresh data")
	}
}

func TestFileStore_Set_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewFileStore(dir)

	if err := store.Set("deep/nested/path/file.json", []byte("data")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	path := filepath.Join(dir, "deep", "nested", "path", "file.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at expected path: %v", err)
	}
}

func TestFileStore_Set_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewFileStore(dir)

	// Write initial data
	if err := store.Set("atomic/test", []byte("original")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Overwrite with new data
	if err := store.Set("atomic/test", []byte("updated")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := store.Get("atomic/test", time.Hour)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(got) != "updated" {
		t.Errorf("Get() = %q, want %q", string(got), "updated")
	}

	// No temp files should be left behind
	entries, _ := os.ReadDir(filepath.Join(dir, "atomic"))
	for _, e := range entries {
		if e.Name() != "test" {
			t.Errorf("unexpected file in cache dir: %q (possible temp file leak)", e.Name())
		}
	}
}

func TestFileStore_Clear(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewFileStore(dir)

	_ = store.Set("a/file1", []byte("1"))
	_ = store.Set("b/file2", []byte("2"))

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	_, err := store.Get("a/file1", time.Hour)
	if !errors.Is(err, ErrCacheMiss) {
		t.Error("Get() should return miss after Clear()")
	}
}

func TestFileStore_ClearKey(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewFileStore(dir)

	_ = store.Set("keep/this", []byte("keep"))
	_ = store.Set("delete/this", []byte("delete"))

	if err := store.ClearKey("delete/this"); err != nil {
		t.Fatalf("ClearKey() error = %v", err)
	}

	_, err := store.Get("delete/this", time.Hour)
	if !errors.Is(err, ErrCacheMiss) {
		t.Error("deleted key should be a miss")
	}

	got, err := store.Get("keep/this", time.Hour)
	if err != nil {
		t.Fatalf("kept key should still exist: %v", err)
	}
	if string(got) != "keep" {
		t.Errorf("kept key = %q, want %q", string(got), "keep")
	}
}

func TestFileStore_Get_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewFileStore(dir)

	// Create a directory where a file is expected (corrupt state)
	path := filepath.Join(dir, "corrupt", "key")
	_ = os.MkdirAll(path, 0o755) // "key" is a dir instead of file

	_, err := store.Get("corrupt/key", time.Hour)
	// Should return an error (not ErrCacheMiss — this is a read failure)
	if err == nil {
		t.Fatal("Get() should return error for corrupt cache entry")
	}
	if errors.Is(err, ErrCacheMiss) {
		t.Error("corrupt entry should not be a simple cache miss")
	}
}
