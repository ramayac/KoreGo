package env

import (
	"os"
	"testing"
)

func TestRunDefaultsToHostEnv(t *testing.T) {
	os.Setenv("COREGO_ENV_TEST", "yes")
	defer os.Unsetenv("COREGO_ENV_TEST")

	result := Run(false, nil)
	if result.Vars["COREGO_ENV_TEST"] != "yes" {
		t.Errorf("expected COREGO_ENV_TEST=yes in env output")
	}
}

func TestRunIgnoreEnvironment(t *testing.T) {
	os.Setenv("COREGO_ENV_TEST", "should_not_appear")
	defer os.Unsetenv("COREGO_ENV_TEST")

	result := Run(true, nil)
	if _, ok := result.Vars["COREGO_ENV_TEST"]; ok {
		t.Error("expected host env to be cleared when -i is set")
	}
}

func TestRunVarAssignment(t *testing.T) {
	result := Run(true, []string{"FOO=bar", "BAZ=qux"})
	if result.Vars["FOO"] != "bar" {
		t.Errorf("FOO: got %q, want bar", result.Vars["FOO"])
	}
	if result.Vars["BAZ"] != "qux" {
		t.Errorf("BAZ: got %q, want qux", result.Vars["BAZ"])
	}
}
