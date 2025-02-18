package tf

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// ExtractTerraformVersion extracts the Terraform version from the parsed content.
func Extract(content *hcl.BodyContent) (string, error) {
	// Iterate over the blocks to find the terraform block
	for _, block := range content.Blocks {
		if block.Type == "terraform" {
			// Decode the attributes of the terraform block
			attrs, diags := block.Body.JustAttributes()
			if diags.HasErrors() {
				return "", fmt.Errorf("failed to decode attributes for terraform block: %s", diags)
			}

			// Get the value of the "required_version" attribute
			versionAttr, exists := attrs["required_version"]
			if !exists {
				// If the terraform block doesn't have a "required_version" attribute, skip it
				continue
			}

			versionValue, diags := versionAttr.Expr.Value(nil)
			if diags.HasErrors() {
				return "", fmt.Errorf("failed to evaluate 'required_version' expression: %s", diags)
			}

			// Return the Terraform version
			return versionValue.AsString(), nil
		}
	}

	// If no terraform block is found, return an empty string
	return "", nil
}
