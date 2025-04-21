// Package apt provides an implementation of the package manager interface for apt-based systems.
package apt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/httpoz/watchturm/pkg/packagemanager"
)

type Manager struct {
	execCommand func(string, ...string) *exec.Cmd
}

func NewManager() *Manager {
	return &Manager{
		execCommand: exec.Command,
	}
}

func (m *Manager) WithExecCommand(execFn func(string, ...string) *exec.Cmd) *Manager {
	return &Manager{
		execCommand: execFn,
	}
}

// GetInstalledPackages returns information about all installed packages.
func (m *Manager) GetInstalledPackages() ([]packagemanager.Info, error) {
	output, err := m.execCommand("dpkg", "-l").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute dpkg command: %w", err)
	}

	var packages []packagemanager.Info
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Properly check for lines starting with "ii" (installed packages)
		if !strings.HasPrefix(line, "ii") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		packages = append(packages, packagemanager.Info{
			Name:         fields[1],
			Version:      fields[2],
			Architecture: fields[3],
			Description:  strings.Join(fields[4:], " "),
		})
	}
	return packages, nil
}

// GetUpgradablePackages returns a map of upgradable packages with their versions.
func (m *Manager) GetUpgradablePackages() map[string]string {
	upgOut, err := m.execCommand("apt", "list", "--upgradable").Output()
	if err != nil {
		return make(map[string]string)
	}

	return parseUpgradablePackages(bytes.NewReader(upgOut))
}

// parseUpgradablePackages parses the output of apt list --upgradable
// and returns a map of package names to new versions.
func parseUpgradablePackages(reader io.Reader) map[string]string {
	upgradableMap := make(map[string]string)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[upgradable from:") {
			// Parse format: name/repo version arch [upgradable from: oldversion]
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			// Extract package name
			parts := strings.SplitN(fields[0], "/", 2)
			if len(parts) < 1 {
				continue
			}

			pkgName := strings.TrimSpace(parts[0])
			// Handle architecture suffix in package name (libperl5.34:arm64 -> libperl5.34)
			if strings.Contains(pkgName, ":") {
				pkgName = strings.Split(pkgName, ":")[0]
			}

			// Extract new version (usually field index 1)
			if len(fields) > 1 {
				newVersion := fields[1]
				upgradableMap[pkgName] = newVersion
			}
		}
	}

	return upgradableMap
}

// GetChangelog retrieves and parses the changelog for a specific package.
func (m *Manager) GetChangelog(pkgName string, stopVersion string) *packagemanager.ChangelogInfo {
	out, err := m.execCommand("apt", "changelog", pkgName).Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(string(out), "\n")

	var b strings.Builder
	var urgency, summary, author, date string
	var cves []string

	for i, line := range lines {
		if strings.Contains(line, stopVersion) {
			break
		}

		// Parse the urgency from lines that follow the pattern: "package (version) distro; urgency=level"
		if strings.Contains(line, "urgency=") {
			parts := strings.Split(line, "urgency=")
			if len(parts) > 1 {
				urgencyPart := parts[1]
				// Extract just the urgency level (medium, high, etc.)
				if idx := strings.Index(urgencyPart, " "); idx > 0 {
					urgency = urgencyPart[:idx]
				} else {
					urgency = strings.TrimSpace(urgencyPart)
				}
			}
		}

		// Extract the first non-empty line after the header as the summary
		if i > 2 && summary == "" && strings.TrimSpace(line) != "" && strings.HasPrefix(strings.TrimSpace(line), "*") {
			summary = strings.TrimSpace(line)
		}

		if strings.Contains(line, "CVE-") {
			words := strings.Fields(line)
			for _, word := range words {
				if strings.HasPrefix(word, "CVE-") {
					cves = append(cves, strings.Trim(word, ".,"))
				}
			}
		}

		if strings.HasPrefix(line, " -- ") {
			authorDate := strings.TrimPrefix(line, " -- ")
			parts := strings.SplitN(authorDate, "  ", 2)
			if len(parts) == 2 {
				author = strings.TrimSpace(parts[0])
				date = strings.TrimSpace(parts[1])
			}
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return &packagemanager.ChangelogInfo{
		Summary: summary,
		Urgency: urgency,
		CVEs:    cves,
		Author:  author,
		Date:    date,
		Raw:     b.String(),
	}
}

// UpdatePackages upgrades the packages deemed safe for upgrading.
func (m *Manager) UpdatePackages(packages []packagemanager.Info) error {
	var safe []string
	for _, pkg := range packages {
		if strings.ToLower(pkg.Upgrade.RiskLevel) == "low" {
			safe = append(safe, pkg.Name)
		}
	}

	if len(safe) == 0 {
		fmt.Println("No packages to upgrade.")
		return nil
	}

	fmt.Printf("Upgrading safe packages: %v\n", safe)
	args := append([]string{"install", "--only-upgrade", "-y"}, safe...)
	cmd := m.execCommand("apt-get", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("apt-get upgrade failed: %w", err)
	}

	fmt.Println("Upgrade completed.")
	return nil
}

func (m *Manager) EnrichWithUpgradeInfo(installedPackages []packagemanager.Info, upgradableMap map[string]string) []packagemanager.Info {
	// Enrich installed packages with upgrade information
	var upgradablePackages []packagemanager.Info
	for i, pkg := range installedPackages {
		if newVersion, ok := upgradableMap[pkg.Name]; ok {
			installedPackages[i].Upgrade = &packagemanager.UpgradeInfo{
				NewVersion: newVersion,
				HasUpgrade: true,
			}

			// Get changelog for the package
			fmt.Printf("Fetching changelog for %s...\n", pkg.Name)
			changelog := m.GetChangelog(pkg.Name, pkg.Version)
			if changelog != nil {
				installedPackages[i].Upgrade.Changelog = changelog
			}

			// Add to upgradable packages list
			upgradablePackages = append(upgradablePackages, installedPackages[i])
		}
	}

	return upgradablePackages
}
