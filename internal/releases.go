package shipinternal

import (
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
	ID             string          `json:"id"`
	CreatedAt      string          `json:"created_at"`
	ProjectPath    string          `json:"project_path,omitempty"`
	ServerID       string          `json:"server_id,omitempty"`
	ServerIP       string          `json:"server_ip,omitempty"`
	GitSHA         string          `json:"git_sha,omitempty"`
	Status         string          `json:"status,omitempty"`
	RollbackOf     string          `json:"rollback_of,omitempty"`
	RemoteCommands []string        `json:"remote_commands,omitempty"`
	Uploads        []ReleaseUpload `json:"uploads,omitempty"`
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

func NewReleaseRecord(serverState *ServerState, remoteCommands []string, uploads []ReleaseUpload) ReleaseRecord {
	projectPath, _ := os.Getwd()
	record := ReleaseRecord{
		ID:             time.Now().UTC().Format("20060102T150405.000000000Z"),
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		ProjectPath:    projectPath,
		GitSHA:         currentGitSHA(),
		Status:         "success",
		RemoteCommands: append([]string(nil), remoteCommands...),
		Uploads:        append([]ReleaseUpload(nil), uploads...),
	}
	if serverState != nil {
		record.ServerID = serverState.ServerID
		record.ServerIP = serverState.IP
	}
	return record
}

func currentGitSHA() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
