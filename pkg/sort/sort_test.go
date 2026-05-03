package sort

import (
	"reflect"
	"strings"
	"testing"
)

func TestSortBasic(t *testing.T) {
	in := "c\nb\na\n"
	items, _ := parseLines(strings.NewReader(in), nil, "", false)
	res := Run(items, nil, false, false, false, false, false, false, false)
	if !reflect.DeepEqual(res, []string{"a", "b", "c"}) {
		t.Errorf("got %v", res)
	}
}

func TestSortReverse(t *testing.T) {
	in := "a\nb\nc\n"
	items, _ := parseLines(strings.NewReader(in), nil, "", false)
	res := Run(items, nil, true, false, false, false, false, false, false)
	if !reflect.DeepEqual(res, []string{"c", "b", "a"}) {
		t.Errorf("got %v", res)
	}
}

func TestSortNumeric(t *testing.T) {
	in := "10\n2\n1\n"
	items, _ := parseLines(strings.NewReader(in), nil, "", false)
	res := Run(items, nil, false, true, false, false, false, false, false)
	if !reflect.DeepEqual(res, []string{"1", "2", "10"}) {
		t.Errorf("got %v", res)
	}
}

func TestSortUnique(t *testing.T) {
	in := "b\na\nb\nc\n"
	items, _ := parseLines(strings.NewReader(in), nil, "", false)
	res := Run(items, nil, false, false, true, false, false, false, false)
	if !reflect.DeepEqual(res, []string{"a", "b", "c"}) {
		t.Errorf("got %v", res)
	}
}
