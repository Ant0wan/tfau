package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/hcl/v2"
)

// ProviderLatestVersion represents the latest version of a provider from the Terraform Registry.
type ProviderLatestVersion struct {
	Version string `json:"version"`
}

// Extract extracts provider names and their versions from the parsed content.
func Extract(content *hcl.BodyContent) (map[string]string, error) {
	// Create a map to store provider names and versions
	providers := make(map[string]string)

	// Iterate over the blocks to find all provider blocks
	for _, block := range content.Blocks {
		if block.Type == "provider" {
			// Get the provider name
			providerName := block.Labels[0]

			// Decode the attributes of the provider block
			attrs, diags := block.Body.JustAttributes()
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to decode attributes for provider '%s': %s", providerName, diags)
			}

			// Get the value of the "version" attribute
			versionAttr, exists := attrs["version"]
			if !exists {
				// If the provider doesn't have a "version" attribute, skip it
				continue
			}

			versionValue, diags := versionAttr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to evaluate 'version' expression for provider '%s': %s", providerName, diags)
			}

			// Store the provider name and version in the map
			providers[providerName] = versionValue.AsString()
		}
	}

	return providers, nil
}

// GetLatestVersion fetches the latest version of a provider from the Terraform Registry.
func GetLatestVersion(providerName string) (string, error) {
	url := fmt.Sprintf("https://registry.terraform.io/v1/providers/%s/latest", providerName)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version for provider '%s': %v", providerName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch latest version for provider '%s': %s", providerName, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body for provider '%s': %v", providerName, err)
	}

	var latestVersion ProviderLatestVersion
	if err := json.Unmarshal(body, &latestVersion); err != nil {
		return "", fmt.Errorf("failed to unmarshal response for provider '%s': %v", providerName, err)
	}

	return latestVersion.Version, nil
}

// ExtractWithLatestVersions extracts provider names, their current versions, and their latest versions.
func ExtractWithLatestVersions(content *hcl.BodyContent) (map[string]string, map[string]string, error) {
	// Extract current versions
	providers, err := Extract(content)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("Extracted providers: %v", providers) // Debug log

	// Create a map to store the latest versions
	latestVersions := make(map[string]string)

	// Fetch the latest version for each provider
	for name := range providers {
		latestVersion, err := GetLatestVersion(name)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get latest version for provider '%s': %v", name, err)
		}
		latestVersions[name] = latestVersion
	}

	log.Printf("Latest versions: %v", latestVersions) // Debug log

	return providers, latestVersions, nil
}
