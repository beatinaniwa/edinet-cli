package extract

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	MaxUncompressedSize = 500 * 1024 * 1024 // 500MB total
	MaxEntrySize        = 100 * 1024 * 1024  // 100MB per entry
	MaxEntryCount       = 10000
)

// ZipEntry represents a file read from a ZIP archive in memory.
type ZipEntry struct {
	Name string
	Data []byte
}

// ExtractedFile represents a file extracted to disk.
type ExtractedFile struct {
	Name string
	Path string
	Size int64
}

// ReadFromZip reads files matching a glob pattern from a ZIP archive in memory.
// Returns entries sorted by name. Does not extract to disk.
func ReadFromZip(zipData []byte, pattern string) ([]ZipEntry, error) {
	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	if len(r.File) > MaxEntryCount {
		return nil, fmt.Errorf("too many entries in zip: %d > %d", len(r.File), MaxEntryCount)
	}

	var entries []ZipEntry
	var totalSize uint64
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		matched, err := path.Match(pattern, f.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}
		if !matched {
			continue
		}

		if f.UncompressedSize64 > uint64(MaxEntrySize) {
			return nil, fmt.Errorf("entry %q exceeds per-entry size limit (%d > %d)", f.Name, f.UncompressedSize64, MaxEntrySize)
		}
		totalSize += f.UncompressedSize64
		if totalSize > uint64(MaxUncompressedSize) {
			return nil, fmt.Errorf("total uncompressed size exceeds limit (%d > %d)", totalSize, MaxUncompressedSize)
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open entry %q: %w", f.Name, err)
		}
		data, err := io.ReadAll(io.LimitReader(rc, int64(MaxEntrySize)+1))
		_ = rc.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read entry %q: %w", f.Name, err)
		}
		if len(data) > MaxEntrySize {
			return nil, fmt.Errorf("entry %q exceeds per-entry size limit", f.Name)
		}

		entries = append(entries, ZipEntry{Name: f.Name, Data: data})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}

// SafeExtract extracts all files from a ZIP archive to the target directory.
// It rejects zip-slip path traversal, oversized entries, and excessive entry counts.
func SafeExtract(zipData []byte, targetDir string) ([]ExtractedFile, error) {
	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	if len(r.File) > MaxEntryCount {
		return nil, fmt.Errorf("too many entries in zip: %d > %d", len(r.File), MaxEntryCount)
	}

	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target dir: %w", err)
	}

	var files []ExtractedFile
	var totalSize uint64

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// Zip-slip prevention
		destPath := filepath.Join(absTarget, filepath.Clean(f.Name))
		if !strings.HasPrefix(destPath, absTarget+string(os.PathSeparator)) {
			return nil, fmt.Errorf("zip slip detected: %q resolves outside target directory", f.Name)
		}

		if f.UncompressedSize64 > uint64(MaxEntrySize) {
			return nil, fmt.Errorf("entry %q exceeds per-entry size limit (%d > %d)", f.Name, f.UncompressedSize64, MaxEntrySize)
		}
		totalSize += f.UncompressedSize64
		if totalSize > uint64(MaxUncompressedSize) {
			return nil, fmt.Errorf("total uncompressed size exceeds limit (%d > %d)", totalSize, MaxUncompressedSize)
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return nil, fmt.Errorf("failed to create directory for %q: %w", f.Name, err)
		}

		// Extract file atomically (temp + rename)
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open entry %q: %w", f.Name, err)
		}

		tmpFile, err := os.CreateTemp(filepath.Dir(destPath), ".extract-*")
		if err != nil {
			_ = rc.Close()
			return nil, fmt.Errorf("failed to create temp file for %q: %w", f.Name, err)
		}

		n, err := io.Copy(tmpFile, io.LimitReader(rc, int64(MaxEntrySize)+1))
		_ = rc.Close()
		_ = tmpFile.Close()
		if err != nil {
			_ = os.Remove(tmpFile.Name())
			return nil, fmt.Errorf("failed to write %q: %w", f.Name, err)
		}
		if n > int64(MaxEntrySize) {
			_ = os.Remove(tmpFile.Name())
			return nil, fmt.Errorf("entry %q exceeds per-entry size limit", f.Name)
		}

		if err := os.Rename(tmpFile.Name(), destPath); err != nil {
			_ = os.Remove(tmpFile.Name())
			return nil, fmt.Errorf("failed to rename temp file for %q: %w", f.Name, err)
		}

		files = append(files, ExtractedFile{
			Name: f.Name,
			Path: destPath,
			Size: n,
		})
	}

	return files, nil
}
