package wc

import (
	"strings"
	"testing"
)

func TestCountProper(t *testing.T) {
	in := strings.NewReader("hello world\nthis is a test\n")
	res, err := CountProper(in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Lines != 2 {
		t.Errorf("expected 2 lines, got %d", res.Lines)
	}
	if res.Words != 6 {
		t.Errorf("expected 6 words, got %d", res.Words)
	}
	if res.Bytes != 27 {
		t.Errorf("expected 27 bytes, got %d", res.Bytes)
	}
}

func TestCountProperUTF8(t *testing.T) {
	in := strings.NewReader("こんにちは\n")
	res, _ := CountProper(in)
	if res.Lines != 1 {
		t.Errorf("expected 1 line")
	}
	if res.Words != 1 {
		t.Errorf("expected 1 word") // No spaces between Japanese chars, counts as 1 word
	}
	if res.Chars != 6 {
		t.Errorf("expected 6 chars, got %d", res.Chars) // 5 chars + newline
	}
	if res.Bytes != 16 {
		t.Errorf("expected 16 bytes, got %d", res.Bytes) // 5*3 + 1
	}
}
