package shipinternal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretsChecksumStableAcrossOrdering(t *testing.T) {
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

	if err := SaveSecrets(map[string]string{"B": "2", "A": "1"}); err != nil {
		t.Fatalf("SaveSecrets returned error: %v", err)
	}
	first, err := SecretsChecksum()
	if err != nil {
		t.Fatalf("SecretsChecksum returned error: %v", err)
	}

	if err := os.WriteFile(secretsPath(), []byte("A=1\nB=2\n"), 0o600); err != nil {
		t.Fatalf("write secrets file: %v", err)
	}
	second, err := SecretsChecksum()
	if err != nil {
		t.Fatalf("SecretsChecksum returned error: %v", err)
	}
	if first == "" || second == "" || first != second {
		t.Fatalf("SecretsChecksum mismatch: %q vs %q", first, second)
	}
}

func TestLoadSecretsRejectsInsecurePermissions(t *testing.T) {
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

	if err := os.MkdirAll(filepath.Join(tempDir, shipDir), 0o755); err != nil {
		t.Fatalf("mkdir .ship: %v", err)
	}
	if err := os.WriteFile(secretsPath(), []byte("API_KEY=secret\n"), 0o644); err != nil {
		t.Fatalf("write secrets file: %v", err)
	}

	if _, err := LoadSecrets(); err == nil {
		t.Fatal("LoadSecrets returned nil error for insecure permissions")
	}
}

func TestSaveSecretsRejectsInvalidKeysAndValues(t *testing.T) {
	if err := SaveSecrets(map[string]string{"NOT-VALID": "secret"}); err == nil {
		t.Fatal("SaveSecrets returned nil error for invalid key")
	}
	if err := SaveSecrets(map[string]string{"VALID_KEY": "line1\nline2"}); err == nil {
		t.Fatal("SaveSecrets returned nil error for multiline value")
	}
}
