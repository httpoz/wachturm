package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/httpoz/watchtower/pkg/packagemanager"
	"github.com/httpoz/watchtower/pkg/storage"
)

// mockPackageManager is a mock implementation of the packagemanager.Manager interface
type mockPackageManager struct {
	installedPackages []packagemanager.Info
	upgradableMap     map[string]string
	changelogs        map[string]*packagemanager.ChangelogInfo
	updateCalled      bool
}

func (m *mockPackageManager) GetInstalledPackages() ([]packagemanager.Info, error) {
	return m.installedPackages, nil
}

func (m *mockPackageManager) GetUpgradablePackages() map[string]string {
	return m.upgradableMap
}

func (m *mockPackageManager) GetChangelog(pkgName string, stopVersion string) *packagemanager.ChangelogInfo {
	return m.changelogs[pkgName]
}

func (m *mockPackageManager) UpdatePackages(packages []packagemanager.Info) error {
	m.updateCalled = true
	return nil
}

// mockRiskAssessor is a mock implementation for testing
type mockRiskAssessor struct {
	scoredPackages []packagemanager.Info
}

func (m *mockRiskAssessor) ScorePackageRisks(ctx context.Context, packages []packagemanager.Info) ([]packagemanager.Info, error) {
	return m.scoredPackages, nil
}

// TestApplicationFlow tests the basic flow of the application
func TestApplicationFlow(t *testing.T) {
	// Set up mocks
	mockPkgMgr := &mockPackageManager{
		installedPackages: []packagemanager.Info{
			{
				Name:         "apt",
				Version:      "2.4.8",
				Architecture: "amd64",
				Description:  "package manager",
			},
			{
				Name:         "bash",
				Version:      "5.1-6ubuntu1",
				Architecture: "amd64",
				Description:  "shell",
			},
		},
		upgradableMap: map[string]string{
			"apt":  "2.4.9",
			"bash": "5.1-6ubuntu1.1",
		},
		changelogs: map[string]*packagemanager.ChangelogInfo{
			"apt": {
				Summary: "Security fix for CVE-2023-12345",
				Urgency: "medium",
				CVEs:    []string{"CVE-2023-12345"},
				Raw:     "apt changelog content",
			},
			"bash": {
				Summary: "Minor improvements",
				Urgency: "low",
				Raw:     "bash changelog content",
			},
		},
	}

	// Create scored packages that would normally be returned by the risk assessor
	scoredPackages := []packagemanager.Info{
		{
			Name:         "apt",
			Version:      "2.4.8",
			Architecture: "amd64",
			Description:  "package manager",
			Upgrade: &packagemanager.UpgradeInfo{
				NewVersion: "2.4.9",
				HasUpgrade: true,
				RiskLevel:  "medium",
				RiskReason: "Contains security fixes",
				Changelog: &packagemanager.ChangelogInfo{
					Summary: "Security fix for CVE-2023-12345",
					Urgency: "medium",
					CVEs:    []string{"CVE-2023-12345"},
					Raw:     "apt changelog content",
				},
			},
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
				RiskReason: "Minor updates only",
				Changelog: &packagemanager.ChangelogInfo{
					Summary: "Minor improvements",
					Urgency: "low",
					Raw:     "bash changelog content",
				},
			},
		},
	}

	mockRiskAssessor := &mockRiskAssessor{
		scoredPackages: scoredPackages,
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "watchtower_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test storage manager
	storageManager := storage.New(tempDir)

	// Run a simplified version of the main application flow
	ctx := context.Background()
	snapshotID := "test-run"

	// Get installed packages
	installedPackages, err := mockPkgMgr.GetInstalledPackages()
	if err != nil {
		t.Fatalf("Error getting installed packages: %v", err)
	}

	// Save snapshot of installed packages
	_, err = storageManager.WriteSnapshot(installedPackages, snapshotID, storage.SnapshotTypeInstalled)
	if err != nil {
		t.Fatalf("Error writing installed packages snapshot: %v", err)
	}

	// Get upgradable packages
	upgradableMap := mockPkgMgr.GetUpgradablePackages()

	// Enrich installed packages with upgrade information
	var upgradablePackages []packagemanager.Info
	for i, pkg := range installedPackages {
		if newVersion, ok := upgradableMap[pkg.Name]; ok {
			installedPackages[i].Upgrade = &packagemanager.UpgradeInfo{
				NewVersion: newVersion,
				HasUpgrade: true,
			}

			// Get changelog for the package
			changelog := mockPkgMgr.GetChangelog(pkg.Name, pkg.Version)
			if changelog != nil {
				installedPackages[i].Upgrade.Changelog = changelog
			}

			// Add to upgradable packages list
			upgradablePackages = append(upgradablePackages, installedPackages[i])
		}
	}

	// Score the risk of upgradable packages
	scoredPackages, err = mockRiskAssessor.ScorePackageRisks(ctx, upgradablePackages)
	if err != nil {
		t.Fatalf("Error scoring package risks: %v", err)
	}

	// Save snapshot of upgradable packages with risk assessment
	_, err = storageManager.WriteSnapshot(scoredPackages, snapshotID, storage.SnapshotTypeUpdates)
	if err != nil {
		t.Fatalf("Error writing updates snapshot: %v", err)
	}

	// Generate and save summary
	summaryPath, err := storageManager.WriteSummaryIfMissing(snapshotID, scoredPackages)
	if err != nil {
		t.Fatalf("Error writing summary: %v", err)
	}

	// Verify the summary file was created
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Errorf("Summary file wasn't created")
	}

	// Update packages
	err = mockPkgMgr.UpdatePackages(scoredPackages)
	if err != nil {
		t.Fatalf("Error updating packages: %v", err)
	}

	// Verify the update was called
	if !mockPkgMgr.updateCalled {
		t.Errorf("UpdatePackages was not called")
	}

	// Verify the snapshot and summary files were created
	installedPath := filepath.Join(tempDir, "snapshots", snapshotID, "installed.json")
	updatesPath := filepath.Join(tempDir, "snapshots", snapshotID, "updates.json")
	summaryPath = filepath.Join(tempDir, "snapshots", snapshotID, "summary.txt")

	for _, path := range []string{installedPath, updatesPath, summaryPath} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", path)
		}
	}
}
