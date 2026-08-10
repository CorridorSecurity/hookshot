package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestInstallCommandIsUnavailable(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "install")
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("hookshot install succeeded")
	}

	message := string(output)
	if !strings.Contains(message, "Unknown command: install") {
		t.Fatalf("unexpected output: %s", message)
	}
	if strings.Contains(message, "Install hooks") {
		t.Fatalf("usage still advertises hook installation: %s", message)
	}
}
