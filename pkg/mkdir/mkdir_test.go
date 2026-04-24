package mkdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunBasic(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "newdir")
	result, err := Run([]string{target}, false, 0755)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Created) != 1 || result.Created[0] != target {
		t.Errorf("unexpected Created: %v", result.Created)
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}

func TestRunParents(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b", "c")
	_, err := Run([]string{target}, true, 0755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Error("parent directories not created")
	}
}

func TestRunExistsFails(t *testing.T) {
	dir := t.TempDir()
	_, err := Run([]string{dir}, false, 0755)
	if err == nil {
		t.Error("expected error for existing directory without -p")
	}
}

func TestRunMode(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "modedir")
	_, err := Run([]string{target}, false, 0700)
	if err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(target)
	if info.Mode().Perm() != 0700 {
		t.Errorf("expected mode 0700, got %v", info.Mode().Perm())
	}
}
