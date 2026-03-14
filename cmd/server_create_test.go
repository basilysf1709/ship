package cmd

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	shipinternal "ship/internal"

	"golang.org/x/crypto/ssh"
)

type fakeProvider struct {
	state shipinternal.ServerState
}

func (p fakeProvider) CreateServer(ctx context.Context, req shipinternal.CreateRequest) (shipinternal.ServerState, error) {
	return p.state, nil
}

func (p fakeProvider) DestroyServer(ctx context.Context, state shipinternal.ServerState) error {
	return nil
}

func TestServerCreateTracksServerBeforeBootstrapFailure(t *testing.T) {
	originalCreateProvider := createProvider
	originalWaitForCreatedServerSSH := waitForCreatedServerSSH
	originalRunCreatedServerSetup := runCreatedServerSetup
	originalSaveCreatedServerState := saveCreatedServerState
	originalAddCreatedServerInventory := addCreatedServerInventory
	originalLoadProjectConfig := loadProjectConfig
	originalApplyBootstrap := applyBootstrap
	defer func() {
		createProvider = originalCreateProvider
		waitForCreatedServerSSH = originalWaitForCreatedServerSSH
		runCreatedServerSetup = originalRunCreatedServerSetup
		saveCreatedServerState = originalSaveCreatedServerState
		addCreatedServerInventory = originalAddCreatedServerInventory
		loadProjectConfig = originalLoadProjectConfig
		applyBootstrap = originalApplyBootstrap
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

	var order []string
	createProvider = func(name string) (shipinternal.Provider, error) {
		return fakeProvider{state: shipinternal.ServerState{
			Provider: "digitalocean",
			ServerID: "srv-1",
			IP:       "1.2.3.4",
			SSHUser:  "root",
		}}, nil
	}
	saveCreatedServerState = func(state shipinternal.ServerState) error {
		order = append(order, "save")
		return nil
	}
	addCreatedServerInventory = func(state shipinternal.ServerState, projectPath string) error {
		order = append(order, "inventory")
		return nil
	}
	waitForCreatedServerSSH = func(ctx context.Context, user, host string, interval time.Duration) (*ssh.Client, error) {
		order = append(order, "ssh")
		return nil, nil
	}
	runCreatedServerSetup = func(ctx context.Context, client *ssh.Client, commands []string) error {
		order = append(order, "setup")
		return nil
	}
	loadProjectConfig = func() (shipinternal.ProjectConfig, error) {
		order = append(order, "config")
		return shipinternal.ProjectConfig{}, nil
	}
	applyBootstrap = func(ctx context.Context, client *ssh.Client, config shipinternal.ProjectConfig) error {
		order = append(order, "bootstrap")
		return errors.New("bootstrap failed")
	}

	cmd := newServerCreateCommand()
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute returned nil error")
	}

	want := []string{"save", "inventory", "ssh", "setup", "config", "bootstrap"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %+v, want %+v", order, want)
	}
}
