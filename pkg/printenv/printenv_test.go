package printenv

import (
	"os"
	"testing"
)

func TestRunAllVars(t *testing.T) {
	result := Run(nil)
	if len(result.Vars) == 0 {
		t.Error("expected at least one environment variable")
	}
}

func TestRunSpecificVar(t *testing.T) {
	os.Setenv("COREGOLINUX_TEST", "hello")
	defer os.Unsetenv("COREGOLINUX_TEST")

	result := Run([]string{"COREGOLINUX_TEST"})
	if result.Vars["COREGOLINUX_TEST"] != "hello" {
		t.Errorf("got %q, want hello", result.Vars["COREGOLINUX_TEST"])
	}
}

func TestRunMissingVar(t *testing.T) {
	result := Run([]string{"COREGOLINUX_DEFINITELY_NOT_SET_XYZ"})
	if _, ok := result.Vars["COREGOLINUX_DEFINITELY_NOT_SET_XYZ"]; ok {
		t.Error("expected missing var to be absent from results")
	}
}
