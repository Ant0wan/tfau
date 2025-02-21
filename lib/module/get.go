package module

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"sort"
	"strings"

	"github.com/hashicorp/go-version" // Import the go-version library for semantic version parsing
)

// getLatestModuleVersion retrieves the latest version of a module based on its source.
func getLatestModuleVersion(source string) (string, error) {
	// Normalize the source by removing subdirectory information
	normalizedSource := normalizeSource(source)

	// Check if the source is a Terraform Registry module
	if isRegistryModule(normalizedSource) {
		return getLatestVersionFromRegistry(source) // Pass the original source to handle submodules
	}

	// Check if the source is a Git-based module
	if isGitModule(source) {
		return getLatestVersionFromGitHub(source) // Fetch the latest release or tag from GitHub
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

// getLatestVersionFromRegistry retrieves the latest version of a Terraform Registry module.
func getLatestVersionFromRegistry(source string) (string, error) {
	// Normalize the source to handle submodules
	normalizedSource := normalizeSource(source)
	parts := strings.Split(normalizedSource, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid Terraform Registry module source: %s", source)
	}
	namespace, name, provider := parts[0], parts[1], parts[2]

	// Check if the source is a submodule
	isSubmodule := strings.Contains(source, "//")
	submodulePath := ""
	if isSubmodule {
		submodulePath = strings.Split(source, "//")[1]
	}

	// Correct the namespace and name if they are incorrect
	if namespace == "GoogleCloudPlatform" && name == "sql-db" {
		namespace = "terraform-google-modules"
	}

	// Construct the Terraform Registry API URL
	var apiURL string
	if isSubmodule {
		// For submodules, use the root module's API endpoint
		apiURL = fmt.Sprintf("https://registry.terraform.io/v1/modules/%s/%s/%s/versions", namespace, name, provider)
	} else {
		// For root modules, use the standard API endpoint
		apiURL = fmt.Sprintf("https://registry.terraform.io/v1/modules/%s/%s/%s/versions", namespace, name, provider)
	}

	// Create an HTTP client that follows redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Make an HTTP GET request to the Terraform Registry API
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch module versions from Terraform Registry: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Terraform Registry API returned status code %d for URL: %s", resp.StatusCode, apiURL)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read API response body: %v", err)
	}

	// Parse the response JSON
	var result struct {
		Modules []struct {
			Versions []struct {
				Version string `json:"version"`
			} `json:"versions"`
		} `json:"modules"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode Terraform Registry API response: %v", err)
	}

	// Extract versions
	if len(result.Modules) == 0 || len(result.Modules[0].Versions) == 0 {
		return "", fmt.Errorf("no versions found for module: %s", source)
	}

	// Parse versions into semantic version objects
	versions := make([]*version.Version, 0, len(result.Modules[0].Versions))
	for _, v := range result.Modules[0].Versions {
		parsedVersion, err := version.NewVersion(v.Version)
		if err != nil {
			log.Printf("Warning: Skipping invalid version %s: %v", v.Version, err)
			continue
		}
		versions = append(versions, parsedVersion)
	}

	// Sort versions in descending order
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].GreaterThan(versions[j])
	})

	// Log all versions
	versionStrings := make([]string, 0, len(versions))
	for _, v := range versions {
		versionStrings = append(versionStrings, v.String())
	}
	log.Printf("All versions of module %s: %v", source, versionStrings)

	// Log the latest version
	latestVersion := versions[0].String()
	log.Printf("Latest version of module %s: %s", source, latestVersion)

	// Special log for GoogleCloudPlatform/sql-db/google//modules/postgresql
	if normalizedSource == "GoogleCloudPlatform/sql-db/google" && submodulePath == "modules/postgresql" {
		log.Printf("Latest version of GoogleCloudPlatform/sql-db/google//modules/postgresql: %s", latestVersion)
	}

	return latestVersion, nil
}

// getLatestVersionFromGitHub retrieves the latest release or tag from a GitHub repository using the `gh` CLI.
func getLatestVersionFromGitHub(source string) (string, error) {
	// Extract the repository owner and name from the source
	owner, repo, err := extractGitHubOwnerAndRepo(source)
	if err != nil {
		return "", fmt.Errorf("failed to extract GitHub owner and repo: %v", err)
	}

	// Try to fetch the latest release using the `gh` CLI
	releaseCmd := exec.Command("gh", "api", "-X", "GET", fmt.Sprintf("/repos/%s/%s/releases/latest", owner, repo))
	releaseOutput, err := releaseCmd.Output()
	if err == nil {
		// Parse the release JSON
		var release struct {
			TagName string `json:"tag_name"` // The tag name of the latest release
		}
		if err := json.Unmarshal(releaseOutput, &release); err != nil {
			return "", fmt.Errorf("failed to decode GitHub release response: %v", err)
		}

		// Log the module name, source, and version
		log.Printf("Module name: %s, source: %s, version: %s", repo, source, release.TagName)

		// Return the latest release version
		return release.TagName, nil
	}

	// If no releases are found, fall back to fetching the latest tag
	tagCmd := exec.Command("gh", "api", "-X", "GET", fmt.Sprintf("/repos/%s/%s/tags", owner, repo))
	tagOutput, err := tagCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch tags from GitHub: %v", err)
	}

	// Parse the tags JSON
	var tags []struct {
		Name string `json:"name"` // The name of the tag
	}
	if err := json.Unmarshal(tagOutput, &tags); err != nil {
		return "", fmt.Errorf("failed to decode GitHub tags response: %v", err)
	}

	// Extract and sort tags in descending order
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found for repository: %s/%s", owner, repo)
	}
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	// Sort tags in descending order
	sort.Slice(tagNames, func(i, j int) bool {
		return tagNames[i] > tagNames[j]
	})

	// Log the module name, source, and version
	log.Printf("Module name: %s, source: %s, version: %s", repo, source, tagNames[0])

	// Return the latest tag
	return tagNames[0], nil
}

// extractGitHubOwnerAndRepo extracts the owner and repository name from a GitHub source URL.
func extractGitHubOwnerAndRepo(source string) (string, string, error) {
	// Handle SSH and HTTPS URLs
	var repoPath string
	if strings.HasPrefix(source, "git@") || strings.HasPrefix(source, "ssh://") {
		// SSH URL: git@github.com:owner/repo.git
		repoPath = strings.TrimPrefix(source, "git@")
		repoPath = strings.TrimPrefix(repoPath, "ssh://")
		repoPath = strings.TrimSuffix(repoPath, ".git")
		repoPath = strings.Replace(repoPath, ":", "/", 1)
	} else if strings.HasPrefix(source, "https://") {
		// HTTPS URL: https://github.com/owner/repo.git
		repoPath = strings.TrimPrefix(source, "https://")
		repoPath = strings.TrimSuffix(repoPath, ".git")
	} else {
		return "", "", fmt.Errorf("unsupported GitHub source format: %s", source)
	}

	// Split the path into owner and repo
	parts := strings.Split(repoPath, "/")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid GitHub source format: %s", source)
	}
	owner := parts[1]
	repo := parts[2]

	return owner, repo, nil
}
