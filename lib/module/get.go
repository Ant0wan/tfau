package module

import (
	"fmt"
	"strings"
)

// GetLatestModuleVersion retrieves the latest version of a module based on its source.
func GetLatestModuleVersion(source string) (string, error) {
	// Normalize the source by removing subdirectory information
	normalizedSource := normalizeSource(source)

	// Check if the source is a Terraform Registry module
	if isRegistryModule(normalizedSource) {
		return getLatestVersionFromRegistry(source) // Pass the original source to handle submodules
	}

	// Check if the source is a Git-based module
	if isGitModule(source) {
		return getLatestVersionFromGit(source) // Fetch the latest tag from the Git repository
	}

	// If the source format is not recognized, return an error
	return "", fmt.Errorf("unsupported module source format: %s", source)
}

// normalizeSource removes subdirectory information from the source.
func normalizeSource(source string) string {
	// Remove any double slashes and subdirectory information
	if strings.Contains(source, "//") {
		parts := strings.Split(source, "//")
		return parts[0]
	}
	return source
}

// isRegistryModule checks if the source is a Terraform Registry module.
func isRegistryModule(source string) bool {
	// Terraform Registry modules are in the format: namespace/name/provider
	parts := strings.Split(source, "/")
	return len(parts) >= 3 && !strings.HasPrefix(source, "git@") && !strings.HasPrefix(source, "ssh://")
}

// isGitModule checks if the source is a Git-based module.
func isGitModule(source string) bool {
	return strings.HasPrefix(source, "git@") || strings.HasPrefix(source, "ssh://") || strings.HasPrefix(source, "https://")
}
