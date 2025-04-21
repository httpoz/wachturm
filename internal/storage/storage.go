// Package storage provides functionality for storing and retrieving package data.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/httpoz/wachturm/internal/packagemanager"
)

// Constants for snapshot types
const (
	SnapshotTypeInstalled = "installed"
	SnapshotTypeUpdates   = "updates"
)

// Storage handles loading and saving package data.
type Storage struct {
	baseDir string
}

// New creates a new Storage instance with the given base directory.
func New(baseDir string) *Storage {
	return &Storage{
		baseDir: baseDir,
	}
}

// DefaultStorage creates a new Storage instance with the default base directory.
func DefaultStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".wachturm")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return New(baseDir), nil
}

// CheckSnapshot checks if a snapshot exists and creates it if it doesn't.
func (s *Storage) CheckSnapshot(id string) (string, error) {
	snapshotDir := filepath.Join(s.baseDir, "snapshots", id)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	snapshotFile := filepath.Join(snapshotDir, "installed.json")

	// Check if the snapshot file already exists
	if _, err := os.Stat(snapshotFile); err == nil {
		return snapshotFile, nil
	}

	return snapshotFile, nil
}

// WriteSnapshot writes package data to a snapshot file.
func (s *Storage) WriteSnapshot(packages []packagemanager.Info, id string, snapshotType string) (string, error) {
	snapshotDir := filepath.Join(s.baseDir, "snapshots", id)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	filename := "installed.json"
	if snapshotType == SnapshotTypeUpdates {
		filename = "updates.json"
	}

	snapshotFile := filepath.Join(snapshotDir, filename)

	file, err := os.Create(snapshotFile)
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(packages); err != nil {
		return "", fmt.Errorf("failed to encode package data: %w", err)
	}

	return snapshotFile, nil
}

// ReadSnapshot reads package data from a snapshot file.
func (s *Storage) ReadSnapshot(id string) ([]packagemanager.Info, error) {
	snapshotFile := filepath.Join(s.baseDir, "snapshots", id, "installed.json")

	file, err := os.Open(snapshotFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open snapshot file: %w", err)
	}
	defer file.Close()

	var packages []packagemanager.Info
	if err := json.NewDecoder(file).Decode(&packages); err != nil {
		return nil, fmt.Errorf("failed to decode package data: %w", err)
	}

	return packages, nil
}

// WriteSummaryIfMissing writes a summary file if it doesn't already exist.
func (s *Storage) WriteSummaryIfMissing(id string, packages []packagemanager.Info) (string, error) {
	summaryFile := filepath.Join(s.baseDir, "snapshots", id, "summary.txt")

	// Check if the summary file already exists
	if _, err := os.Stat(summaryFile); err == nil {
		return summaryFile, nil
	}

	file, err := os.Create(summaryFile)
	if err != nil {
		return "", fmt.Errorf("failed to create summary file: %w", err)
	}
	defer file.Close()

	var totalCount, highRiskCount, mediumRiskCount, lowRiskCount int
	var highRiskPackages, mediumRiskPackages []string

	for _, pkg := range packages {
		if pkg.Upgrade == nil {
			continue
		}

		totalCount++
		switch pkg.Upgrade.RiskLevel {
		case "high":
			highRiskCount++
			highRiskPackages = append(highRiskPackages, pkg.Name)
		case "medium":
			mediumRiskCount++
			mediumRiskPackages = append(mediumRiskPackages, pkg.Name)
		case "low":
			lowRiskCount++
		}
	}

	timestamp := time.Now().Format(time.RFC1123)

	fmt.Fprintf(file, "Update Summary - %s\n\n", timestamp)
	fmt.Fprintf(file, "Total packages available for update: %d\n", totalCount)
	fmt.Fprintf(file, "High risk updates: %d\n", highRiskCount)
	fmt.Fprintf(file, "Medium risk updates: %d\n", mediumRiskCount)
	fmt.Fprintf(file, "Low risk updates: %d\n\n", lowRiskCount)

	if len(highRiskPackages) > 0 {
		fmt.Fprintf(file, "High risk packages (manual review recommended):\n")
		for _, pkg := range highRiskPackages {
			fmt.Fprintf(file, "- %s\n", pkg)
		}
		fmt.Fprintln(file)
	}

	if len(mediumRiskPackages) > 0 {
		fmt.Fprintf(file, "Medium risk packages (caution advised):\n")
		for _, pkg := range mediumRiskPackages {
			fmt.Fprintf(file, "- %s\n", pkg)
		}
		fmt.Fprintln(file)
	}

	fmt.Fprintf(file, "Low risk packages will be updated automatically.\n")

	return summaryFile, nil
}

// FilterUpgradable returns only the packages that have available upgrades.
func FilterUpgradable(packages []packagemanager.Info) []packagemanager.Info {
	var upgradable []packagemanager.Info
	for _, pkg := range packages {
		if pkg.Upgrade != nil && pkg.Upgrade.HasUpgrade {
			upgradable = append(upgradable, pkg)
		}
	}
	return upgradable
}
