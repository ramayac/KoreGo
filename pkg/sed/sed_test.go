package sed

import (
	"bytes"
	"os"
	"testing"
)

func TestSedSubstitute(t *testing.T) {
	in := "hello world\n"
	
	f, _ := os.CreateTemp("", "sedtest")
	defer os.Remove(f.Name())
	f.WriteString(in)
	f.Close()

	var out bytes.Buffer
	insts, _ := Parse("s/world/korego/")
	runEngine(insts, []string{f.Name()}, false, false, &out)

	if out.String() != "hello korego\n" {
		t.Errorf("got %q", out.String())
	}
}

func TestSedGlobal(t *testing.T) {
	in := "a a a\n"

	f, _ := os.CreateTemp("", "sedtest")
	defer os.Remove(f.Name())
	f.WriteString(in)
	f.Close()

	var out bytes.Buffer
	insts, _ := Parse("s/a/b/g")
	runEngine(insts, []string{f.Name()}, false, false, &out)

	if out.String() != "b b b\n" {
		t.Errorf("got %q", out.String())
	}
}
