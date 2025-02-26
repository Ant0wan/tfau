package hcl

import (
	"fmt"
	"log"

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
			Type:       "data",
			LabelNames: []string{"type", "name"},
		},
	},
}

// ParseFile parses a .tf file and returns the parsed content.
func ParseFile(filename string) (*hcl.BodyContent, error) {
	// Create a new HCL parser
	parser := hclparse.NewParser()

	// Parse the .tf file
	file, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		// Filter out "unsupported block type" errors
		diags = filterUnsupportedBlockErrors(diags)
		if diags.HasErrors() {
			return nil, diagnosticsToError(diags)
		}
	}

	// Extract the content based on the schema
	content, diags := file.Body.Content(Schema)
	if diags.HasErrors() {
		// Filter out "unsupported block type" errors
		diags = filterUnsupportedBlockErrors(diags)
		if diags.HasErrors() {
			return nil, diagnosticsToError(diags)
		}
	}

	return content, nil
}

// diagnosticsToError converts hcl.Diagnostics to a standard error.
func diagnosticsToError(diags hcl.Diagnostics) error {
	if !diags.HasErrors() {
		return nil
	}
	return fmt.Errorf(diags.Error())
}

// filterUnsupportedBlockErrors removes "unsupported block type" errors from diagnostics.
func filterUnsupportedBlockErrors(diags hcl.Diagnostics) hcl.Diagnostics {
	var filtered hcl.Diagnostics
	for _, diag := range diags {
		if diag.Summary == "Unsupported block type" {
			log.Printf("Ignored unsupported block: %s", diag.Detail)
		} else {
			filtered = append(filtered, diag)
		}
	}
	return filtered
}
