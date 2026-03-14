package shipinternal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const projectConfigFile = "ship.json"

type ProjectConfig struct {
	Deploy    *DeployConfig    `json:"deploy,omitempty"`
	Bootstrap *BootstrapConfig `json:"bootstrap,omitempty"`
	Proxy     *ProxyConfig     `json:"proxy,omitempty"`
	Status    *StatusConfig    `json:"status,omitempty"`
}

type DeployConfig struct {
	LocalCommands  []string       `json:"local_commands,omitempty"`
	Uploads        []DeployUpload `json:"uploads,omitempty"`
	RemoteCommands []string       `json:"remote_commands,omitempty"`
	CleanupLocal   []string       `json:"cleanup_local,omitempty"`
}

type BootstrapConfig struct {
	Packages       []string `json:"packages,omitempty"`
	RemoteCommands []string `json:"remote_commands,omitempty"`
}

type ProxyConfig struct {
	Domains []string `json:"domains,omitempty"`
	AppPort int      `json:"app_port,omitempty"`
}

type StatusConfig struct {
	HealthcheckURL  string `json:"healthcheck_url,omitempty"`
	HealthcheckPath string `json:"healthcheck_path,omitempty"`
}

type DeployUpload struct {
	Source      string      `json:"source"`
	Destination string      `json:"destination"`
	Mode        interface{} `json:"mode,omitempty"`
}

type ResolvedDeployUpload struct {
	Source      string
	Destination string
	Mode        os.FileMode
}

func (c DeployConfig) RequiresServer() bool {
	return len(c.Uploads) > 0 || len(c.RemoteCommands) > 0
}

func (c ProxyConfig) EffectiveAppPort() int {
	if c.AppPort <= 0 {
		return 8080
	}
	return c.AppPort
}

func (c ProxyConfig) HasDomains() bool {
	return len(c.Domains) > 0
}

func LoadProjectConfig() (ProjectConfig, error) {
	var projectConfig ProjectConfig
	data, err := os.ReadFile(projectConfigFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return ProjectConfig{}, fmt.Errorf("read %s: %w", projectConfigFile, err)
		}
	} else if err := json.Unmarshal(data, &projectConfig); err != nil {
		return ProjectConfig{}, fmt.Errorf("parse %s: %w", projectConfigFile, err)
	}

	if err := projectConfig.validate(); err != nil {
		return ProjectConfig{}, err
	}
	runtimeConfig, err := LoadRuntimeConfig()
	if err != nil {
		return ProjectConfig{}, err
	}
	if projectConfig.Proxy == nil && runtimeConfig.Proxy != nil {
		proxyCopy := *runtimeConfig.Proxy
		projectConfig.Proxy = &proxyCopy
	}

	return projectConfig, nil
}

func LoadDeployConfig() (DeployConfig, error) {
	projectConfig, err := LoadProjectConfig()
	if err != nil {
		return DeployConfig{}, err
	}
	return projectConfig.EffectiveDeployConfig()
}

func (c ProjectConfig) EffectiveDeployConfig() (DeployConfig, error) {
	if c.Deploy != nil {
		return *c.Deploy, nil
	}
	return defaultDeployConfig(c.Proxy), nil
}

func defaultDeployConfig(proxy *ProxyConfig) DeployConfig {
	portBinding := "80:80"
	if proxy != nil && proxy.HasDomains() {
		portBinding = fmt.Sprintf("127.0.0.1:%d:80", proxy.EffectiveAppPort())
	}
	return DeployConfig{
		LocalCommands: []string{
			"docker build -t app .",
			"docker save app -o app.tar",
		},
		Uploads: []DeployUpload{
			{
				Source:      "app.tar",
				Destination: "/root/app.tar",
				Mode:        "0644",
			},
		},
		RemoteCommands: []string{
			"docker load -i /root/app.tar",
			"docker stop app || true",
			"docker rm app || true",
			fmt.Sprintf("if [ -f /root/.ship/secrets.env ]; then docker run -d --name app --env-file /root/.ship/secrets.env -p %s app; else docker run -d --name app -p %s app; fi", portBinding, portBinding),
		},
		CleanupLocal: []string{"app.tar"},
	}
}

func (c DeployConfig) ResolvedUploads(baseDir string) ([]ResolvedDeployUpload, error) {
	resolved := make([]ResolvedDeployUpload, 0, len(c.Uploads))
	for _, upload := range c.Uploads {
		mode, err := parseDeployFileMode(upload.Mode)
		if err != nil {
			return nil, fmt.Errorf("invalid upload mode for %s: %w", upload.Source, err)
		}
		source := upload.Source
		if !filepath.IsAbs(source) {
			source = filepath.Join(baseDir, source)
		}
		resolved = append(resolved, ResolvedDeployUpload{
			Source:      source,
			Destination: upload.Destination,
			Mode:        mode,
		})
	}
	return resolved, nil
}

func (c DeployConfig) ResolvedCleanupPaths(baseDir string) []string {
	paths := make([]string, 0, len(c.CleanupLocal))
	for _, path := range c.CleanupLocal {
		if filepath.IsAbs(path) {
			paths = append(paths, path)
			continue
		}
		paths = append(paths, filepath.Join(baseDir, path))
	}
	return paths
}

func (c DeployConfig) validate() error {
	if len(c.LocalCommands) == 0 && len(c.Uploads) == 0 && len(c.RemoteCommands) == 0 {
		return fmt.Errorf("%s deploy config must define at least one local command, upload, or remote command", projectConfigFile)
	}

	for _, upload := range c.Uploads {
		if upload.Source == "" {
			return fmt.Errorf("%s deploy uploads require source", projectConfigFile)
		}
		if upload.Destination == "" {
			return fmt.Errorf("%s deploy uploads require destination", projectConfigFile)
		}
		if _, err := parseDeployFileMode(upload.Mode); err != nil {
			return fmt.Errorf("%s deploy upload %s has invalid mode: %w", projectConfigFile, upload.Source, err)
		}
	}

	return nil
}

func (c ProjectConfig) validate() error {
	if c.Deploy != nil {
		if err := c.Deploy.validate(); err != nil {
			return err
		}
	}
	if c.Proxy != nil {
		if c.Proxy.AppPort < 0 {
			return fmt.Errorf("%s proxy app_port must be greater than zero", projectConfigFile)
		}
		for _, domain := range c.Proxy.Domains {
			if strings.TrimSpace(domain) == "" {
				return fmt.Errorf("%s proxy domains cannot contain empty values", projectConfigFile)
			}
		}
	}
	return nil
}

func parseDeployFileMode(value interface{}) (os.FileMode, error) {
	if value == nil {
		return 0o644, nil
	}

	switch mode := value.(type) {
	case string:
		parsed, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return 0, fmt.Errorf("parse mode %q: %w", mode, err)
		}
		return os.FileMode(parsed), nil
	default:
		return 0, fmt.Errorf("unsupported mode type %T; use a quoted octal string like 0644", value)
	}
}
