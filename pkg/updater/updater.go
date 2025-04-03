package updater

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CheckUpdates uses dnf to check for available system updates
func CheckUpdates() ([]string, error) {
	cmd := exec.Command("chroot", "/host", "dnf", "check-update")
	output, err := cmd.CombinedOutput()

	// DNF returns exit code 100 when updates are available
	if err != nil && cmd.ProcessState.ExitCode() != 100 {
		return nil, fmt.Errorf("error checking updates: %v, output: %s", err, string(output))
	}

	return parseUpdateList(string(output)), nil
}

// UpgradePackages performs a system upgrade using dnf and reboots if kernel was updated
func UpgradePackages() error {
	// First check what packages will be updated
	updatesBeforeUpgrade, err := CheckUpdates()
	if err != nil {
		return fmt.Errorf("error checking updates before upgrade: %v", err)
	}

	// Check if any kernel packages are in the update list
	kernelUpdateNeeded := false
	for _, pkg := range updatesBeforeUpgrade {
		if strings.HasPrefix(pkg, "kernel") || strings.HasPrefix(pkg, "linux-firmware") {
			kernelUpdateNeeded = true
			break
		}
	}

	// Perform the upgrade
	cmd := exec.Command("chroot", "/host", "dnf", "upgrade", "-y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error upgrading packages: %v, output: %s", err, string(output))
	}

	// If kernel was updated, reboot the system
	if kernelUpdateNeeded && os.Getenv("KERNEL_UPDATE_REBOOT") == "true" {
		fmt.Println("Kernel update detected. Rebooting system...")
		rebootCmd := exec.Command("chroot", "/host", "systemctl", "reboot")
		rebootOutput, rebootErr := rebootCmd.CombinedOutput()
		if rebootErr != nil {
			return fmt.Errorf("error rebooting system: %v, output: %s", rebootErr, string(rebootOutput))
		}
	}

	return nil
}

// Helper function to parse the output of dnf check-update
func parseUpdateList(output string) []string {
	lines := strings.Split(output, "\n")
	var updates []string
	var startParsing bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// The actual update list starts after the metadata information
		if !startParsing {
			// Look for the divider line that contains "="
			if strings.Contains(line, "") {
				startParsing = true
			}
			continue
		}

		// Process package lines
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			// First field is the package name
			updates = append(updates, fields[0])
		}
	}

	return updates
}
