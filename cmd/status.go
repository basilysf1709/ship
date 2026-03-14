package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

type statusOutput struct {
	ServerID      string `json:"server_id,omitempty"`
	ServerIP      string `json:"server_ip,omitempty"`
	Provider      string `json:"provider,omitempty"`
	SSHReachable  bool   `json:"ssh_reachable"`
	AppStatus     string `json:"app_status,omitempty"`
	HealthURL     string `json:"health_url,omitempty"`
	HealthStatus  string `json:"health_status,omitempty"`
	LastReleaseID string `json:"last_release_id,omitempty"`
	LastReleaseAt string `json:"last_release_at,omitempty"`
	GitSHA        string `json:"git_sha,omitempty"`
}

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show server and app status",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := loadCurrentServerState()
			if err != nil {
				return err
			}

			result := statusOutput{
				ServerID: state.ServerID,
				ServerIP: state.IP,
				Provider: state.Provider,
			}

			if release, err := latestReleaseRecord(); err == nil && release != nil {
				result.LastReleaseID = release.ID
				result.LastReleaseAt = release.CreatedAt
				result.GitSHA = release.GitSHA
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()

			client, err := waitForSSH(ctx, state.EffectiveSSHUser(), state.IP, 5*time.Second)
			if err == nil {
				result.SSHReachable = true
				defer client.Close()
				output, runErr := runRemoteCommand(ctx, client, "docker ps --filter name=^/app$ --format '{{.Status}}'")
				if runErr == nil {
					result.AppStatus = strings.TrimSpace(output)
				}
			}

			projectConfig, err := loadProjectConfig()
			if err == nil {
				result.HealthURL = buildHealthURL(projectConfig, state.IP)
			}
			if result.HealthURL != "" {
				result.HealthStatus = checkHealthURL(result.HealthURL)
			}

			text := fmt.Sprintf(
				"SERVER_ID=%s\nSERVER_IP=%s\nPROVIDER=%s\nSSH_REACHABLE=%t\nAPP_STATUS=%s\nHEALTH_URL=%s\nHEALTH_STATUS=%s\nLAST_RELEASE_ID=%s\nLAST_RELEASE_AT=%s\nGIT_SHA=%s\n",
				result.ServerID,
				result.ServerIP,
				result.Provider,
				result.SSHReachable,
				result.AppStatus,
				result.HealthURL,
				result.HealthStatus,
				result.LastReleaseID,
				result.LastReleaseAt,
				result.GitSHA,
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
