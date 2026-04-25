package truefalse

import "testing"

func TestTrueExitsZero(t *testing.T) {
	if code := runTrue(nil, nil); code != 0 {
		t.Errorf("true: expected exit 0, got %d", code)
	}
}

func TestFalseExitsOne(t *testing.T) {
	if code := runFalse(nil, nil); code != 1 {
		t.Errorf("false: expected exit 1, got %d", code)
	}
}
