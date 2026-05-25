package app

import (
	"bytes"
	"testing"
)

func TestBuildVersionUsesOverride(t *testing.T) {
	prevVersionOverride := versionOverride
	defer func() { versionOverride = prevVersionOverride }()

	versionOverride = "v0.1.0"
	if got := buildVersion(); got != "v0.1.0" {
		t.Fatalf("buildVersion() = %q, want v0.1.0", got)
	}
}

func TestRootVersionCommandPrintsVersion(t *testing.T) {
	prevVersionOverride := versionOverride
	defer func() { versionOverride = prevVersionOverride }()
	versionOverride = "v0.1.0"

	cmd := NewRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.String() != "v0.1.0\n" {
		t.Fatalf("unexpected output %q", out.String())
	}
}

func TestRootVersionFlagPrintsVersion(t *testing.T) {
	prevVersionOverride := versionOverride
	defer func() { versionOverride = prevVersionOverride }()
	versionOverride = "v0.1.0"

	cmd := NewRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.String() != "v0.1.0\n" {
		t.Fatalf("unexpected output %q", out.String())
	}
}
