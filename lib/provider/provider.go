package provider

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ProviderVersion represents the response from the Terraform Registry API
type ProviderVersion struct {
	Versions []struct {
		Version string `json:"version"`
	} `json:"versions"`
}

// getLatestVersion fetches the latest version of a provider from the Terraform Registry
func getLatestVersion(providerName string) (string, error) {
	// Construct the URL for the Terraform Registry API
	url := fmt.Sprintf("https://registry.terraform.io/v1/providers/%s/versions", providerName)

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch provider versions: %s", err)
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var providerVersion ProviderVersion
	if err := json.NewDecoder(resp.Body).Decode(&providerVersion); err != nil {
		return "", fmt.Errorf("failed to decode provider versions: %s", err)
	}

	// Check if there are any versions available
	if len(providerVersion.Versions) == 0 {
		return "", fmt.Errorf("no versions found for provider '%s'", providerName)
	}

	// Return the latest version
	return providerVersion.Versions[0].Version, nil
}

// Extract extracts provider names and their versions from the parsed content.
func Extract(content *hcl.BodyContent) (map[string]string, error) {
	// Create a map to store provider names and versions
	providers := make(map[string]string)

	// Iterate over the blocks to find all provider blocks
	for _, block := range content.Blocks {
		if block.Type == "provider" {
			// Handle provider blocks
			providerName := block.Labels[0]

			// Decode the attributes of the provider block
			attrs, diags := block.Body.JustAttributes()
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to decode attributes for provider '%s': %s", providerName, diags)
			}

			// Get the value of the "version" attribute
			versionAttr, exists := attrs["version"]
			var versionValue string
			if exists {
				// If the version attribute exists, evaluate it
				value, diags := versionAttr.Expr.Value(nil)
				if diags.HasErrors() {
					return nil, fmt.Errorf("failed to evaluate 'version' expression for provider '%s': %s", providerName, diags)
				}
				versionValue = value.AsString()
			} else {
				// If the version attribute does not exist, fetch the latest version
				latestVersion, err := getLatestVersion(providerName)
				if err != nil {
					return nil, fmt.Errorf("failed to get latest version for provider '%s': %s", providerName, err)
				}
				versionValue = latestVersion
			}

			// Store the provider name and version in the map
			providers[providerName] = versionValue
		} else if block.Type == "terraform" {
			// Handle the terraform block
			attrs, diags := block.Body.JustAttributes()
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to decode attributes for terraform block: %s", diags)
			}

			// Check if the terraform block has a required_providers attribute
			requiredProvidersAttr, exists := attrs["required_providers"]
			if exists {
				// Decode the required_providers attribute
				expr, ok := requiredProvidersAttr.Expr.(*hclsyntax.ObjectConsExpr)
				if !ok {
					return nil, fmt.Errorf("required_providers attribute is not an object")
				}

				// Iterate over the key-value pairs in the required_providers object
				for _, item := range expr.Items {
					key, diags := item.KeyExpr.Value(nil)
					if diags.HasErrors() {
						return nil, fmt.Errorf("failed to evaluate key in required_providers: %s", diags)
					}

					providerName := key.AsString()

					// Check if the value is a string (simple version constraint) or an object (detailed configuration)
					value, diags := item.ValueExpr.Value(nil)
					if diags.HasErrors() {
						return nil, fmt.Errorf("failed to evaluate value in required_providers: %s", diags)
					}

					if value.Type().IsObjectType() {
						// If the value is an object, decode it to get the version
						obj := value.AsValueMap()
						versionVal, exists := obj["version"]
						if !exists {
							// If no version is specified, fetch the latest version
							latestVersion, err := getLatestVersion(providerName)
							if err != nil {
								return nil, fmt.Errorf("failed to get latest version for provider '%s': %s", providerName, err)
							}
							providers[providerName] = latestVersion
						} else {
							// Use the specified version
							providers[providerName] = versionVal.AsString()
						}
					} else {
						// If the value is a string, treat it as a version constraint
						versionConstraint := value.AsString()
						if versionConstraint == "" {
							// If the version constraint is empty, fetch the latest version
							latestVersion, err := getLatestVersion(providerName)
							if err != nil {
								return nil, fmt.Errorf("failed to get latest version for provider '%s': %s", providerName, err)
							}
							providers[providerName] = latestVersion
						} else {
							// Otherwise, use the provided version constraint
							providers[providerName] = versionConstraint
						}
					}
				}
			}
		}
	}

	return providers, nil
}
