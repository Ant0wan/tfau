package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

func Extract(content *hcl.BodyContent) (string, error) {
	// Iterate over the blocks to find the terraform block
	for _, block := range content.Blocks {
		if block.Type == "terraform" {
			// Define the schema for the terraform block
			schema := &hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: "required_version",
					},
				},
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "required_providers",
					},
				},
			}

			// Decode the terraform block content
			terraformContent, diags := block.Body.Content(schema)
			if diags.HasErrors() {
				return "", fmt.Errorf("failed to decode terraform block: %s", diags)
			}

			// Get the value of the "required_version" attribute
			if attr, exists := terraformContent.Attributes["required_version"]; exists {
				versionValue, diags := attr.Expr.Value(nil)
				if diags.HasErrors() {
					return "", fmt.Errorf("failed to evaluate 'required_version' expression: %s", diags)
				}
				return versionValue.AsString(), nil
			}

			// If no "required_version" attribute is found, return an empty string
			return "", nil
		}
	}

	// If no terraform block is found, return an empty string
	return "", nil
}
