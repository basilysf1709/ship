package shipinternal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const releaseFile = "releases.json"
const remoteReleaseRoot = "/root/.ship/releases"

type ReleaseRecord struct {
	ID               string          `json:"id"`
	CreatedAt        string          `json:"created_at"`
	ProjectPath      string          `json:"project_path,omitempty"`
	ServerID         string          `json:"server_id,omitempty"`
	ServerIP         string          `json:"server_ip,omitempty"`
	GitSHA           string          `json:"git_sha,omitempty"`
	Status           string          `json:"status,omitempty"`
	Stage            string          `json:"stage,omitempty"`
	ErrorMessage     string          `json:"error_message,omitempty"`
	RollbackOf       string          `json:"rollback_of,omitempty"`
	RollbackEligible bool            `json:"rollback_eligible,omitempty"`
	RollbackReason   string          `json:"rollback_reason,omitempty"`
	DeployHash       string          `json:"deploy_hash,omitempty"`
	SecretsChecksum  string          `json:"secrets_checksum,omitempty"`
	RemoteCommands   []string        `json:"remote_commands,omitempty"`
	Uploads          []ReleaseUpload `json:"uploads,omitempty"`
}

type ReleaseUpload struct {
	Destination string `json:"destination"`
	BackupPath  string `json:"backup_path"`
}

func releaseHistoryPath() string {
	return filepath.Join(shipDir, releaseFile)
}

func releaseHistoryPathAt(projectPath string) string {
	return filepath.Join(projectPath, shipDir, releaseFile)
}

func ListReleaseHistory() ([]ReleaseRecord, error) {
	return listReleaseHistoryAtPath(releaseHistoryPath())
}

func ListReleaseHistoryAt(projectPath string) ([]ReleaseRecord, error) {
	return listReleaseHistoryAtPath(releaseHistoryPathAt(projectPath))
}

func listReleaseHistoryAtPath(path string) ([]ReleaseRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []ReleaseRecord{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var records []ReleaseRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	for i := range records {
		records[i].UpdateRollbackEligibility()
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})
	return records, nil
}

func SaveReleaseRecord(record ReleaseRecord) error {
	records, err := ListReleaseHistory()
	if err != nil {
		return err
	}
	records = append(records, record)
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})

	if err := os.MkdirAll(shipDir, 0o755); err != nil {
		return fmt.Errorf("create .ship directory: %w", err)
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal release history: %w", err)
	}
	if err := os.WriteFile(releaseHistoryPath(), data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", releaseHistoryPath(), err)
	}
	return nil
}

func FindReleaseRecord(releaseID string) (*ReleaseRecord, error) {
	records, err := ListReleaseHistory()
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.ID == releaseID {
			recordCopy := record
			return &recordCopy, nil
		}
	}
	return nil, fmt.Errorf("release %s not found", releaseID)
}

func LatestReleaseRecord() (*ReleaseRecord, error) {
	records, err := ListReleaseHistory()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	record := records[0]
	return &record, nil
}

func PreviousReleaseRecord() (*ReleaseRecord, error) {
	records, err := ListReleaseHistory()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}
	record := records[1]
	return &record, nil
}

func DefaultRollbackTargetRecord() (*ReleaseRecord, error) {
	records, err := ListReleaseHistory()
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(records); i++ {
		if !records[i].RollbackEligible {
			continue
		}
		record := records[i]
		return &record, nil
	}
	return nil, nil
}

func NewReleaseRecord(serverState *ServerState, deployConfig DeployConfig) (ReleaseRecord, error) {
	projectPath, _ := os.Getwd()
	deployHash, err := DeployConfigHash(deployConfig)
	if err != nil {
		return ReleaseRecord{}, err
	}
	secretsChecksum, err := SecretsChecksum()
	if err != nil {
		return ReleaseRecord{}, err
	}
	record := ReleaseRecord{
		ID:              time.Now().UTC().Format("20060102T150405.000000000Z"),
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		ProjectPath:     projectPath,
		GitSHA:          CurrentGitSHA(),
		Status:          "failed",
		Stage:           "initializing",
		DeployHash:      deployHash,
		SecretsChecksum: secretsChecksum,
		RemoteCommands:  append([]string(nil), deployConfig.RemoteCommands...),
	}
	if serverState != nil {
		record.ServerID = serverState.ServerID
		record.ServerIP = serverState.IP
	}
	return record, nil
}

func (r *ReleaseRecord) UpdateRollbackEligibility() {
	if r == nil {
		return
	}
	r.RollbackEligible = false
	r.RollbackReason = ""
	if r.ServerID == "" || r.ServerIP == "" {
		r.RollbackReason = "local-only release"
		return
	}
	if r.Status != "success" && r.Status != "rollback" {
		r.RollbackReason = "release did not complete successfully"
		return
	}
	if len(r.Uploads) == 0 {
		r.RollbackReason = "release has no tracked upload backups"
		return
	}
	for _, upload := range r.Uploads {
		if upload.Destination == "" || upload.BackupPath == "" {
			r.RollbackReason = "release is missing rollback backup metadata"
			return
		}
	}
	r.RollbackEligible = true
}

func DeployConfigHash(config DeployConfig) (string, error) {
	type hashUpload struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Mode        string `json:"mode"`
	}
	payload := struct {
		LocalCommands  []string     `json:"local_commands,omitempty"`
		Uploads        []hashUpload `json:"uploads,omitempty"`
		RemoteCommands []string     `json:"remote_commands,omitempty"`
		CleanupLocal   []string     `json:"cleanup_local,omitempty"`
	}{
		LocalCommands:  append([]string(nil), config.LocalCommands...),
		RemoteCommands: append([]string(nil), config.RemoteCommands...),
		CleanupLocal:   append([]string(nil), config.CleanupLocal...),
	}

	for _, upload := range config.Uploads {
		mode, err := parseDeployFileMode(upload.Mode)
		if err != nil {
			return "", err
		}
		payload.Uploads = append(payload.Uploads, hashUpload{
			Source:      upload.Source,
			Destination: upload.Destination,
			Mode:        fmt.Sprintf("%04o", mode.Perm()),
		})
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal deploy config hash: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func CurrentGitSHA() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
