package uniq

import (
	"strings"
	"testing"
)

func TestUniqBasic(t *testing.T) {
	in := "a\na\nb\nc\nc\n"
	res, _ := Run(strings.NewReader(in), false, false, false, false, 0, 0, 0)
	if len(res) != 3 || res[0].Line != "a" || res[1].Line != "b" || res[2].Line != "c" {
		t.Errorf("got %v", res)
	}
}

func TestUniqCount(t *testing.T) {
	in := "a\na\nb\n"
	res, _ := Run(strings.NewReader(in), true, false, false, false, 0, 0, 0)
	if len(res) != 2 || res[0].Count != 2 || res[1].Count != 1 {
		t.Errorf("got %v", res)
	}
}

func TestUniqDuplicates(t *testing.T) {
	in := "a\na\nb\nc\nc\n"
	res, _ := Run(strings.NewReader(in), false, true, false, false, 0, 0, 0)
	if len(res) != 2 || res[0].Line != "a" || res[1].Line != "c" {
		t.Errorf("got %v", res)
	}
}

func TestUniqUnique(t *testing.T) {
	in := "a\na\nb\nc\nc\n"
	res, _ := Run(strings.NewReader(in), false, false, true, false, 0, 0, 0)
	if len(res) != 1 || res[0].Line != "b" {
		t.Errorf("got %v", res)
	}
}

func TestUniqIgnoreCase(t *testing.T) {
	in := "a\nA\nb\n"
	res, _ := Run(strings.NewReader(in), false, false, false, true, 0, 0, 0)
	if len(res) != 2 || res[0].Line != "a" || res[0].Count != 2 {
		t.Errorf("got %v", res)
	}
}
