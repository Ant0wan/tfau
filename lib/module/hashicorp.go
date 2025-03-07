package module

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
)

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
