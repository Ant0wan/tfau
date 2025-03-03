package terraform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// TerraformReleases represents the response from the Terraform Releases API.
type TerraformReleases struct {
	Versions map[string]struct {
		Version string `json:"version"`
	} `json:"versions"`
}

// Extract extracts the required Terraform version from the parsed content.
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

// GetLatestVersion fetches the latest Terraform version from the Terraform Releases API.
func GetLatestVersion() (string, error) {
	// Construct the URL for the Terraform Releases API
	url := "https://releases.hashicorp.com/terraform/index.json"
	log.Printf("Fetching latest Terraform version (URL: %s)", url) // Debug log

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Terraform versions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch Terraform versions: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var releases TerraformReleases
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(releases.Versions) == 0 {
		return "", fmt.Errorf("no Terraform versions found")
	}

	// Parse versions and sort them
	var versionList []*version.Version
	for versionStr := range releases.Versions {
		parsedVersion, err := version.NewVersion(versionStr)
		if err != nil {
			log.Printf("Failed to parse version '%s': %v", versionStr, err)
			continue
		}
		versionList = append(versionList, parsedVersion)
	}

	if len(versionList) == 0 {
		return "", fmt.Errorf("no valid Terraform versions found")
	}

	// Sort versions in descending order
	sort.Sort(sort.Reverse(version.Collection(versionList)))

	// The latest version is the first item in the sorted list
	latestVersion := versionList[0].String()
	return latestVersion, nil
}

// ExtractWithLatestVersion extracts the current Terraform version and fetches the latest version.
func ExtractWithLatestVersion(content *hcl.BodyContent) (string, string, error) {
	// Extract current version
	currentVersion, err := Extract(content)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract Terraform version: %v", err)
	}

	log.Printf("Extracted Terraform version: %s", currentVersion) // Debug log

	// Fetch the latest version
	latestVersion, err := GetLatestVersion()
	if err != nil {
		return "", "", fmt.Errorf("failed to get latest Terraform version: %v", err)
	}

	log.Printf("Latest Terraform version: %s", latestVersion) // Debug log

	return currentVersion, latestVersion, nil
}

// UpdateRequiredVersion updates the required_version in the HCL content and writes it back to the file.
func UpdateRequiredVersion(filename string, newVersion string) error {
	// Read the file content
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Parse the HCL content
	file, diags := hclwrite.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("failed to parse HCL content: %s", diags)
	}

	// Find the terraform block
	body := file.Body()
	for _, block := range body.Blocks() {
		if block.Type() == "terraform" {
			// Update the required_version attribute
			block.Body().SetAttributeValue("required_version", cty.StringVal(newVersion))
			break
		}
	}

	// Write the updated content back to the file
	if err := ioutil.WriteFile(filename, file.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}
