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

func TestParseHumanVal(t *testing.T) {
	cases := []struct {
		in     string
		num    float64
		hasSuf bool
		suffix int
	}{
		{"", 0, false, 0},
		{"5", 5, false, 0},
		{"1.5K", 1.5, true, 1},
		{"10M", 10, true, 3},
		{"100G", 100, true, 4},
		{"2.5T", 2.5, true, 5},
		{"abc", 0, false, 0},
	}
	for _, c := range cases {
		got := parseHumanVal(c.in)
		if got.num != c.num || got.hasSuffix != c.hasSuf || got.suffix != c.suffix {
			t.Errorf("parseHumanVal(%q) = {num:%v, hasSuffix:%v, suffix:%v}, want {num:%v, hasSuffix:%v, suffix:%v}",
				c.in, got.num, got.hasSuffix, got.suffix, c.num, c.hasSuf, c.suffix)
		}
	}
}

func TestParseNumericPrefix(t *testing.T) {
	cases := []struct {
		in    string
		want  float64
		valid bool
	}{
		{"123", 123, true},
		{"-5.5", -5.5, true},
		{"+10", 10, true},
		{"abc", 0, false},
		{"  42", 42, true},
		{"3.14xxx", 3.14, true},
		{"", 0, false},
	}
	for _, c := range cases {
		got, valid := parseNumericPrefix(c.in)
		if got != c.want || valid != c.valid {
			t.Errorf("parseNumericPrefix(%q) = (%v, %v), want (%v, %v)", c.in, got, valid, c.want, c.valid)
		}
	}
}

func TestParseMonth(t *testing.T) {
	cases := []struct {
		in    string
		want  int
		valid bool
	}{
		{"Jan", 1, true},
		{"jan", 1, true},
		{"JANUARY", 1, true},
		{"feb", 2, true},
		{"DEC", 12, true},
		{"December", 12, true},
		{"xyz", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, valid := parseMonth(c.in)
		if got != c.want || valid != c.valid {
			t.Errorf("parseMonth(%q) = (%v, %v), want (%v, %v)", c.in, got, valid, c.want, c.valid)
		}
	}
}

func TestExtractKey(t *testing.T) {
	cases := []struct {
		line      string
		ks        keySpec
		delimiter string
		want      string
	}{
		{"a b c", keySpec{startField: 2}, "", "b"},
		{"a,b,c", keySpec{startField: 2}, ",", "b"},
		{"a b c d", keySpec{startField: 2, endField: 3}, "", "b c"},
		{"single", keySpec{startField: 0}, "", "single"},
		{"abcdef", keySpec{startField: 1, startChar: 2}, "", "bcdef"},
		{"abcdef", keySpec{startField: 1, startChar: 2, endChar: 4}, "", "bcd"},
	}
	for _, c := range cases {
		got := extractKey(c.line, c.ks, c.delimiter)
		if got != c.want {
			t.Errorf("extractKey(%q, %+v, %q) = %q, want %q", c.line, c.ks, c.delimiter, got, c.want)
		}
	}
}

func TestParseKeySpec(t *testing.T) {
	specs := parseKeySpec("2,3nr", false, false, false, false)
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	ks := specs[0]
	if ks.startField != 2 || ks.endField != 3 || !ks.numeric || !ks.reverse {
		t.Errorf("got %+v", ks)
	}
}

func TestParseFieldChar(t *testing.T) {
	f, c := parseFieldChar("3.5")
	if f != 3 || c != 5 {
		t.Errorf("got field=%d char=%d, want field=3 char=5", f, c)
	}
	f2, c2 := parseFieldChar("4")
	if f2 != 4 || c2 != 0 {
		t.Errorf("got field=%d char=%d, want field=4 char=0", f2, c2)
	}
}

func TestScanNUL(t *testing.T) {
	data := []byte("a\x00b\x00c")
	adv, tok, err := scanNUL(data, false)
	if err != nil || adv != 2 || string(tok) != "a" {
		t.Errorf("first token: adv=%d tok=%q err=%v", adv, tok, err)
	}

	adv, tok, err = scanNUL(data[2:], false)
	if err != nil || adv != 2 || string(tok) != "b" {
		t.Errorf("second token: adv=%d tok=%q err=%v", adv, tok, err)
	}

	adv, tok, err = scanNUL(data[4:], true)
	if err != nil || adv != 1 || string(tok) != "c" {
		t.Errorf("last token (atEOF): adv=%d tok=%q err=%v", adv, tok, err)
	}
}

func TestCompareHuman(t *testing.T) {
	a := parseHumanVal("5K")
	b := parseHumanVal("10K")
	if compareHuman(a, b) >= 0 {
		t.Error("5K should be less than 10K")
	}

	c := parseHumanVal("1M")
	d := parseHumanVal("500K")
	if compareHuman(c, d) <= 0 {
		t.Error("1M should be greater than 500K")
	}

	e := parseHumanVal("5")
	f := parseHumanVal("5K")
	if compareHuman(e, f) >= 0 {
		t.Error("no-suffix should sort before suffix")
	}
}

func TestCmpFloat(t *testing.T) {
	if cmpFloat(1.0, 2.0) >= 0 {
		t.Error("1 < 2")
	}
	if cmpFloat(2.0, 1.0) <= 0 {
		t.Error("2 > 1")
	}
	if cmpFloat(1.0, 1.0) != 0 {
		t.Error("1 == 1")
	}
}

func TestSortStable(t *testing.T) {
	in := "c\nb\na\n"
	items, _ := parseLines(strings.NewReader(in), nil, "", false)
	res := Run(items, nil, false, false, false, false, false, true, false)
	if !reflect.DeepEqual(res, []string{"a", "b", "c"}) {
		t.Errorf("stable sort got %v", res)
	}
}

func TestParseLinesNUL(t *testing.T) {
	in := strings.NewReader("z\x00a\x00m\x00")
	items, err := parseLines(in, nil, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}
