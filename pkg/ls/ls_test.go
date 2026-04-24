package ls

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunBasicDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("h"), 0644)

	results, err := Run([]string{dir}, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Should only show a.txt (hidden excluded by default).
	found := false
	for _, f := range results[0].Files {
		if f.Name == "a.txt" {
			found = true
		}
		if f.Name == ".hidden" {
			t.Error("-a not set: .hidden should be excluded")
		}
	}
	if !found {
		t.Error("a.txt not found in result")
	}
}

func TestRunShowAll(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".dotfile"), []byte("x"), 0644)

	results, err := Run([]string{dir}, true, false, false)
	if err != nil {
		t.Fatal(err)
	}
	foundDot := false
	for _, f := range results[0].Files {
		if f.Name == "." || f.Name == ".dotfile" {
			foundDot = true
		}
	}
	if !foundDot {
		t.Error("-a: expected dotfiles and . to be present")
	}
}

func TestRunAlmostAll(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".dotfile"), []byte("x"), 0644)

	results, err := Run([]string{dir}, false, true, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range results[0].Files {
		if f.Name == "." || f.Name == ".." {
			t.Errorf("-A: %q should be excluded", f.Name)
		}
	}
}

func TestRunNonExistent(t *testing.T) {
	_, err := Run([]string{"/this/path/does/not/exist"}, false, false, false)
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestRunSingleFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("data"), 0644)

	results, err := Run([]string{f}, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || len(results[0].Files) != 1 {
		t.Errorf("expected 1 result with 1 file, got %+v", results)
	}
}

func TestHumanSize(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
	}
	for _, c := range cases {
		got := humanSize(c.n)
		if got != c.want {
			t.Errorf("humanSize(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestSortFiles(t *testing.T) {
	files := []FileInfo{
		{Name: "z"}, {Name: "a"}, {Name: "m"},
	}
	sorted := sortFiles(files, false, false, false)
	if sorted[0].Name != "a" || sorted[2].Name != "z" {
		t.Errorf("sort by name failed: %v", sorted)
	}
}

func TestSortFilesReverse(t *testing.T) {
	files := []FileInfo{
		{Name: "a"}, {Name: "z"},
	}
	sorted := sortFiles(files, false, false, true)
	if sorted[0].Name != "z" {
		t.Errorf("reverse sort failed: %v", sorted)
	}
}
