package apt

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// mockCommand creates a mock command for testing
func mockCommand(t *testing.T, expectedCommand string, output string, err error) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"EXPECTED_COMMAND=" + expectedCommand,
			"MOCK_OUTPUT=" + output,
		}
		return cmd
	}
}

// TestHelperProcess isn't a real test - it's used to mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	expectedCommand := os.Getenv("EXPECTED_COMMAND")
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) == 0 {
		os.Exit(1)
	}

	if args[0] != expectedCommand {
		os.Exit(1)
	}

	os.Stdout.WriteString(os.Getenv("MOCK_OUTPUT"))
}

func TestGetInstalledPackages(t *testing.T) {
	mockOutput := `Desired=Unknown/Install/Remove/Purge/Hold
| Status=Not/Inst/Conf-files/Unpacked/halF-conf/Half-inst/trig-aWait/Trig-pend
|/ Err?=(none)/Reinst-required (Status,Err: uppercase=bad)
||/ Name           Version      Architecture Description
+++-==============-============-============-=================================
ii  apt            2.4.9        amd64        commandline package manager
ii  bash           5.1-6ubuntu1 amd64        GNU Bourne Again SHell`

	manager := NewManager().WithExecCommand(mockCommand(t, "dpkg", mockOutput, nil))

	packages, err := manager.GetInstalledPackages()
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	expected := []struct {
		name    string
		version string
		arch    string
	}{
		{"apt", "2.4.9", "amd64"},
		{"bash", "5.1-6ubuntu1", "amd64"},
	}

	if len(packages) != len(expected) {
		t.Fatalf("Expected %d packages but got %d", len(expected), len(packages))
	}

	for i, pkg := range packages {
		if pkg.Name != expected[i].name {
			t.Errorf("Expected package name %s but got %s", expected[i].name, pkg.Name)
		}
		if pkg.Version != expected[i].version {
			t.Errorf("Expected version %s but got %s", expected[i].version, pkg.Version)
		}
		if pkg.Architecture != expected[i].arch {
			t.Errorf("Expected architecture %s but got %s", expected[i].arch, pkg.Architecture)
		}
	}
}

func TestGetUpgradablePackages(t *testing.T) {
	mockOutput := `Listing... Done
libc6/jammy-updates 2.35-0ubuntu3.4 amd64 [upgradable from: 2.35-0ubuntu3.3]
bash/jammy-updates 5.1-6ubuntu1.1 amd64 [upgradable from: 5.1-6ubuntu1]`

	manager := NewManager().WithExecCommand(mockCommand(t, "apt", mockOutput, nil))

	upgradable := manager.GetUpgradablePackages()

	expected := map[string]string{
		"libc6": "2.35-0ubuntu3.4",
		"bash":  "5.1-6ubuntu1.1",
	}

	if len(upgradable) != len(expected) {
		t.Fatalf("Expected %d upgradable packages but got %d", len(expected), len(upgradable))
	}

	for pkg, version := range expected {
		if upgVer, ok := upgradable[pkg]; !ok {
			t.Errorf("Expected package %s to be upgradable, but it wasn't", pkg)
		} else if upgVer != version {
			t.Errorf("For package %s, expected version %s but got %s", pkg, version, upgVer)
		}
	}
}

func TestParseUpgradablePackages(t *testing.T) {
	testInput := `Listing... Done
libperl5.34/jammy-updates,jammy-security 5.34.0-3ubuntu1.4 arm64 [upgradable from: 5.34.0-3ubuntu1.3]
perl-base/jammy-updates,jammy-security 5.34.0-3ubuntu1.4 arm64 [upgradable from: 5.34.0-3ubuntu1.3]`

	testInputWithArch := `Listing... Done
libperl5.34:arm64/jammy-updates,jammy-security 5.34.0-3ubuntu1.4 arm64 [upgradable from: 5.34.0-3ubuntu1.3]`

	t.Run("TestNormalPackages", func(t *testing.T) {
		reader := strings.NewReader(testInput)
		result := parseUpgradablePackages(reader)

		expected := map[string]string{
			"libperl5.34": "5.34.0-3ubuntu1.4",
			"perl-base":   "5.34.0-3ubuntu1.4",
		}

		if len(result) != len(expected) {
			t.Fatalf("Expected %d upgradable packages but got %d", len(expected), len(result))
		}

		for pkg, version := range expected {
			if resultVer, ok := result[pkg]; !ok {
				t.Errorf("Expected package %s to be in results", pkg)
			} else if resultVer != version {
				t.Errorf("Incorrect version for %s: got %s, want %s", pkg, resultVer, version)
			}
		}
	})

	t.Run("TestArchitectureInName", func(t *testing.T) {
		reader := strings.NewReader(testInputWithArch)
		result := parseUpgradablePackages(reader)

		if version, ok := result["libperl5.34"]; !ok {
			t.Error("Failed to parse package with architecture in name")
		} else if version != "5.34.0-3ubuntu1.4" {
			t.Errorf("Incorrect version: got %s, want %s", version, "5.34.0-3ubuntu1.4")
		}
	})
}

func TestGetChangelog(t *testing.T) {
	mockChangelogOutput := `apt (2.4.9) jammy-updates; urgency=medium

  * Security fix for CVE-2023-12345
  * Performance improvements

 -- Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>  Wed, 16 Apr 2025 10:00:00 +0000

apt (2.4.8) jammy; urgency=low

  * Initial release
  
 -- Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>  Wed, 10 Apr 2025 10:00:00 +0000`

	manager := NewManager().WithExecCommand(mockCommand(t, "apt", mockChangelogOutput, nil))

	changelog := manager.GetChangelog("apt", "2.4.8")

	if changelog == nil {
		t.Fatal("Expected changelog, but got nil")
	}

	if changelog.Urgency != "medium" {
		t.Errorf("Expected urgency 'medium', got %q", changelog.Urgency)
	}

	if len(changelog.CVEs) != 1 {
		t.Errorf("Expected 1 CVE, got %d", len(changelog.CVEs))
	}

	if len(changelog.CVEs) > 0 && changelog.CVEs[0] != "CVE-2023-12345" {
		t.Errorf("Expected CVE-2023-12345, got %s", changelog.CVEs[0])
	}
}
