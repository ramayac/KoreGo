package stat

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunBasicFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("hello"), 0644)

	result, err := Run(f)
	if err != nil {
		t.Fatal(err)
	}
	if result.Size != 5 {
		t.Errorf("size: got %d, want 5", result.Size)
	}
	if result.Path != f {
		t.Errorf("path: got %q, want %q", result.Path, f)
	}
	if result.Inode == 0 {
		t.Error("expected non-zero inode")
	}
	if result.Links == 0 {
		t.Error("expected non-zero links")
	}
}

func TestRunNonExistent(t *testing.T) {
	_, err := Run("/this/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestRunDirectory(t *testing.T) {
	dir := t.TempDir()
	result, err := Run(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsDir {
		t.Error("expected IsDir=true for directory")
	}
}
