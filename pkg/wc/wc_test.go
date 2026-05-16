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
		t.Errorf("expected 1 word")
	}
	if res.Chars != 6 {
		t.Errorf("expected 6 chars, got %d", res.Chars)
	}
	if res.Bytes != 16 {
		t.Errorf("expected 16 bytes, got %d", res.Bytes)
	}
}

func TestCountProperEmpty(t *testing.T) {
	in := strings.NewReader("")
	res, err := CountProper(in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Lines != 0 || res.Words != 0 || res.Bytes != 0 || res.Chars != 0 {
		t.Errorf("expected all zeros, got %+v", res)
	}
}

func TestCountProperSingleLine(t *testing.T) {
	in := strings.NewReader("hello")
	res, err := CountProper(in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Lines != 0 {
		t.Errorf("expected 0 lines (no newline), got %d", res.Lines)
	}
	if res.Words != 1 {
		t.Errorf("expected 1 word, got %d", res.Words)
	}
	if res.Bytes != 5 {
		t.Errorf("expected 5 bytes, got %d", res.Bytes)
	}
}

func TestCountProperWhitespace(t *testing.T) {
	in := strings.NewReader("   \t  \n")
	res, err := CountProper(in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Lines != 1 {
		t.Errorf("expected 1 line, got %d", res.Lines)
	}
	if res.Words != 0 {
		t.Errorf("expected 0 words, got %d", res.Words)
	}
}

func TestCount(t *testing.T) {
	in := strings.NewReader("hello world\n")
	res, err := Count(in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Lines != 1 {
		t.Errorf("expected 1 line, got %d", res.Lines)
	}
	if res.Words != 2 {
		t.Errorf("expected 2 words, got %d", res.Words)
	}
}

// --- BusyBox test suite hardening ---

func TestBusyBox_Wc_MaxLineLength(t *testing.T) {
	// BusyBox: echo "i'm a little teapot" | wc -L → 19
	// -L returns the length of the longest line.
	r := strings.NewReader("i'm a little teapot\n")
	res, err := CountProper(r)
	if err != nil {
		t.Fatal(err)
	}
	if res.MaxLineLength != 19 {
		t.Errorf("got %d, want 19", res.MaxLineLength)
	}
}
