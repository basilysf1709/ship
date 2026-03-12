package shipinternal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func WaitForSSH(ctx context.Context, user, host string, interval time.Duration) (*ssh.Client, error) {
	config, err := clientConfig(user)
	if err != nil {
		return nil, err
	}

	address := net.JoinHostPort(host, "22")
	const maxAuthRetries = 5
	authRetries := 0
	for {
		client, err := ssh.Dial("tcp", address, config)
		if err == nil {
			return client, nil
		}
		if isSSHAuthError(err) {
			authRetries++
			if authRetries >= maxAuthRetries {
				return nil, fmt.Errorf("SSH authentication failed for %s@%s; ensure this machine's SSH key is registered with the provider: %w", user, host, err)
			}
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for SSH on %s: %w", host, ctx.Err())
		case <-time.After(interval):
		}
	}
}

func isSSHAuthError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ssh.ErrNoAuth) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unable to authenticate") || strings.Contains(msg, "permission denied")
}

func RunCommands(ctx context.Context, client *ssh.Client, commands []string) error {
	for _, command := range commands {
		if _, err := RunCommand(ctx, client, command); err != nil {
			return err
		}
	}
	return nil
}

func RunCommand(ctx context.Context, client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		_ = session.Close()
		return "", fmt.Errorf("run remote command: %w", ctx.Err())
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return "", fmt.Errorf("run remote command %q: %s", command, bytes.TrimSpace(stderr.Bytes()))
			}
			return "", fmt.Errorf("run remote command %q: %w", command, err)
		}
	}

	if stderr.Len() > 0 && stdout.Len() == 0 {
		return stderr.String(), nil
	}
	return stdout.String(), nil
}

func CopyFile(ctx context.Context, client *ssh.Client, localPath, remotePath string, mode os.FileMode) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", localPath, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", localPath, err)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("open SCP stdin: %w", err)
	}

	var stderr bytes.Buffer
	session.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- session.Run(fmt.Sprintf("scp -t %s", shellQuote(remotePath)))
	}()

	go func() {
		defer stdin.Close()
		fmt.Fprintf(stdin, "C%04o %d %s\n", mode.Perm(), info.Size(), filepath.Base(remotePath))
		_, _ = io.Copy(stdin, file)
		fmt.Fprint(stdin, "\x00")
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("copy file to server: %w", ctx.Err())
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return fmt.Errorf("copy file to server: %s", bytes.TrimSpace(stderr.Bytes()))
			}
			return fmt.Errorf("copy file to server: %w", err)
		}
	}
	return nil
}

func clientConfig(user string) (*ssh.ClientConfig, error) {
	authMethods, err := authMethods()
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}, nil
}

func authMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			agentClient := agent.NewClient(conn)
			signers, err := agentClient.Signers()
			if err == nil && len(signers) > 0 {
				methods = append(methods, ssh.PublicKeysCallback(agentClient.Signers))
			}
		}
	}

	for _, path := range []string{
		"~/.ssh/id_ed25519",
		"~/.ssh/id_rsa",
	} {
		signer, err := signerFromFile(expandHome(path))
		if err == nil {
			methods = append(methods, ssh.PublicKeys(signer))
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no SSH authentication method available; ensure SSH agent or default keys are configured")
	}
	return methods, nil
}

func signerFromFile(path string) (ssh.Signer, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(key)
}

func expandHome(path string) string {
	if len(path) < 2 || path[:2] != "~/" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
