package tr

import (
	"bytes"
	"strings"
	"testing"
)

func TestTrTranslate(t *testing.T) {
	in := "hello world"
	var out bytes.Buffer
	Run(strings.NewReader(in), &out, "a-z", "A-Z", false, false, false)
	if out.String() != "HELLO WORLD" {
		t.Errorf("got %v", out.String())
	}
}

func TestTrDelete(t *testing.T) {
	in := "hello world"
	var out bytes.Buffer
	Run(strings.NewReader(in), &out, "l", "", true, false, false)
	if out.String() != "heo word" {
		t.Errorf("got %v", out.String())
	}
}

func TestTrSqueeze(t *testing.T) {
	in := "heeeello   world"
	var out bytes.Buffer
	Run(strings.NewReader(in), &out, "e ", "", false, true, false)
	if out.String() != "hello world" {
		t.Errorf("got %v", out.String())
	}
}
