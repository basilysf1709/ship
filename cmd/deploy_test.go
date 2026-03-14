package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	shipinternal "ship/internal"
)

func TestDeployCommandAllowsLocalOnlyConfigWithoutServerState(t *testing.T) {
	originalLoadDeployConfig := loadDeployConfig
	originalLoadServerState := loadServerState
	originalRunDeploy := runDeploy
	defer func() {
		loadDeployConfig = originalLoadDeployConfig
		loadServerState = originalLoadServerState
		runDeploy = originalRunDeploy
	}()

	loadDeployConfig = func() (shipinternal.DeployConfig, error) {
		return shipinternal.DeployConfig{
			LocalCommands: []string{"npm run build"},
		}, nil
	}
	loadServerState = func() (shipinternal.ServerState, error) {
		t.Fatal("loadServerState should not be called for a local-only deploy")
		return shipinternal.ServerState{}, nil
	}

	var gotOpts shipinternal.Options
	runDeploy = func(ctx context.Context, opts shipinternal.Options) error {
		gotOpts = opts
		return nil
	}

	cmd := newDeployCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if gotOpts != (shipinternal.Options{}) {
		t.Fatalf("runDeploy got opts %+v, want zero-value opts", gotOpts)
	}
	output := stdout.String()
	if !strings.Contains(output, "STATUS=DEPLOY_COMPLETE") {
		t.Fatalf("command output = %q", output)
	}
	if strings.Contains(output, "SERVER_IP=") {
		t.Fatalf("command output unexpectedly contained SERVER_IP: %q", output)
	}
}
