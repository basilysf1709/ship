package shipinternal

import "testing"

func TestDeployConfigHashNormalizesModes(t *testing.T) {
	first, err := DeployConfigHash(DeployConfig{
		LocalCommands: []string{"docker build -t app ."},
		Uploads: []DeployUpload{{
			Source:      "app.tar",
			Destination: "/root/app.tar",
			Mode:        "0644",
		}},
		RemoteCommands: []string{"docker load -i /root/app.tar"},
		CleanupLocal:   []string{"app.tar"},
	})
	if err != nil {
		t.Fatalf("DeployConfigHash returned error: %v", err)
	}
	second, err := DeployConfigHash(DeployConfig{
		LocalCommands: []string{"docker build -t app ."},
		Uploads: []DeployUpload{{
			Source:      "app.tar",
			Destination: "/root/app.tar",
		}},
		RemoteCommands: []string{"docker load -i /root/app.tar"},
		CleanupLocal:   []string{"app.tar"},
	})
	if err != nil {
		t.Fatalf("DeployConfigHash returned error: %v", err)
	}

	if first != second {
		t.Fatalf("DeployConfigHash mismatch: %q vs %q", first, second)
	}
}

func TestReleaseRecordRollbackEligibilityRequiresBackups(t *testing.T) {
	record := &ReleaseRecord{
		Status:   "success",
		ServerID: "srv-1",
		ServerIP: "1.2.3.4",
	}
	record.UpdateRollbackEligibility()
	if record.RollbackEligible {
		t.Fatal("RollbackEligible = true, want false")
	}
	if record.RollbackReason == "" {
		t.Fatal("RollbackReason = empty, want reason")
	}

	record.Uploads = []ReleaseUpload{{
		Destination: "/root/app.tar",
		BackupPath:  "/root/.ship/releases/r1/00-app.tar",
	}}
	record.UpdateRollbackEligibility()
	if !record.RollbackEligible {
		t.Fatalf("RollbackEligible = false, reason=%q", record.RollbackReason)
	}
}
