package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/httpoz/watchturm/pkg/packagemanager"
)

func TestStorage_WriteAndReadSnapshot(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test storage instance
	storage := New(tempDir)

	// Create test package data
	testPackages := []packagemanager.Info{
		{
			Name:         "apt",
			Version:      "2.4.9",
			Architecture: "amd64",
			Description:  "package manager",
		},
		{
			Name:         "bash",
			Version:      "5.1-6ubuntu1",
			Architecture: "amd64",
			Description:  "shell",
			Upgrade: &packagemanager.UpgradeInfo{
				NewVersion: "5.1-6ubuntu1.1",
				HasUpgrade: true,
				RiskLevel:  "low",
			},
		},
	}

	// Test writing a snapshot
	snapshotID := "test-snapshot"
	snapshotPath, err := storage.WriteSnapshot(testPackages, snapshotID, SnapshotTypeInstalled)
	if err != nil {
		t.Fatalf("WriteSnapshot failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		t.Errorf("Snapshot file wasn't created at %s", snapshotPath)
	}

	// Read the file directly to verify its contents
	file, err := os.Open(snapshotPath)
	if err != nil {
		t.Fatalf("Failed to open snapshot file: %v", err)
	}
	defer file.Close()

	var readPackages []packagemanager.Info
	if err := json.NewDecoder(file).Decode(&readPackages); err != nil {
		t.Fatalf("Failed to decode snapshot file: %v", err)
	}

	// Verify the data
	if len(readPackages) != len(testPackages) {
		t.Errorf("Expected %d packages, got %d", len(testPackages), len(readPackages))
	}

	// Test ReadSnapshot
	readResult, err := storage.ReadSnapshot(snapshotID)
	if err != nil {
		t.Fatalf("ReadSnapshot failed: %v", err)
	}

	if len(readResult) != len(testPackages) {
		t.Errorf("ReadSnapshot: expected %d packages, got %d", len(testPackages), len(readResult))
	}
}

func TestStorage_WriteSummaryIfMissing(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test storage instance
	storage := New(tempDir)

	// Create test package data with risk levels
	testPackages := []packagemanager.Info{
		{
			Name:    "apt",
			Version: "2.4.9",
			Upgrade: &packagemanager.UpgradeInfo{
				HasUpgrade: true,
				RiskLevel:  "low",
			},
		},
		{
			Name:    "bash",
			Version: "5.1-6ubuntu1",
			Upgrade: &packagemanager.UpgradeInfo{
				HasUpgrade: true,
				RiskLevel:  "medium",
			},
		},
		{
			Name:    "openssl",
			Version: "3.0.2",
			Upgrade: &packagemanager.UpgradeInfo{
				HasUpgrade: true,
				RiskLevel:  "high",
			},
		},
	}

	// Ensure snapshot directory exists
	snapshotID := "test-summary"
	snapshotDir := filepath.Join(tempDir, "snapshots", snapshotID)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		t.Fatalf("Failed to create snapshot directory: %v", err)
	}

	// Test creating summary
	summaryPath, err := storage.WriteSummaryIfMissing(snapshotID, testPackages)
	if err != nil {
		t.Fatalf("WriteSummaryIfMissing failed: %v", err)
	}

	// Verify summary file exists
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Errorf("Summary file wasn't created at %s", summaryPath)
	}

	// Read the file contents
	content, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("Failed to read summary file: %v", err)
	}

	// Verify content has the basic information
	contentStr := string(content)
	if !contains(contentStr, "Total packages available for update: 3") {
		t.Errorf("Summary doesn't contain expected total count")
	}
	if !contains(contentStr, "High risk updates: 1") {
		t.Errorf("Summary doesn't contain expected high risk count")
	}
	if !contains(contentStr, "Medium risk updates: 1") {
		t.Errorf("Summary doesn't contain expected medium risk count")
	}
	if !contains(contentStr, "Low risk updates: 1") {
		t.Errorf("Summary doesn't contain expected low risk count")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
