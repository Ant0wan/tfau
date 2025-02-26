package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ProviderLatestVersion represents the latest version of a provider from the Terraform Registry.
type ProviderLatestVersion struct {
	Version string `json:"version"`
}

// ProviderVersions represents the response from the Terraform Registry API for provider versions.
type ProviderVersions struct {
	Versions []struct {
		Version string `json:"version"`
	} `json:"versions"`
}

// Extract extracts provider names and their versions from the parsed content.
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
		} else if block.Type == "terraform" {
			// Handle the `terraform` block to extract `required_providers`
			body, ok := block.Body.(*hclsyntax.Body)
			if !ok {
				return nil, fmt.Errorf("failed to parse terraform block body")
			}

			// Look for the `required_providers` block
			for _, innerBlock := range body.Blocks {
				if innerBlock.Type == "required_providers" {
					// Decode the attributes of the `required_providers` block
					attrs, diags := innerBlock.Body.JustAttributes()
					if diags.HasErrors() {
						return nil, fmt.Errorf("failed to decode attributes for required_providers block: %s", diags)
					}

					// Extract provider versions from the attributes
					for providerName, attr := range attrs {
						// Evaluate the attribute value
						value, diags := attr.Expr.Value(nil)
						if diags.HasErrors() {
							return nil, fmt.Errorf("failed to evaluate expression for provider '%s': %s", providerName, diags)
						}

						// Handle both object and string formats
						if value.Type().IsObjectType() {
							// Object format: { source = "hashicorp/google", version = "6.22.0" }
							providerMap := value.AsValueMap()
							sourceValue, ok := providerMap["source"]
							if !ok {
								return nil, fmt.Errorf("provider '%s' is missing 'source' attribute", providerName)
							}
							source := sourceValue.AsString()

							versionValue, ok := providerMap["version"]
							if !ok {
								return nil, fmt.Errorf("provider '%s' is missing 'version' attribute", providerName)
							}
							version := versionValue.AsString()

							// Store the provider source and version in the map
							providers[source] = version
						} else if value.Type().IsPrimitiveType() {
							// String format: ">=4.84"
							version := value.AsString()
							// Assume the provider is from the "hashicorp" namespace
							providers["hashicorp/"+providerName] = version
						} else {
							return nil, fmt.Errorf("provider '%s' has an unsupported format", providerName)
						}
					}
				}
			}
		}
	}

	return providers, nil
}

// GetLatestVersion fetches the latest version of a provider from the Terraform Registry.
func GetLatestVersion(providerName string) (string, error) {
	// Construct the URL for the Terraform Registry API
	url := fmt.Sprintf("https://registry.terraform.io/v1/providers/%s/versions", providerName)
	log.Printf("Fetching versions for provider: %s (URL: %s)", providerName, url) // Debug log

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch versions for provider '%s': %v", providerName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch versions for provider '%s': %s", providerName, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body for provider '%s': %v", providerName, err)
	}

	var versions ProviderVersions
	if err := json.Unmarshal(body, &versions); err != nil {
		return "", fmt.Errorf("failed to unmarshal response for provider '%s': %v", providerName, err)
	}

	if len(versions.Versions) == 0 {
		return "", fmt.Errorf("no versions found for provider '%s'", providerName)
	}

	// The latest version is the first item in the list
	latestVersion := versions.Versions[0].Version
	return latestVersion, nil
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
