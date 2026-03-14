package shipinternal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDeployConfigMissingFileUsesDefault(t *testing.T) {
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

	config, err := LoadDeployConfig()
	if err != nil {
		t.Fatalf("LoadDeployConfig returned error: %v", err)
	}

	if len(config.LocalCommands) != 2 || config.LocalCommands[1] != "docker save app -o app.tar" {
		t.Fatalf("unexpected default local commands: %+v", config.LocalCommands)
	}
	if len(config.Uploads) != 1 || config.Uploads[0].Source != "app.tar" {
		t.Fatalf("unexpected default uploads: %+v", config.Uploads)
	}
	if len(config.RemoteCommands) != 4 || config.RemoteCommands[0] != "docker load -i /root/app.tar" {
		t.Fatalf("unexpected default remote commands: %+v", config.RemoteCommands)
	}
	if len(config.CleanupLocal) != 1 || config.CleanupLocal[0] != "app.tar" {
		t.Fatalf("unexpected default cleanup paths: %+v", config.CleanupLocal)
	}
}

func TestLoadDeployConfigCustomProjectConfig(t *testing.T) {
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

	configFile := `{
  "deploy": {
    "local_commands": [
      "npm ci",
      "tar -czf dist.tar.gz dist"
    ],
    "uploads": [
      {
        "source": "dist.tar.gz",
        "destination": "/opt/app/dist.tar.gz",
        "mode": "0600"
      }
    ],
    "remote_commands": [
      "cd /opt/app && tar -xzf dist.tar.gz"
    ],
    "cleanup_local": [
      "dist.tar.gz"
    ]
  }
}`
	if err := os.WriteFile(projectConfigFile, []byte(configFile), 0o600); err != nil {
		t.Fatalf("write %s: %v", projectConfigFile, err)
	}

	config, err := LoadDeployConfig()
	if err != nil {
		t.Fatalf("LoadDeployConfig returned error: %v", err)
	}

	uploads, err := config.ResolvedUploads(tempDir)
	if err != nil {
		t.Fatalf("ResolvedUploads returned error: %v", err)
	}
	if len(uploads) != 1 {
		t.Fatalf("ResolvedUploads len = %d, want 1", len(uploads))
	}
	if uploads[0].Source != filepath.Join(tempDir, "dist.tar.gz") {
		t.Fatalf("ResolvedUploads source = %q", uploads[0].Source)
	}
	if uploads[0].Destination != "/opt/app/dist.tar.gz" {
		t.Fatalf("ResolvedUploads destination = %q", uploads[0].Destination)
	}
	if uploads[0].Mode != 0o600 {
		t.Fatalf("ResolvedUploads mode = %o, want 0600", uploads[0].Mode)
	}

	cleanup := config.ResolvedCleanupPaths(tempDir)
	if len(cleanup) != 1 || cleanup[0] != filepath.Join(tempDir, "dist.tar.gz") {
		t.Fatalf("ResolvedCleanupPaths = %+v", cleanup)
	}
}

func TestLoadDeployConfigRejectsNumericUploadMode(t *testing.T) {
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

	configFile := `{
  "deploy": {
    "uploads": [
      {
        "source": "dist.tar.gz",
        "destination": "/opt/app/dist.tar.gz",
        "mode": 644
      }
    ]
  }
}`
	if err := os.WriteFile(projectConfigFile, []byte(configFile), 0o600); err != nil {
		t.Fatalf("write %s: %v", projectConfigFile, err)
	}

	if _, err := LoadDeployConfig(); err == nil {
		t.Fatal("LoadDeployConfig returned nil error for numeric mode")
	}
}

func TestLoadDeployConfigRejectsEmptyDeployBlock(t *testing.T) {
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

	configFile := `{
  "deploy": {}
}`
	if err := os.WriteFile(projectConfigFile, []byte(configFile), 0o600); err != nil {
		t.Fatalf("write %s: %v", projectConfigFile, err)
	}

	if _, err := LoadDeployConfig(); err == nil {
		t.Fatal("LoadDeployConfig returned nil error for empty deploy block")
	}
}

func TestLoadDeployConfigUsesSavedRuntimeProxyWhenProjectConfigHasNone(t *testing.T) {
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

	if err := os.MkdirAll(".ship", 0o755); err != nil {
		t.Fatalf("mkdir .ship: %v", err)
	}
	runtimeConfig := `{
  "proxy": {
    "domains": ["example.com"],
    "app_port": 9090
  }
}`
	if err := os.WriteFile(filepath.Join(".ship", "runtime.json"), []byte(runtimeConfig), 0o600); err != nil {
		t.Fatalf("write runtime.json: %v", err)
	}

	config, err := LoadDeployConfig()
	if err != nil {
		t.Fatalf("LoadDeployConfig returned error: %v", err)
	}
	if len(config.RemoteCommands) == 0 {
		t.Fatal("RemoteCommands was empty")
	}
	if !strings.Contains(config.RemoteCommands[len(config.RemoteCommands)-1], "127.0.0.1:9090:80") {
		t.Fatalf("default deploy command = %q", config.RemoteCommands[len(config.RemoteCommands)-1])
	}
}
