package module

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ModuleVersion represents the version information of a Terraform module.
type ModuleVersion struct {
	Version string `json:"version"`
}

// getLatestModuleVersion retrieves the latest version of a public Terraform module from the Terraform Registry.
func getLatestModuleVersion(moduleSource string) (string, error) {
	// Construct the API URL for the module
	apiURL, err := getModuleAPIURL(moduleSource)
	if err != nil {
		return "", fmt.Errorf("failed to construct API URL: %w", err)
	}

	// Make an HTTP GET request to the Terraform Registry API
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch module versions: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response JSON
	var versions []ModuleVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return "", fmt.Errorf("failed to parse module versions: %w", err)
	}

	// Check if any versions are available
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for module: %s", moduleSource)
	}

	// Return the latest version (first in the list)
	return versions[0].Version, nil
}

// getModuleAPIURL constructs the API URL for a given module source.
func getModuleAPIURL(moduleSource string) (string, error) {
	// Example module source: "terraform-aws-modules/vpc/aws"
	// Convert it to the API endpoint: "https://registry.terraform.io/v1/modules/terraform-aws-modules/vpc/aws/versions"
	parts := strings.Split(moduleSource, "/")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid module source format: %s (expected 'namespace/name/provider')", moduleSource)
	}

	namespace, name, provider := parts[0], parts[1], parts[2]
	return fmt.Sprintf("https://registry.terraform.io/v1/modules/%s/%s/%s/versions", namespace, name, provider), nil
}
