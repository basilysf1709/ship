package shipinternal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const shipDir = ".ship"
const serverFile = "server.json"
const inventoryFile = "servers.json"
const runtimeFile = "runtime.json"

var userHomeDir = os.UserHomeDir

type ServerState struct {
	Provider string `json:"provider,omitempty"`
	ServerID string `json:"server_id"`
	IP       string `json:"ip"`
	SSHUser  string `json:"ssh_user,omitempty"`
}

type ServerRecord struct {
	Provider    string `json:"provider,omitempty"`
	ServerID    string `json:"server_id"`
	IP          string `json:"ip"`
	SSHUser     string `json:"ssh_user,omitempty"`
	ProjectPath string `json:"project_path,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type RuntimeConfig struct {
	Proxy *ProxyConfig `json:"proxy,omitempty"`
}

func serverStatePath() string {
	return filepath.Join(shipDir, serverFile)
}

func inventoryPath() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, shipDir, inventoryFile), nil
}

func runtimeConfigPath() string {
	return filepath.Join(shipDir, runtimeFile)
}

func SaveServerState(state ServerState) error {
	if err := os.MkdirAll(shipDir, 0o755); err != nil {
		return fmt.Errorf("create .ship directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal server state: %w", err)
	}

	if err := os.WriteFile(serverStatePath(), data, 0o600); err != nil {
		return fmt.Errorf("write .ship/server.json: %w", err)
	}
	return nil
}

func LoadServerState() (ServerState, error) {
	data, err := os.ReadFile(serverStatePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ServerState{}, errors.New("missing .ship/server.json; create a server first")
		}
		return ServerState{}, fmt.Errorf("read .ship/server.json: %w", err)
	}

	var state ServerState
	if err := json.Unmarshal(data, &state); err != nil {
		return ServerState{}, fmt.Errorf("parse .ship/server.json: %w", err)
	}
	if state.ServerID == "" || state.IP == "" {
		return ServerState{}, errors.New("invalid .ship/server.json: server_id and ip are required")
	}
	if state.Provider == "" {
		state.Provider = "digitalocean"
	}
	if state.SSHUser == "" {
		state.SSHUser = "root"
	}
	return state, nil
}

func DeleteServerState() error {
	if err := os.Remove(serverStatePath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove .ship/server.json: %w", err)
	}
	return nil
}

func (s ServerState) EffectiveSSHUser() string {
	if s.SSHUser == "" {
		return "root"
	}
	return s.SSHUser
}

func (s ServerState) Link() string {
	if s.IP == "" {
		return ""
	}
	return "http://" + s.IP
}

func (r ServerRecord) Link() string {
	if r.IP == "" {
		return ""
	}
	return "http://" + r.IP
}

func ListServerInventory() ([]ServerRecord, error) {
	path, err := inventoryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []ServerRecord{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var records []ServerRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})
	return records, nil
}

func AddServerInventoryRecord(state ServerState, projectPath string) error {
	records, err := ListServerInventory()
	if err != nil {
		return err
	}

	record := ServerRecord{
		Provider:    state.Provider,
		ServerID:    state.ServerID,
		IP:          state.IP,
		SSHUser:     state.EffectiveSSHUser(),
		ProjectPath: projectPath,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	replaced := false
	for i := range records {
		if records[i].Provider == record.Provider && records[i].ServerID == record.ServerID {
			records[i] = record
			replaced = true
			break
		}
	}
	if !replaced {
		records = append(records, record)
	}

	return saveServerInventory(records)
}

func RemoveServerInventoryRecord(state ServerState) error {
	records, err := ListServerInventory()
	if err != nil {
		return err
	}

	filtered := records[:0]
	for _, record := range records {
		if record.Provider == state.Provider && record.ServerID == state.ServerID {
			continue
		}
		filtered = append(filtered, record)
	}

	return saveServerInventory(filtered)
}

func saveServerInventory(records []ServerRecord) error {
	path, err := inventoryPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s directory: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal server inventory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func LoadRuntimeConfig() (RuntimeConfig, error) {
	data, err := os.ReadFile(runtimeConfigPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return RuntimeConfig{}, nil
		}
		return RuntimeConfig{}, fmt.Errorf("read %s: %w", runtimeConfigPath(), err)
	}

	var config RuntimeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return RuntimeConfig{}, fmt.Errorf("parse %s: %w", runtimeConfigPath(), err)
	}
	return config, nil
}

func SaveRuntimeConfig(config RuntimeConfig) error {
	if err := os.MkdirAll(shipDir, 0o755); err != nil {
		return fmt.Errorf("create .ship directory: %w", err)
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal runtime config: %w", err)
	}
	if err := os.WriteFile(runtimeConfigPath(), data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", runtimeConfigPath(), err)
	}
	return nil
}

func SaveProxyRuntimeConfig(proxy ProxyConfig) error {
	runtimeConfig, err := LoadRuntimeConfig()
	if err != nil {
		return err
	}
	proxyCopy := proxy
	runtimeConfig.Proxy = &proxyCopy
	return SaveRuntimeConfig(runtimeConfig)
}
