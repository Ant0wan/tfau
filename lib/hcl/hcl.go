package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// ParseModules parses an HCL file and extracts module names and their versions.
func ParseModules(filename string) (map[string]string, error) {
	// Create a new HCL parser
	parser := hclparse.NewParser()

	// Parse the HCL file
	file, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL file: %s", diags)
	}

	// Decode the body into a generic hcl.Body
	body := file.Body

	// Extract the module blocks
	content, diags := body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "module",
				LabelNames: []string{"name"},
			},
		},
	})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode body: %s", diags)
	}

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
				return nil, fmt.Errorf("module '%s' does not have a 'version' attribute", moduleName)
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
