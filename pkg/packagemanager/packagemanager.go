// Package packagemanager provides interfaces and types for system package management.
package packagemanager

// Info represents information about a system package.
type Info struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Architecture string       `json:"architecture"`
	Description  string       `json:"description"`
	Upgrade      *UpgradeInfo `json:"upgrade,omitempty"`
}

// UpgradeInfo contains information about available upgrades for a package.
type UpgradeInfo struct {
	NewVersion string         `json:"new_version"`
	HasUpgrade bool           `json:"has_upgrade"`
	Changelog  *ChangelogInfo `json:"changelog,omitempty"`
	RiskLevel  string         `json:"risk_level,omitempty"`
	RiskReason string         `json:"risk_reason,omitempty"`
}

// ChangelogInfo contains parsed changelog information.
type ChangelogInfo struct {
	Summary string   `json:"summary"`
	Urgency string   `json:"urgency"`
	CVEs    []string `json:"cves,omitempty"`
	Author  string   `json:"author,omitempty"`
	Date    string   `json:"date,omitempty"`
	Raw     string   `json:"raw"`
}

// Manager defines the interface for package management operations.
type Manager interface {
	// GetInstalledPackages returns a list of all installed packages.
	GetInstalledPackages() ([]Info, error)

	// GetUpgradablePackages returns a map of package names to their upgrade versions.
	GetUpgradablePackages() map[string]string

	// GetChangelog retrieves and parses changelog information for a package.
	GetChangelog(pkgName string, stopVersion string) *ChangelogInfo

	// UpdatePackages performs safe upgrades on the provided packages.
	UpdatePackages(packages []Info) error

	EnrichWithUpgradeInfo(installedPackages []Info, upgradableMap map[string]string) []Info
}
