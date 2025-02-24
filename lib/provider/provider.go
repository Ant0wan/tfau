package provider

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// ExtractProviders extracts provider names and their versions from the parsed content.
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
