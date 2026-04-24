package echo

import (
	"testing"
)

func TestRunBasic(t *testing.T) {
	result := Run(false, false, []string{"hello", "world"})
	if result.Text != "hello world" {
		t.Errorf("got %q, want %q", result.Text, "hello world")
	}
}

func TestRunEscapes(t *testing.T) {
	result := Run(false, true, []string{`hello\nworld`})
	if result.Text != "hello\nworld" {
		t.Errorf("got %q, want %q", result.Text, "hello\nworld")
	}
}

func TestRunNoEscapes(t *testing.T) {
	// Without -e, backslash sequences should NOT be expanded.
	result := Run(false, false, []string{`hello\nworld`})
	if result.Text != `hello\nworld` {
		t.Errorf("got %q, want literal backslash-n", result.Text)
	}
}

func TestRunEmpty(t *testing.T) {
	result := Run(false, false, []string{})
	if result.Text != "" {
		t.Errorf("got %q, want empty string", result.Text)
	}
}

func TestRunMultipleEscapes(t *testing.T) {
	result := Run(false, true, []string{`a\tb`})
	if result.Text != "a\tb" {
		t.Errorf("got %q, want tab", result.Text)
	}
}
