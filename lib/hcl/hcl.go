package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

var Schema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "module",
			LabelNames: []string{"name"},
		},
		{
			Type:       "provider",
			LabelNames: []string{"name"},
		},
		{
			Type:       "terraform",
			LabelNames: []string{},
		},
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
	},
}

// ParseFile parses a .tf file and returns the parsed content.
func ParseFile(filename string) (*hcl.File, error) {
	// Create a new HCL parser
	parser := hclparse.NewParser()

	// Parse the .tf file
	file, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse .tf file: %s", diags)
	}

	return file, nil
}
