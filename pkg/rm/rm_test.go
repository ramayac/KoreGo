package rm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunBasic(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "todelete.txt")
	os.WriteFile(f, []byte("x"), 0644)

	result, err := Run([]string{f}, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 1 {
		t.Errorf("expected 1 removed, got %v", result.Removed)
	}
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error("file still exists after rm")
	}
}

func TestRunRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(sub, "f"), []byte("x"), 0644)

	_, err := Run([]string{sub}, true, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sub); !os.IsNotExist(err) {
		t.Error("directory still exists after rm -r")
	}
}

func TestRunRefusesRoot(t *testing.T) {
	_, err := Run([]string{"/"}, true, false, false)
	if err == nil {
		t.Error("expected error when trying to rm /")
	}
}

func TestIsSafeToRemove(t *testing.T) {
	if isSafeToRemove("/") {
		t.Error("/ should not be safe to remove")
	}
	if !isSafeToRemove("/tmp") {
		t.Error("/tmp should be safe to remove")
	}
}
