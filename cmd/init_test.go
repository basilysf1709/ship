package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestInitCommandWritesTemplate(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	cmd := newInitCommand()
	cmd.SetArgs([]string{"--template", "docker"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	data, err := os.ReadFile("ship.json")
	if err != nil {
		t.Fatalf("read ship.json: %v", err)
	}
	if !strings.Contains(string(data), "\"healthcheck_path\": \"/\"") {
		t.Fatalf("ship.json = %q", string(data))
	}
}
