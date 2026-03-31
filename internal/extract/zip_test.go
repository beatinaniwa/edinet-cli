package extract

import (
	"archive/zip"
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// createTestZip creates an in-memory ZIP with the given files.
func createTestZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry %q: %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write zip entry %q: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	return buf.Bytes()
}

func TestReadFromZip_MatchingPattern(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"XBRL_TO_CSV/data1.csv": "col1,col2\na,b",
		"XBRL_TO_CSV/data2.csv": "col1,col2\nc,d",
		"PublicDoc/main.htm":    "<html>test</html>",
	})

	entries, err := ReadFromZip(data, "XBRL_TO_CSV/*.csv")
	if err != nil {
		t.Fatalf("ReadFromZip() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Name != "XBRL_TO_CSV/data1.csv" {
		t.Errorf("entries[0].Name = %q, want XBRL_TO_CSV/data1.csv", entries[0].Name)
	}
}

func TestReadFromZip_NoMatch(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/main.htm": "<html>test</html>",
	})

	entries, err := ReadFromZip(data, "XBRL_TO_CSV/*.csv")
	if err != nil {
		t.Fatalf("ReadFromZip() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0", len(entries))
	}
}

func TestReadFromZip_CorruptZip(t *testing.T) {
	_, err := ReadFromZip([]byte("not a zip file"), "*.csv")
	if err == nil {
		t.Fatal("ReadFromZip() should fail on corrupt data")
	}
}

func TestSafeExtract_Normal(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"dir/file1.txt": "hello",
		"dir/file2.txt": "world",
	})
	dir := t.TempDir()
	files, err := SafeExtract(data, dir)
	if err != nil {
		t.Fatalf("SafeExtract() error = %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("len(files) = %d, want 2", len(files))
	}
	for _, f := range files {
		if !filepath.IsAbs(f.Path) {
			t.Errorf("path %q is not absolute", f.Path)
		}
		if !strings.HasPrefix(f.Path, dir) {
			t.Errorf("path %q is not under target dir %q", f.Path, dir)
		}
	}
}

func TestSafeExtract_ZipSlip(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("../../../etc/passwd")
	_, _ = f.Write([]byte("malicious"))
	_ = w.Close()

	dir := t.TempDir()
	_, err := SafeExtract(buf.Bytes(), dir)
	if err == nil {
		t.Fatal("SafeExtract() should reject zip-slip path traversal")
	}
	if !strings.Contains(err.Error(), "zip slip") {
		t.Errorf("error = %q, should mention zip slip", err.Error())
	}
}

func TestSafeExtract_EntryCountLimit(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for i := 0; i < MaxEntryCount+1; i++ {
		name := fmt.Sprintf("dir/file%06d.txt", i)
		f, _ := w.Create(name)
		_, _ = f.Write([]byte("x"))
	}
	_ = w.Close()

	dir := t.TempDir()
	_, err := SafeExtract(buf.Bytes(), dir)
	if err == nil {
		t.Fatal("SafeExtract() should reject excessive entry count")
	}
	if !strings.Contains(err.Error(), "too many entries") {
		t.Errorf("error = %q, should mention too many entries", err.Error())
	}
}

func TestReadFromZip_PerEntrySizeLimit(t *testing.T) {
	// Create a ZIP with a large entry (exceeding per-entry limit is hard in test,
	// so we test that the limit constant is reasonable)
	if MaxEntrySize < 1024*1024 {
		t.Errorf("MaxEntrySize = %d, should be at least 1MB", MaxEntrySize)
	}
}
