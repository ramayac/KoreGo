package common

import (
	"testing"
)

// testSpec is a convenience spec used by most tests.
var testSpec = FlagSpec{
	Defs: []FlagDef{
		{Short: "l", Long: "list", Type: FlagBool},
		{Short: "a", Long: "all", Type: FlagBool},
		{Short: "R", Long: "recursive", Type: FlagBool},
		{Short: "v", Long: "verbose", Type: FlagBool},
		{Short: "n", Long: "no-newline", Type: FlagBool},
		{Short: "e", Long: "escape", Type: FlagBool},
		{Short: "P", Long: "physical", Type: FlagBool},
		{Short: "i", Long: "ignore-env", Type: FlagBool},
		{Short: "o", Long: "output", Type: FlagValue},
	},
}

func TestGroupedShortFlags(t *testing.T) {
	args := []string{"-laR", "/tmp"}
	result, err := ParseFlags(args, testSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Has("l") || !result.Has("a") || !result.Has("R") {
		t.Errorf("expected l, a, R flags to be set; got %+v", result.Bools)
	}
	if len(result.Positional) != 1 || result.Positional[0] != "/tmp" {
		t.Errorf("expected positional [/tmp], got %v", result.Positional)
	}
}

func TestLongFlagBool(t *testing.T) {
	result, err := ParseFlags([]string{"--all", "--recursive"}, testSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Has("all") || !result.Has("recursive") {
		t.Errorf("long flags not set: %+v", result.Bools)
	}
	// Short aliases should also be set.
	if !result.Has("a") || !result.Has("R") {
		t.Errorf("short aliases not set: %+v", result.Bools)
	}
}

func TestLongFlagEqValue(t *testing.T) {
	result, err := ParseFlags([]string{"--output=foo.txt"}, testSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Get("output") != "foo.txt" {
		t.Errorf("expected output=foo.txt, got %q", result.Get("output"))
	}
}

func TestLongFlagSpaceValue(t *testing.T) {
	result, err := ParseFlags([]string{"--output", "bar.txt"}, testSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Get("output") != "bar.txt" {
		t.Errorf("expected output=bar.txt, got %q", result.Get("output"))
	}
}

func TestEndOfFlags(t *testing.T) {
	args := []string{"--", "-not-a-flag"}
	result, err := ParseFlags(args, FlagSpec{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Positional) != 1 || result.Positional[0] != "-not-a-flag" {
		t.Errorf("expected positional [-not-a-flag], got %v", result.Positional)
	}
}

func TestStdinMarker(t *testing.T) {
	result, err := ParseFlags([]string{"-"}, testSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Stdin {
		t.Error("expected Stdin=true")
	}
	if len(result.Positional) != 1 || result.Positional[0] != "-" {
		t.Errorf("expected '-' in Positional, got %v", result.Positional)
	}
}

func TestUnknownFlag(t *testing.T) {
	_, err := ParseFlags([]string{"-z"}, FlagSpec{})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	fe, ok := err.(*FlagError)
	if !ok {
		t.Fatalf("expected *FlagError, got %T", err)
	}
	if fe.ExitCode != 2 {
		t.Errorf("expected ExitCode 2, got %d", fe.ExitCode)
	}
}

func TestUnknownLongFlag(t *testing.T) {
	_, err := ParseFlags([]string{"--nope"}, FlagSpec{})
	if err == nil {
		t.Fatal("expected error for unknown long flag")
	}
	fe, _ := err.(*FlagError)
	if fe.ExitCode != 2 {
		t.Errorf("expected ExitCode 2, got %d", fe.ExitCode)
	}
}

func TestFlagRepetitionCounting(t *testing.T) {
	result, err := ParseFlags([]string{"-v", "-v", "-v"}, testSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count["v"] != 3 {
		t.Errorf("expected verbosity count 3, got %d", result.Count["v"])
	}
}

func TestFlagsMixedWithPositional(t *testing.T) {
	result, err := ParseFlags([]string{"-l", "/tmp", "-a", "/home"}, testSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Has("l") || !result.Has("a") {
		t.Errorf("expected l and a flags; got %+v", result.Bools)
	}
	if len(result.Positional) != 2 {
		t.Errorf("expected 2 positionals, got %v", result.Positional)
	}
}

func TestEmptyArgs(t *testing.T) {
	result, err := ParseFlags([]string{}, testSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Positional) != 0 {
		t.Errorf("expected no positionals, got %v", result.Positional)
	}
}

func TestShortValueInCluster(t *testing.T) {
	// -ofoo.txt should parse as -o foo.txt (value in same cluster)
	result, err := ParseFlags([]string{"-ofoo.txt"}, testSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Get("o") != "foo.txt" {
		t.Errorf("expected o=foo.txt, got %q", result.Get("o"))
	}
}
