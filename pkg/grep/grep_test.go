package grep

import (
	"regexp"
	"strings"
	"testing"
)

func TestGrepBasic(t *testing.T) {
	in := "hello\nworld\nhello world\n"
	matches, _ := Run(strings.NewReader(in), "file", nil, []string{"hello"}, false, true, false)
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestGrepInvert(t *testing.T) {
	in := "hello\nworld\nhello world\n"
	matches, _ := Run(strings.NewReader(in), "file", nil, []string{"hello"}, true, true, false)
	if len(matches) != 1 || matches[0].Text != "world" {
		t.Errorf("expected 1 match 'world', got %v", matches)
	}
}

func TestGrepRegex(t *testing.T) {
	in := "hello\nworld\nhello world\n"
	re := regexp.MustCompile(`^w`)
	matches, _ := Run(strings.NewReader(in), "file", re, nil, false, false, false)
	if len(matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Text != "world" {
		t.Errorf("expected 'world', got %q", matches[0].Text)
	}
}

func TestGrepRegexIgnoreCase(t *testing.T) {
	in := "Hello\nWORLD\n"
	re := regexp.MustCompile(`(?i)hello`)
	matches, _ := Run(strings.NewReader(in), "file", re, nil, false, false, false)
	if len(matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(matches))
	}
}

func TestGrepLineRegexp(t *testing.T) {
	in := "hello\nworld\nhello world\n"
	re := regexp.MustCompile(`^hello$`)
	matches, _ := Run(strings.NewReader(in), "file", re, nil, false, false, false)
	if len(matches) != 1 {
		t.Errorf("expected 1 exact line match, got %d", len(matches))
	}
}

func TestGrepWordRegexp(t *testing.T) {
	in := "hello\nworld\nhelloworld\n"
	re := regexp.MustCompile(`\bhello\b`)
	matches, _ := Run(strings.NewReader(in), "file", re, nil, false, false, false)
	if len(matches) != 1 {
		t.Errorf("expected 1 word-boundary match, got %d", len(matches))
	}
	if matches[0].Text != "hello" {
		t.Errorf("expected 'hello', got %q", matches[0].Text)
	}
}

func TestGrepEmptyInput(t *testing.T) {
	matches, err := Run(strings.NewReader(""), "file", nil, []string{"x"}, false, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestGrepNoMatch(t *testing.T) {
	in := "hello\nworld\n"
	matches, _ := Run(strings.NewReader(in), "file", nil, []string{"xyz"}, false, true, false)
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestGrepMultiplePatterns(t *testing.T) {
	in := "apple\nbanana\ncherry\ndate\n"
	matches, _ := Run(strings.NewReader(in), "file", nil, []string{"apple", "cherry"}, false, true, false)
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestGrepLineNumbers(t *testing.T) {
	in := "a\nb\nc\na\n"
	matches, _ := Run(strings.NewReader(in), "file", nil, []string{"a"}, false, true, false)
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Line != 1 {
		t.Errorf("expected line 1, got %d", matches[0].Line)
	}
	if matches[1].Line != 4 {
		t.Errorf("expected line 4, got %d", matches[1].Line)
	}
}
