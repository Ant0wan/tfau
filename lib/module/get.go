package module

import (
	"fmt"
	"strings"
)

// getLatestModuleVersion retrieves the latest version of a module based on its source.
func getLatestModuleVersion(source string) (string, error) {
	// Check if the source is a Terraform Registry module
	if isRegistryModule(source) {
		return getLatestVersionFromRegistry(source)
	}

	// Check if the source is a Git-based module
	if isGitModule(source) {
		return getLatestVersionFromGit(source)
	}

	// If the source format is not recognized, return an error
	return "", fmt.Errorf("unsupported module source format: %s", source)
}

// isRegistryModule checks if the source is a Terraform Registry module.
func isRegistryModule(source string) bool {
	// Terraform Registry modules are in the format: namespace/name/provider
	parts := strings.Split(source, "/")
	return len(parts) == 3 && !strings.HasPrefix(source, "git@") && !strings.HasPrefix(source, "ssh://")
}

// isGitModule checks if the source is a Git-based module.
func isGitModule(source string) bool {
	return strings.HasPrefix(source, "git@") || strings.HasPrefix(source, "ssh://") || strings.HasPrefix(source, "https://")
}

// getLatestVersionFromRegistry retrieves the latest version of a Terraform Registry module.
func getLatestVersionFromRegistry(source string) (string, error) {
	// Implement logic to query the Terraform Registry API
	// Example: https://registry.terraform.io/v1/modules/namespace/name/provider/versions
	return "", fmt.Errorf("not implemented: registry module support")
}

// getLatestVersionFromGit retrieves the latest version of a Git-based module.
func getLatestVersionFromGit(source string) (string, error) {
	// Implement logic to query Git tags or releases
	// Example: Use `git ls-remote` or GitHub API to fetch the latest tag
	return "", fmt.Errorf("not implemented: Git module support")
}
