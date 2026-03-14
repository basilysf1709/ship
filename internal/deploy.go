package shipinternal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Options struct {
	ServerIP string
	ServerID string
	User     string
}

var loadDeployConfig = LoadDeployConfig
var runLocalCommand = runLocalShellCommand
var waitForSSHClient = WaitForSSH
var copyDeployFile = CopyFile
var runRemoteCommands = RunCommands
var closeSSHClient = func(client *ssh.Client) error {
	return client.Close()
}

func Run(ctx context.Context, opts Options) error {
	deployConfig, err := loadDeployConfig()
	if err != nil {
		return err
	}

	cleanupPaths := deployConfig.ResolvedCleanupPaths(".")
	preexistingCleanupPaths, err := existingPaths(cleanupPaths)
	if err != nil {
		return err
	}
	scheduledCleanupPaths := make(map[string]bool, len(cleanupPaths))

	for _, command := range deployConfig.LocalCommands {
		if err := runLocalCommand(ctx, command); err != nil {
			return err
		}
		for _, cleanupPath := range cleanupCandidates(cleanupPaths, preexistingCleanupPaths, scheduledCleanupPaths) {
			scheduledCleanupPaths[cleanupPath] = true
			defer os.Remove(cleanupPath)
		}
	}

	uploads, err := deployConfig.ResolvedUploads(".")
	if err != nil {
		return err
	}
	if opts.ServerIP != "" && HasLocalSecrets() {
		uploads = append(uploads, ResolvedDeployUpload{
			Source:      secretsPath(),
			Destination: remoteSecretsPath,
			Mode:        0o600,
		})
	}

	if len(uploads) == 0 && len(deployConfig.RemoteCommands) == 0 {
		return SaveReleaseRecord(NewReleaseRecord(serverStateFromOptions(opts), nil, nil))
	}

	client, err := waitForSSHClient(ctx, opts.User, opts.ServerIP, 10*time.Second)
	if err != nil {
		return err
	}
	defer closeSSHClient(client)

	releaseRecord := NewReleaseRecord(serverStateFromOptions(opts), deployConfig.RemoteCommands, nil)
	releaseDir := path.Join(remoteReleaseRoot, releaseRecord.ID)

	preCopyCommands := remoteUploadMkdirCommands(uploads)
	if len(uploads) > 0 {
		preCopyCommands = append(preCopyCommands, fmt.Sprintf("mkdir -p %s", shellQuote(releaseDir)))
	}
	if err := runRemoteCommands(ctx, client, preCopyCommands); err != nil {
		return err
	}

	releaseUploads := make([]ReleaseUpload, 0, len(uploads))
	for _, upload := range uploads {
		if err := copyDeployFile(ctx, client, upload.Source, upload.Destination, upload.Mode); err != nil {
			return err
		}
		if upload.Destination == remoteSecretsPath {
			continue
		}
		backupPath := path.Join(releaseDir, releaseBackupName(len(releaseUploads), upload.Destination))
		if err := runRemoteCommands(ctx, client, []string{
			fmt.Sprintf("cp %s %s", shellQuote(upload.Destination), shellQuote(backupPath)),
		}); err != nil {
			return err
		}
		releaseUploads = append(releaseUploads, ReleaseUpload{
			Destination: upload.Destination,
			BackupPath:  backupPath,
		})
	}

	if err := runRemoteCommands(ctx, client, deployConfig.RemoteCommands); err != nil {
		return err
	}

	releaseRecord.Uploads = releaseUploads
	return SaveReleaseRecord(releaseRecord)
}

func runLocalShellCommand(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s: %w", strings.TrimSpace(command), err)
	}
	return nil
}

func existingPaths(paths []string) (map[string]bool, error) {
	existing := make(map[string]bool, len(paths))
	for _, path := range paths {
		_, err := os.Lstat(path)
		if err == nil {
			existing[path] = true
			continue
		}
		if os.IsNotExist(err) {
			existing[path] = false
			continue
		}
		return nil, fmt.Errorf("stat cleanup path %s: %w", path, err)
	}
	return existing, nil
}

func cleanupCandidates(paths []string, preexisting map[string]bool, scheduled map[string]bool) []string {
	var cleanup []string
	for _, cleanupPath := range paths {
		if preexisting[cleanupPath] {
			continue
		}
		if scheduled[cleanupPath] {
			continue
		}
		if _, err := os.Lstat(cleanupPath); err == nil {
			cleanup = append(cleanup, cleanupPath)
		}
	}
	return cleanup
}

func remoteUploadMkdirCommands(uploads []ResolvedDeployUpload) []string {
	var commands []string
	seen := make(map[string]struct{}, len(uploads))
	for _, upload := range uploads {
		parent := path.Dir(upload.Destination)
		if parent == "." || parent == "/" || parent == "" {
			continue
		}
		if _, ok := seen[parent]; ok {
			continue
		}
		seen[parent] = struct{}{}
		commands = append(commands, fmt.Sprintf("mkdir -p %s", shellQuote(parent)))
	}
	return commands
}

func serverStateFromOptions(opts Options) *ServerState {
	if opts.ServerIP == "" && opts.ServerID == "" {
		return nil
	}
	return &ServerState{
		ServerID: opts.ServerID,
		IP:       opts.ServerIP,
		SSHUser:  opts.User,
	}
}

func releaseBackupName(index int, destination string) string {
	return fmt.Sprintf("%02d-%s", index, path.Base(destination))
}
