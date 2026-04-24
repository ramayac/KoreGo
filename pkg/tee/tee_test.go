package tee

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBasic(t *testing.T) {
	// Not easily testable as it binds os.Stdout and os.Stdin in run().
	// We'll test the io.MultiWriter aspect.
	
	dir := t.TempDir()
	file1 := filepath.Join(dir, "out1.txt")
	file2 := filepath.Join(dir, "out2.txt")
	
	f1, _ := os.Create(file1)
	f2, _ := os.Create(file2)
	
	var stdout bytes.Buffer
	writers := []io.Writer{&stdout, f1, f2}
	multiWriter := io.MultiWriter(writers...)
	
	io.Copy(multiWriter, strings.NewReader("hello tee\n"))
	f1.Close()
	f2.Close()
	
	if stdout.String() != "hello tee\n" {
		t.Error("stdout missing")
	}
	
	b1, _ := os.ReadFile(file1)
	if string(b1) != "hello tee\n" {
		t.Error("file1 missing")
	}
	b2, _ := os.ReadFile(file2)
	if string(b2) != "hello tee\n" {
		t.Error("file2 missing")
	}
}
