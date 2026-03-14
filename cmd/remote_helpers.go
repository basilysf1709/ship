package cmd

import (
	"context"
	"time"

	shipinternal "ship/internal"

	"golang.org/x/crypto/ssh"
)

var waitForSSH = shipinternal.WaitForSSH
var runRemoteCommand = shipinternal.RunCommand
var runRemoteCommands = shipinternal.RunCommands
var applyBootstrap = shipinternal.ApplyBootstrap
var configureProxy = shipinternal.ConfigureProxy
var syncSecretsToServer = shipinternal.SyncSecretsToServer
var latestReleaseRecord = shipinternal.LatestReleaseRecord
var previousReleaseRecord = shipinternal.PreviousReleaseRecord
var findReleaseRecord = shipinternal.FindReleaseRecord
var saveReleaseRecord = shipinternal.SaveReleaseRecord
var listReleaseHistory = shipinternal.ListReleaseHistory
var listReleaseHistoryAt = shipinternal.ListReleaseHistoryAt
var loadProjectConfig = shipinternal.LoadProjectConfig
var loadCurrentServerState = shipinternal.LoadServerState

func currentServerClient(ctx context.Context, timeout time.Duration) (shipinternal.ServerState, *ssh.Client, error) {
	state, err := loadCurrentServerState()
	if err != nil {
		return shipinternal.ServerState{}, nil, err
	}

	childCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client, err := waitForSSH(childCtx, state.EffectiveSSHUser(), state.IP, 10*time.Second)
	if err != nil {
		return shipinternal.ServerState{}, nil, err
	}
	return state, client, nil
}
