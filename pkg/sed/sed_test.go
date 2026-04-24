package sed

import (
	"bytes"
	"strings"
	"testing"
)

func TestSedSubstitute(t *testing.T) {
	in := "hello world"
	var out bytes.Buffer
	cmd, _ := parseExpr("s/world/korego/")
	Run(strings.NewReader(in), &out, cmd, false)
	if out.String() != "hello korego\n" {
		t.Errorf("got %q", out.String())
	}
}

func TestSedGlobal(t *testing.T) {
	in := "a a a"
	var out bytes.Buffer
	cmd, _ := parseExpr("s/a/b/g")
	Run(strings.NewReader(in), &out, cmd, false)
	if out.String() != "b b b\n" {
		t.Errorf("got %q", out.String())
	}
}
