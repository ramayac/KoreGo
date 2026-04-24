package readlink

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	os.WriteFile(target, []byte("x"), 0644)
	link := filepath.Join(dir, "link")
	os.Symlink(target, link)

	result, err := Run(link, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Target != target {
		t.Errorf("got %q, want %q", result.Target, target)
	}
}

func TestRunCanonicalize(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real.txt")
	os.WriteFile(target, []byte("x"), 0644)
	link := filepath.Join(dir, "link")
	os.Symlink(target, link)

	result, err := Run(link, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Target == "" {
		t.Error("expected non-empty canonical target")
	}
}

func TestRunNotSymlink(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "plain.txt")
	os.WriteFile(f, []byte("x"), 0644)
	_, err := Run(f, false)
	if err == nil {
		t.Error("expected error for non-symlink")
	}
}
