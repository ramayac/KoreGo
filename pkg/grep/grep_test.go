package grep

import (
	"strings"
	"testing"
)

func TestGrepBasic(t *testing.T) {
	in := "hello\nworld\nhello world\n"
	matches, _ := Run(strings.NewReader(in), "file", nil, "hello", false, true, false)
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestGrepInvert(t *testing.T) {
	in := "hello\nworld\nhello world\n"
	matches, _ := Run(strings.NewReader(in), "file", nil, "hello", true, true, false)
	if len(matches) != 1 || matches[0].Text != "world" {
		t.Errorf("expected 1 match 'world', got %v", matches)
	}
}
