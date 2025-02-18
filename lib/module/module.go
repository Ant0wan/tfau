package module

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// ExtractModules extracts module names and their versions from the parsed content.
func Extract(content *hcl.BodyContent) (map[string]string, error) {
	// Create a map to store module names and versions
	modules := make(map[string]string)

	// Iterate over the blocks to find all module blocks
	for _, block := range content.Blocks {
		if block.Type == "module" {
			// Get the module name
			moduleName := block.Labels[0]

			// Decode the attributes of the module block
			attrs, diags := block.Body.JustAttributes()
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to decode attributes for module '%s': %s", moduleName, diags)
			}

			// Get the value of the "version" attribute
			versionAttr, exists := attrs["version"]
			if !exists {
				// If the module doesn't have a "version" attribute, skip it
				continue
			}

			versionValue, diags := versionAttr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to evaluate 'version' expression for module '%s': %s", moduleName, diags)
			}

			// Store the module name and version in the map
			modules[moduleName] = versionValue.AsString()
		}
	}

	return modules, nil
}
