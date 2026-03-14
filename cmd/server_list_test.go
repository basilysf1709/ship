package cmd

import (
	"bytes"
	"strings"
	"testing"

	shipinternal "ship/internal"
)

func TestServerListCommandRendersTable(t *testing.T) {
	originalListServerInventory := listServerInventory
	defer func() {
		listServerInventory = originalListServerInventory
	}()

	listServerInventory = func() ([]shipinternal.ServerRecord, error) {
		return []shipinternal.ServerRecord{
			{
				Provider:    "digitalocean",
				ServerID:    "srv-1",
				IP:          "1.2.3.4",
				ProjectPath: "/tmp/demo-app",
				CreatedAt:   "2026-03-13T20:00:00Z",
			},
			{
				Provider:    "hetzner",
				ServerID:    "srv-2",
				IP:          "5.6.7.8",
				ProjectPath: "/tmp/api",
				CreatedAt:   "2026-03-12T18:30:00Z",
			},
		}, nil
	}

	cmd := newServerListCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"PROVIDER",
		"srv-1",
		"http://1.2.3.4",
		"demo-app",
		"srv-2",
		"http://5.6.7.8",
		"api",
		"TOTAL_SERVERS=2",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q: %q", want, output)
		}
	}
}
