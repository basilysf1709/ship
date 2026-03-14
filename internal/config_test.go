package shipinternal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadDeleteServerState(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	state := ServerState{
		ServerID: "12345",
		IP:       "1.2.3.4",
	}

	if err := SaveServerState(state); err != nil {
		t.Fatalf("SaveServerState returned error: %v", err)
	}

	loaded, err := LoadServerState()
	if err != nil {
		t.Fatalf("LoadServerState returned error: %v", err)
	}
	expected := ServerState{
		Provider: "digitalocean",
		ServerID: "12345",
		IP:       "1.2.3.4",
		SSHUser:  "root",
	}
	if loaded != expected {
		t.Fatalf("LoadServerState = %+v, want %+v", loaded, expected)
	}

	if _, err := os.Stat(filepath.Join(".ship", "server.json")); err != nil {
		t.Fatalf("server state file not written: %v", err)
	}

	if err := DeleteServerState(); err != nil {
		t.Fatalf("DeleteServerState returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(".ship", "server.json")); !os.IsNotExist(err) {
		t.Fatalf("server state file still exists after delete, stat err=%v", err)
	}
}

func TestLoadServerStateMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	_, err = LoadServerState()
	if err == nil {
		t.Fatal("LoadServerState returned nil error for missing file")
	}
}

func TestAddListRemoveServerInventoryRecord(t *testing.T) {
	tempDir := t.TempDir()
	originalHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() {
		userHomeDir = originalHomeDir
	}()

	state := ServerState{
		Provider: "digitalocean",
		ServerID: "12345",
		IP:       "1.2.3.4",
		SSHUser:  "root",
	}

	if err := AddServerInventoryRecord(state, "/tmp/project-a"); err != nil {
		t.Fatalf("AddServerInventoryRecord returned error: %v", err)
	}

	records, err := ListServerInventory()
	if err != nil {
		t.Fatalf("ListServerInventory returned error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("ListServerInventory len = %d, want 1", len(records))
	}
	if records[0].ServerID != state.ServerID || records[0].ProjectPath != "/tmp/project-a" {
		t.Fatalf("ListServerInventory[0] = %+v", records[0])
	}
	if records[0].Link() != "http://1.2.3.4" {
		t.Fatalf("Link() = %q", records[0].Link())
	}

	if err := AddServerInventoryRecord(state, "/tmp/project-b"); err != nil {
		t.Fatalf("AddServerInventoryRecord update returned error: %v", err)
	}

	records, err = ListServerInventory()
	if err != nil {
		t.Fatalf("ListServerInventory returned error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("updated ListServerInventory len = %d, want 1", len(records))
	}
	if records[0].ProjectPath != "/tmp/project-b" {
		t.Fatalf("updated project path = %q, want /tmp/project-b", records[0].ProjectPath)
	}

	if _, err := os.Stat(filepath.Join(tempDir, ".ship", "servers.json")); err != nil {
		t.Fatalf("inventory file not written: %v", err)
	}

	if err := RemoveServerInventoryRecord(state); err != nil {
		t.Fatalf("RemoveServerInventoryRecord returned error: %v", err)
	}

	records, err = ListServerInventory()
	if err != nil {
		t.Fatalf("ListServerInventory after remove returned error: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("ListServerInventory after remove len = %d, want 0", len(records))
	}
}
