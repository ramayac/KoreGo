package gzip

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGzipGunzipCycle(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	content := "hello world gzip test"
	os.WriteFile(file, []byte(content), 0644)

	var buf bytes.Buffer
	code := runGzip([]string{file}, &buf)
	if code != 0 {
		t.Fatalf("gzip exit code %d", code)
	}

	if _, err := os.Stat(file); err == nil {
		t.Errorf("original file should be deleted")
	}

	gzFile := file + ".gz"
	if _, err := os.Stat(gzFile); os.IsNotExist(err) {
		t.Fatalf("gz file not created")
	}

	code = runGunzip([]string{gzFile}, &buf)
	if code != 0 {
		t.Fatalf("gunzip exit code %d", code)
	}

	if _, err := os.Stat(gzFile); err == nil {
		t.Errorf("gz file should be deleted")
	}

	unpacked, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("unpacked file read err: %v", err)
	}

	if string(unpacked) != content {
		t.Errorf("content mismatch: got %q, want %q", unpacked, content)
	}
}

func TestGzipKeep(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("hello"), 0644)

	var buf bytes.Buffer
	code := runGzip([]string{"-k", file}, &buf)
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Errorf("original file should be kept")
	}
	if _, err := os.Stat(file + ".gz"); os.IsNotExist(err) {
		t.Errorf("gz file should be created")
	}
}

func TestGzipForce(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("hello"), 0644)
	os.WriteFile(file+".gz", []byte("existing"), 0644)

	var buf bytes.Buffer
	code := runGzip([]string{file}, &buf)
	if code != 1 {
		t.Errorf("should fail without force")
	}

	code = runGzip([]string{"-f", file}, &buf)
	if code != 0 {
		t.Errorf("should succeed with force")
	}
}

func TestGzipStdout(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("hello stdout"), 0644)

	var buf bytes.Buffer
	code := runGzip([]string{"-c", file}, &buf)
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Errorf("original file should be kept with -c")
	}
	if _, err := os.Stat(file + ".gz"); err == nil {
		t.Errorf("gz file should not be created with -c")
	}

	if buf.Len() == 0 {
		t.Errorf("stdout should contain compressed data")
	}
}

func TestGzipJSON(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte(strings.Repeat("a", 100)), 0644)

	var buf bytes.Buffer
	code := runGzip([]string{"-j", file}, &buf)
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}

	var env map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	data := env["data"].([]interface{})
	stat := data[0].(map[string]interface{})
	
	if stat["file"] != file {
		t.Errorf("wrong file name in json")
	}
	if stat["originalSize"].(float64) != 100 {
		t.Errorf("wrong original size: %v", stat["originalSize"])
	}
	if stat["newSize"].(float64) == 0 {
		t.Errorf("new size is zero")
	}
}
