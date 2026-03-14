package shipinternal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestRunLocalOnlyDeploySkipsSSH(t *testing.T) {
	originalLoadDeployConfig := loadDeployConfig
	originalRunLocalCommand := runLocalCommand
	originalWaitForSSHClient := waitForSSHClient
	defer func() {
		loadDeployConfig = originalLoadDeployConfig
		runLocalCommand = originalRunLocalCommand
		waitForSSHClient = originalWaitForSSHClient
	}()

	var ran []string
	loadDeployConfig = func() (DeployConfig, error) {
		return DeployConfig{
			LocalCommands: []string{
				"npm ci",
				"npm run build",
			},
		}, nil
	}
	runLocalCommand = func(ctx context.Context, command string) error {
		ran = append(ran, command)
		return nil
	}
	waitForSSHClient = func(ctx context.Context, user, host string, interval time.Duration) (*ssh.Client, error) {
		t.Fatal("waitForSSHClient should not be called for a local-only deploy")
		return nil, nil
	}

	err := Run(context.Background(), Options{
		ServerIP: "1.2.3.4",
		User:     "root",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(ran) != 2 {
		t.Fatalf("Run executed %d local commands, want 2", len(ran))
	}
	if ran[0] != "npm ci" || ran[1] != "npm run build" {
		t.Fatalf("Run executed local commands %+v", ran)
	}
}

func TestRunCreatesRemoteUploadDirectoriesBeforeCopy(t *testing.T) {
	originalLoadDeployConfig := loadDeployConfig
	originalRunLocalCommand := runLocalCommand
	originalWaitForSSHClient := waitForSSHClient
	originalCopyDeployFile := copyDeployFile
	originalRunRemoteCommands := runRemoteCommands
	originalCloseSSHClient := closeSSHClient
	defer func() {
		loadDeployConfig = originalLoadDeployConfig
		runLocalCommand = originalRunLocalCommand
		waitForSSHClient = originalWaitForSSHClient
		copyDeployFile = originalCopyDeployFile
		runRemoteCommands = originalRunRemoteCommands
		closeSSHClient = originalCloseSSHClient
	}()

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

	if err := os.WriteFile("release.tar.gz", []byte("artifact"), 0o600); err != nil {
		t.Fatalf("write release.tar.gz: %v", err)
	}

	var events []string
	loadDeployConfig = func() (DeployConfig, error) {
		return DeployConfig{
			Uploads: []DeployUpload{
				{
					Source:      "release.tar.gz",
					Destination: "/opt/app/release.tar.gz",
				},
			},
			RemoteCommands: []string{
				"tar -xzf /opt/app/release.tar.gz -C /opt/app",
			},
		}, nil
	}
	runLocalCommand = func(ctx context.Context, command string) error {
		t.Fatalf("runLocalCommand should not be called, got %q", command)
		return nil
	}
	waitForSSHClient = func(ctx context.Context, user, host string, interval time.Duration) (*ssh.Client, error) {
		events = append(events, "ssh")
		return nil, nil
	}
	closeSSHClient = func(client *ssh.Client) error {
		events = append(events, "close")
		return nil
	}
	copyDeployFile = func(ctx context.Context, client *ssh.Client, localPath, remotePath string, mode os.FileMode) error {
		events = append(events, "copy:"+remotePath)
		return nil
	}
	runRemoteCommands = func(ctx context.Context, client *ssh.Client, commands []string) error {
		events = append(events, "remote:"+strings.Join(commands, " && "))
		return nil
	}

	err = Run(context.Background(), Options{ServerIP: "1.2.3.4", User: "root"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := []string{
		"ssh",
		"remote:mkdir -p '/opt/app'",
		"copy:/opt/app/release.tar.gz",
		"remote:tar -xzf /opt/app/release.tar.gz -C /opt/app",
		"close",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("Run events = %+v, want %+v", events, want)
	}
}

func TestRunDoesNotDeletePreexistingCleanupTargetOnLocalFailure(t *testing.T) {
	originalLoadDeployConfig := loadDeployConfig
	originalRunLocalCommand := runLocalCommand
	defer func() {
		loadDeployConfig = originalLoadDeployConfig
		runLocalCommand = originalRunLocalCommand
	}()

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

	artifactPath := filepath.Join(tempDir, "app.tar")
	if err := os.WriteFile(artifactPath, []byte("preexisting"), 0o600); err != nil {
		t.Fatalf("write app.tar: %v", err)
	}

	loadDeployConfig = func() (DeployConfig, error) {
		return DeployConfig{
			LocalCommands: []string{"docker build -t app ."},
			CleanupLocal:  []string{"app.tar"},
		}, nil
	}
	runLocalCommand = func(ctx context.Context, command string) error {
		return errors.New("boom")
	}

	err = Run(context.Background(), Options{ServerIP: "1.2.3.4", User: "root"})
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if _, err := os.Stat(artifactPath); err != nil {
		t.Fatalf("app.tar was removed after failed local phase: %v", err)
	}
}

func TestRunCleansGeneratedArtifactWhenLaterLocalCommandFails(t *testing.T) {
	originalLoadDeployConfig := loadDeployConfig
	originalRunLocalCommand := runLocalCommand
	defer func() {
		loadDeployConfig = originalLoadDeployConfig
		runLocalCommand = originalRunLocalCommand
	}()

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

	loadDeployConfig = func() (DeployConfig, error) {
		return DeployConfig{
			LocalCommands: []string{
				"build artifact",
				"fail later",
			},
			CleanupLocal: []string{"release.tar.gz"},
		}, nil
	}
	runLocalCommand = func(ctx context.Context, command string) error {
		switch command {
		case "build artifact":
			return os.WriteFile("release.tar.gz", []byte("artifact"), 0o600)
		case "fail later":
			return errors.New("boom")
		default:
			t.Fatalf("unexpected command %q", command)
			return nil
		}
	}

	err = Run(context.Background(), Options{ServerIP: "1.2.3.4", User: "root"})
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "release.tar.gz")); !os.IsNotExist(err) {
		t.Fatalf("release.tar.gz still exists after failed local phase, stat err=%v", err)
	}
}
