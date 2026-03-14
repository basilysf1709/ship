package shipinternal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

func ApplyBootstrap(ctx context.Context, client *ssh.Client, config ProjectConfig) error {
	if config.Bootstrap != nil {
		if len(config.Bootstrap.Packages) > 0 {
			commands := []string{
				"apt update",
				fmt.Sprintf("apt install -y %s", strings.Join(config.Bootstrap.Packages, " ")),
			}
			if err := RunCommands(ctx, client, commands); err != nil {
				return err
			}
		}
		if len(config.Bootstrap.RemoteCommands) > 0 {
			if err := RunCommands(ctx, client, config.Bootstrap.RemoteCommands); err != nil {
				return err
			}
		}
	}

	if config.Proxy != nil && config.Proxy.HasDomains() {
		if err := ConfigureProxy(ctx, client, *config.Proxy); err != nil {
			return err
		}
	}

	return nil
}

func ConfigureProxy(ctx context.Context, client *ssh.Client, proxy ProxyConfig) error {
	if !proxy.HasDomains() {
		return nil
	}

	if err := RunCommands(ctx, client, []string{
		"apt update",
		"apt install -y caddy",
	}); err != nil {
		return err
	}

	if err := uploadTextFile(ctx, client, renderCaddyfile(proxy), "/tmp/ship.Caddyfile", 0o644); err != nil {
		return err
	}

	return RunCommands(ctx, client, []string{
		"mkdir -p /etc/caddy",
		"mv /tmp/ship.Caddyfile /etc/caddy/Caddyfile",
		"systemctl enable caddy",
		"systemctl restart caddy",
	})
}

func SyncSecretsToServer(ctx context.Context, client *ssh.Client) error {
	if !HasLocalSecrets() {
		return nil
	}
	if _, err := SecretsChecksum(); err != nil {
		return err
	}
	if err := RunCommands(ctx, client, []string{"mkdir -p /root/.ship"}); err != nil {
		return err
	}
	return CopyFile(ctx, client, secretsPath(), remoteSecretsPath, 0o600)
}

func renderCaddyfile(proxy ProxyConfig) string {
	return fmt.Sprintf("%s {\n\treverse_proxy 127.0.0.1:%d\n}\n", strings.Join(proxy.Domains, ", "), proxy.EffectiveAppPort())
}

func uploadTextFile(ctx context.Context, client *ssh.Client, content, remotePath string, mode os.FileMode) error {
	tempDir, err := os.MkdirTemp("", "ship-upload-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	localPath := filepath.Join(tempDir, filepath.Base(remotePath))
	if err := os.WriteFile(localPath, []byte(content), mode); err != nil {
		return fmt.Errorf("write temp file %s: %w", localPath, err)
	}
	return CopyFile(ctx, client, localPath, remotePath, mode)
}
