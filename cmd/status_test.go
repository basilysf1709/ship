package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	shipinternal "ship/internal"
)

func TestStatusCommandReportsDriftAndRemoteState(t *testing.T) {
	originalLoadCurrentServerState := loadCurrentServerState
	originalLatestReleaseRecord := latestReleaseRecord
	originalWaitForSSH := waitForSSH
	originalRunRemoteCommand := runRemoteCommand
	originalLoadProjectConfig := loadProjectConfig
	originalCurrentGitSHA := currentGitSHA
	originalCurrentSecretsChecksum := currentSecretsChecksum
	originalLoadDeployConfig := loadDeployConfig
	originalDeployConfigHash := deployConfigHash
	originalCheckHealthURLFunc := checkHealthURLFunc
	defer func() {
		loadCurrentServerState = originalLoadCurrentServerState
		latestReleaseRecord = originalLatestReleaseRecord
		waitForSSH = originalWaitForSSH
		runRemoteCommand = originalRunRemoteCommand
		loadProjectConfig = originalLoadProjectConfig
		currentGitSHA = originalCurrentGitSHA
		currentSecretsChecksum = originalCurrentSecretsChecksum
		loadDeployConfig = originalLoadDeployConfig
		deployConfigHash = originalDeployConfigHash
		checkHealthURLFunc = originalCheckHealthURLFunc
	}()

	loadCurrentServerState = func() (shipinternal.ServerState, error) {
		return shipinternal.ServerState{
			ServerID: "srv-1",
			IP:       "1.2.3.4",
			Provider: "digitalocean",
			SSHUser:  "root",
		}, nil
	}
	latestReleaseRecord = func() (*shipinternal.ReleaseRecord, error) {
		return &shipinternal.ReleaseRecord{
			ID:               "rel-1",
			CreatedAt:        "2026-03-13T12:00:00Z",
			Status:           "success",
			Stage:            "complete",
			GitSHA:           "oldsha",
			DeployHash:       "deploy-old",
			SecretsChecksum:  "local-secrets",
			RollbackEligible: true,
		}, nil
	}
	waitForSSH = func(ctx context.Context, user, host string, interval time.Duration) (*ssh.Client, error) {
		return nil, nil
	}
	runRemoteCommand = func(ctx context.Context, client *ssh.Client, command string) (string, error) {
		switch {
		case strings.Contains(command, "docker inspect -f '{{.State.Status}}'"):
			return "running\n", nil
		case strings.Contains(command, "docker inspect -f '{{if .State.Health}}"):
			return "healthy\n", nil
		case strings.Contains(command, "systemctl is-active caddy"):
			return "active\n", nil
		case strings.Contains(command, "sha256sum"):
			return "remote-secrets\n", nil
		default:
			return "", nil
		}
	}
	loadProjectConfig = func() (shipinternal.ProjectConfig, error) {
		return shipinternal.ProjectConfig{
			Proxy:  &shipinternal.ProxyConfig{Domains: []string{"example.com"}},
			Status: &shipinternal.StatusConfig{HealthcheckPath: "/health"},
		}, nil
	}
	currentGitSHA = func() string {
		return "newsha"
	}
	currentSecretsChecksum = func() (string, error) {
		return "local-secrets", nil
	}
	loadDeployConfig = func() (shipinternal.DeployConfig, error) {
		return shipinternal.DeployConfig{LocalCommands: []string{"docker build -t app ."}}, nil
	}
	deployConfigHash = func(config shipinternal.DeployConfig) (string, error) {
		return "deploy-new", nil
	}
	checkHealthURLFunc = func(url string) string {
		if url != "https://example.com/health" {
			t.Fatalf("health url = %q", url)
		}
		return "ok:200"
	}

	cmd := newStatusCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"SSH_STATUS=reachable",
		"APP_STATUS=running",
		"APP_HEALTH=healthy",
		"PROXY_STATUS=active",
		"HEALTH_STATUS=ok:200",
		"GIT_DRIFT=true",
		"DEPLOY_DRIFT=true",
		"SECRETS_STATUS=drifted",
		"ROLLBACK_STATUS=available",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q: %q", want, output)
		}
	}
}
