package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	exitCode := run([]string{"mt", "version"}, &out, &errOut)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(out.String(), "mt") {
		t.Fatalf("expected version output, got %s", out.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	exitCode := run([]string{"mt", "nope"}, &out, &errOut)
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code")
	}
	if !strings.Contains(errOut.String(), "unknown command") {
		t.Fatalf("expected error output, got %s", errOut.String())
	}
}

func TestMainUsesExitCode(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"mt", "version"}
	defer func() {
		os.Args = originalArgs
	}()

	originalExit := exit
	defer func() {
		exit = originalExit
	}()

	var got int
	exit = func(code int) {
		got = code
	}

	main()

	if got != 0 {
		t.Fatalf("expected exit code 0, got %d", got)
	}
}
