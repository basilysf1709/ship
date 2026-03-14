package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"

	shipinternal "ship/internal"
)

type statusOutput struct {
	ServerID             string `json:"server_id,omitempty"`
	ServerIP             string `json:"server_ip,omitempty"`
	Provider             string `json:"provider,omitempty"`
	SSHReachable         bool   `json:"ssh_reachable"`
	SSHStatus            string `json:"ssh_status,omitempty"`
	AppStatus            string `json:"app_status,omitempty"`
	AppHealth            string `json:"app_health,omitempty"`
	ProxyStatus          string `json:"proxy_status,omitempty"`
	HealthURL            string `json:"health_url,omitempty"`
	HealthStatus         string `json:"health_status,omitempty"`
	LastReleaseID        string `json:"last_release_id,omitempty"`
	LastReleaseAt        string `json:"last_release_at,omitempty"`
	LastReleaseStatus    string `json:"last_release_status,omitempty"`
	LastReleaseStage     string `json:"last_release_stage,omitempty"`
	GitSHA               string `json:"git_sha,omitempty"`
	CurrentGitSHA        string `json:"current_git_sha,omitempty"`
	GitDrift             bool   `json:"git_drift"`
	DeployDrift          bool   `json:"deploy_drift"`
	SecretsStatus        string `json:"secrets_status,omitempty"`
	LocalSecretsPresent  bool   `json:"local_secrets_present"`
	RemoteSecretsPresent bool   `json:"remote_secrets_present"`
	RollbackStatus       string `json:"rollback_status,omitempty"`
}

var checkHealthURLFunc = checkHealthURL

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show server, release, and drift status",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := loadCurrentServerState()
			if err != nil {
				return err
			}

			result := statusOutput{
				ServerID:  state.ServerID,
				ServerIP:  state.IP,
				Provider:  state.Provider,
				SSHStatus: "unreachable",
			}

			release, _ := latestReleaseRecord()
			if release != nil {
				result.LastReleaseID = release.ID
				result.LastReleaseAt = release.CreatedAt
				result.LastReleaseStatus = release.Status
				result.LastReleaseStage = release.Stage
				result.GitSHA = release.GitSHA
				if release.RollbackEligible {
					result.RollbackStatus = "available"
				} else if release.RollbackReason != "" {
					result.RollbackStatus = release.RollbackReason
				} else {
					result.RollbackStatus = "not_available"
				}
			}

			result.CurrentGitSHA = currentGitSHA()
			result.GitDrift = release != nil &&
				release.GitSHA != "" &&
				result.CurrentGitSHA != "" &&
				release.GitSHA != result.CurrentGitSHA

			if deployConfig, err := loadDeployConfig(); err == nil && release != nil && release.DeployHash != "" {
				if currentHash, hashErr := deployConfigHash(deployConfig); hashErr == nil && currentHash != "" {
					result.DeployDrift = currentHash != release.DeployHash
				}
			}

			localSecretsChecksum := ""
			if checksum, err := currentSecretsChecksum(); err == nil {
				localSecretsChecksum = checksum
				result.LocalSecretsPresent = checksum != ""
			} else {
				result.SecretsStatus = "invalid_local"
			}

			projectConfig, err := loadProjectConfig()
			if err == nil {
				result.HealthURL = buildHealthURL(projectConfig, state.IP)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Second)
			defer cancel()

			client, err := waitForSSH(ctx, state.EffectiveSSHUser(), state.IP, 5*time.Second)
			if err == nil {
				result.SSHReachable = true
				result.SSHStatus = "reachable"
				if client != nil {
					defer client.Close()
				}

				result.AppStatus = statusRemoteTrim(ctx, client, "docker inspect -f '{{.State.Status}}' app 2>/dev/null || true")
				result.AppHealth = statusRemoteTrim(ctx, client, "docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{end}}' app 2>/dev/null || true")
				if projectConfig.Proxy != nil && projectConfig.Proxy.HasDomains() {
					result.ProxyStatus = statusRemoteTrim(ctx, client, "systemctl is-active caddy 2>/dev/null || true")
					if result.ProxyStatus == "" {
						result.ProxyStatus = "unknown"
					}
				}

				remoteSecretsChecksum := statusRemoteTrim(ctx, client, "if [ -f /root/.ship/secrets.env ]; then grep -v '^[[:space:]]*#' /root/.ship/secrets.env | sed '/^[[:space:]]*$/d' | LC_ALL=C sort | sha256sum | awk '{print $1}'; fi")
				result.RemoteSecretsPresent = remoteSecretsChecksum != ""
				if result.SecretsStatus == "" {
					result.SecretsStatus = summarizeSecretsStatus(localSecretsChecksum, remoteSecretsChecksum)
				}
			} else if result.SecretsStatus == "" {
				result.SecretsStatus = summarizeSecretsStatus(localSecretsChecksum, "")
			}

			if result.HealthURL != "" {
				result.HealthStatus = checkHealthURLFunc(result.HealthURL)
			}
			if result.SecretsStatus == "" {
				result.SecretsStatus = "not_configured"
			}

			text := fmt.Sprintf(
				"SERVER_ID=%s\nSERVER_IP=%s\nPROVIDER=%s\nSSH_REACHABLE=%t\nSSH_STATUS=%s\nAPP_STATUS=%s\nAPP_HEALTH=%s\nPROXY_STATUS=%s\nHEALTH_URL=%s\nHEALTH_STATUS=%s\nLAST_RELEASE_ID=%s\nLAST_RELEASE_AT=%s\nLAST_RELEASE_STATUS=%s\nLAST_RELEASE_STAGE=%s\nGIT_SHA=%s\nCURRENT_GIT_SHA=%s\nGIT_DRIFT=%t\nDEPLOY_DRIFT=%t\nSECRETS_STATUS=%s\nLOCAL_SECRETS_PRESENT=%t\nREMOTE_SECRETS_PRESENT=%t\nROLLBACK_STATUS=%s\n",
				result.ServerID,
				result.ServerIP,
				result.Provider,
				result.SSHReachable,
				result.SSHStatus,
				result.AppStatus,
				result.AppHealth,
				result.ProxyStatus,
				result.HealthURL,
				result.HealthStatus,
				result.LastReleaseID,
				result.LastReleaseAt,
				result.LastReleaseStatus,
				result.LastReleaseStage,
				result.GitSHA,
				result.CurrentGitSHA,
				result.GitDrift,
				result.DeployDrift,
				result.SecretsStatus,
				result.LocalSecretsPresent,
				result.RemoteSecretsPresent,
				result.RollbackStatus,
			)
			return writeCommandOutput(cmd, text, result)
		},
	}
}

func buildHealthURL(config shipinternal.ProjectConfig, ip string) string {
	statusConfig := config.Status
	proxyConfig := config.Proxy
	if statusConfig != nil && statusConfig.HealthcheckURL != "" {
		return statusConfig.HealthcheckURL
	}

	base := "http://" + ip
	if proxyConfig != nil && len(proxyConfig.Domains) > 0 {
		base = "https://" + proxyConfig.Domains[0]
	}
	path := "/"
	if statusConfig != nil && statusConfig.HealthcheckPath != "" {
		path = statusConfig.HealthcheckPath
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "http") {
		path = "/" + path
	}
	return base + path
}

func checkHealthURL(url string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "unreachable"
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return fmt.Sprintf("ok:%d", resp.StatusCode)
	}
	return fmt.Sprintf("error:%d", resp.StatusCode)
}

func statusRemoteTrim(ctx context.Context, client *ssh.Client, command string) string {
	output, err := runRemoteCommand(ctx, client, command)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

func summarizeSecretsStatus(localChecksum, remoteChecksum string) string {
	switch {
	case localChecksum == "" && remoteChecksum == "":
		return "not_configured"
	case localChecksum != "" && remoteChecksum == "":
		return "local_only"
	case localChecksum == "" && remoteChecksum != "":
		return "remote_only"
	case localChecksum == remoteChecksum:
		return "synced"
	default:
		return "drifted"
	}
}
