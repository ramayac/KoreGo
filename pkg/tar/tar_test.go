package tar

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTarCreateExtract(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file inside a test directory
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	testFile := filepath.Join(srcDir, "test.txt")
	content := "hello tar"
	os.WriteFile(testFile, []byte(content), 0644)

	archiveFile := filepath.Join(tmpDir, "archive.tar")

	// Create archive
	var buf bytes.Buffer
	code := run([]string{"-c", "-f", archiveFile, "-C", tmpDir, "src"}, &buf)
	if code != 0 {
		t.Fatalf("create exit code %d", code)
	}

	if _, err := os.Stat(archiveFile); os.IsNotExist(err) {
		t.Fatalf("archive file not created")
	}

	// Extract archive into a new directory
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(destDir, 0755)

	var buf2 bytes.Buffer
	code = run([]string{"-x", "-f", archiveFile, "-C", destDir}, &buf2)
	if code != 0 {
		t.Fatalf("extract exit code %d", code)
	}

	// Check if file exists in the extracted location
	// The path inside the tar will be absolute since we passed an absolute path (srcDir)
	// So it gets extracted to destDir/tmp/...
	// This is standard tar behavior for absolute paths if not stripped.
	// Wait, filepath.Walk uses absolute paths if given absolute paths.
	// Let's check where it got extracted.
	// Actually, just find the file.
	found := false
	filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Name() == "test.txt" {
			data, _ := os.ReadFile(path)
			if string(data) == content {
				found = true
			}
		}
		return nil
	})

	if !found {
		t.Errorf("extracted file not found or content mismatch")
	}
}

func TestTarList(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	testFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(testFile, []byte("hello tar"), 0644)

	archiveFile := filepath.Join(tmpDir, "archive.tar")
	var buf bytes.Buffer
	code := run([]string{"-c", "-f", archiveFile, "-C", tmpDir, "src"}, &buf)
	if code != 0 {
		t.Fatalf("create exit code %d", code)
	}

	var buf2 bytes.Buffer
	code = run([]string{"-t", "-f", archiveFile}, &buf2)
	if code != 0 {
		t.Fatalf("list exit code %d", code)
	}

	out := buf2.String()
	if !strings.Contains(out, "test.txt") {
		t.Errorf("list output missing filename: %s", out)
	}
}

func TestTarGzip(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	testFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(testFile, []byte(strings.Repeat("a", 1000)), 0644)

	archiveFile := filepath.Join(tmpDir, "archive.tar.gz")
	
	// Create with -z
	var buf bytes.Buffer
	code := run([]string{"-c", "-z", "-f", archiveFile, "-C", tmpDir, "src"}, &buf)
	if code != 0 {
		t.Fatalf("create exit code %d", code)
	}

	// Verify it's gzipped by checking magic number
	f, _ := os.Open(archiveFile)
	magic := make([]byte, 2)
	f.Read(magic)
	f.Close()
	if magic[0] != 0x1f || magic[1] != 0x8b {
		t.Errorf("file is not gzipped")
	}

	// List with -z
	var buf2 bytes.Buffer
	code = run([]string{"-t", "-z", "-f", archiveFile}, &buf2)
	if code != 0 {
		t.Fatalf("list exit code %d", code)
	}
}

func TestTarJSON(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	testFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(testFile, []byte("hello tar"), 0644)

	archiveFile := filepath.Join(tmpDir, "archive.tar")
	var buf bytes.Buffer
	code := run([]string{"-j", "-c", "-f", archiveFile, "-C", tmpDir, "src"}, &buf)
	if code != 0 {
		t.Fatalf("create exit code %d", code)
	}

	var env map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	data := env["data"].([]interface{})
	if len(data) == 0 {
		t.Errorf("expected files in json output")
	}
}

func TestTarMissingArgs(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"-c"}, &buf)
	if code != 1 {
		t.Errorf("should fail without -f")
	}

	var buf2 bytes.Buffer
	code = run([]string{"-f", "test.tar"}, &buf2)
	if code != 1 {
		t.Errorf("should fail without mode (-c, -x, -t)")
	}

	var buf3 bytes.Buffer
	code = run([]string{"-c", "-x", "-f", "test.tar"}, &buf3)
	if code != 1 {
		t.Errorf("should fail with multiple modes")
	}
}
