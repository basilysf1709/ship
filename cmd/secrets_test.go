package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSecretsSetCommandUpdatesSecrets(t *testing.T) {
	originalLoadSecrets := loadSecrets
	originalSaveSecrets := saveSecrets
	defer func() {
		loadSecrets = originalLoadSecrets
		saveSecrets = originalSaveSecrets
	}()

	loadSecrets = func() (map[string]string, error) {
		return map[string]string{"EXISTING": "1"}, nil
	}

	var saved map[string]string
	saveSecrets = func(secrets map[string]string) error {
		saved = secrets
		return nil
	}

	cmd := newSecretsSetCommand()
	cmd.SetArgs([]string{"API_KEY=secret", "PORT=3000"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if saved["API_KEY"] != "secret" || saved["PORT"] != "3000" || saved["EXISTING"] != "1" {
		t.Fatalf("saved secrets = %+v", saved)
	}
}

func TestSecretsListCommandOutputsKeys(t *testing.T) {
	originalLoadSecrets := loadSecrets
	defer func() {
		loadSecrets = originalLoadSecrets
	}()

	loadSecrets = func() (map[string]string, error) {
		return map[string]string{"API_KEY": "secret", "PORT": "3000"}, nil
	}

	cmd := newSecretsListCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "API_KEY") || !strings.Contains(output, "PORT") || !strings.Contains(output, "TOTAL_SECRETS=2") {
		t.Fatalf("output = %q", output)
	}
}
