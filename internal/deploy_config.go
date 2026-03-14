package shipinternal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const projectConfigFile = "ship.json"

type ProjectConfig struct {
	Deploy *DeployConfig `json:"deploy,omitempty"`
}

type DeployConfig struct {
	LocalCommands  []string       `json:"local_commands,omitempty"`
	Uploads        []DeployUpload `json:"uploads,omitempty"`
	RemoteCommands []string       `json:"remote_commands,omitempty"`
	CleanupLocal   []string       `json:"cleanup_local,omitempty"`
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

func LoadDeployConfig() (DeployConfig, error) {
	data, err := os.ReadFile(projectConfigFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultDeployConfig(), nil
		}
		return DeployConfig{}, fmt.Errorf("read %s: %w", projectConfigFile, err)
	}

	var projectConfig ProjectConfig
	if err := json.Unmarshal(data, &projectConfig); err != nil {
		return DeployConfig{}, fmt.Errorf("parse %s: %w", projectConfigFile, err)
	}

	if projectConfig.Deploy == nil {
		return defaultDeployConfig(), nil
	}

	if err := projectConfig.Deploy.validate(); err != nil {
		return DeployConfig{}, err
	}

	return *projectConfig.Deploy, nil
}

func defaultDeployConfig() DeployConfig {
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
			"docker run -d --name app -p 80:80 app",
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
