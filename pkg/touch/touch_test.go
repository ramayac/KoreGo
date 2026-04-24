package touch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunCreate(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "newfile.txt")
	ts := time.Now()
	result, err := Run([]string{f}, ts)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Touched) != 1 {
		t.Fatalf("expected 1 touched, got %v", result.Touched)
	}
	if _, err := os.Stat(f); os.IsNotExist(err) {
		t.Error("file was not created by touch")
	}
}

func TestRunUpdateMtime(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "existing.txt")
	os.WriteFile(f, []byte("data"), 0644)

	past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := Run([]string{f}, past)
	if err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(f)
	// Allow ±2s tolerance for filesystem mtime resolution.
	delta := info.ModTime().Unix() - past.Unix()
	if delta < -2 || delta > 2 {
		t.Errorf("mtime not updated: got %v, want ~%v", info.ModTime(), past)
	}
}
