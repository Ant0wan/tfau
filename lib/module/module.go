package module

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// Extract extracts module names, their sources, and versions from the parsed content.
func Extract(content *hcl.BodyContent) (map[string]map[string]string, error) {
	// Create a map to store module information (name -> {source, version})
	modules := make(map[string]map[string]string)

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

			// Create a map to store source and version for the current module
			moduleInfo := make(map[string]string)

			// Get the value of the "source" attribute
			sourceAttr, exists := attrs["source"]
			if !exists {
				return nil, fmt.Errorf("module '%s' is missing the 'source' attribute", moduleName)
			}

			sourceValue, diags := sourceAttr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to evaluate 'source' expression for module '%s': %s", moduleName, diags)
			}

			source := sourceValue.AsString()

			// Handle Git SSH URLs (e.g., git@github.com:user/repo.git)
			if strings.HasPrefix(source, "git@") {
				source = strings.Replace(source, ":", "/", 1) // Replace the first colon with a slash
				source = "ssh://" + source                    // Prepend with ssh://
			}

			// Parse the source URL to extract the version from the query parameter if it exists
			if strings.Contains(source, "?") {
				parsedURL, err := url.Parse(source)
				if err != nil {
					return nil, fmt.Errorf("failed to parse source URL for module '%s': %s", moduleName, err)
				}

				// Extract the version from the query parameter
				version := parsedURL.Query().Get("ref")
				if version != "" {
					moduleInfo["version"] = version
				}

				// Remove the query parameter from the source
				source = strings.Split(source, "?")[0]
			}

			moduleInfo["source"] = source

			// Get the value of the "version" attribute (if it exists)
			versionAttr, exists := attrs["version"]
			if exists {
				versionValue, diags := versionAttr.Expr.Value(nil)
				if diags.HasErrors() {
					return nil, fmt.Errorf("failed to evaluate 'version' expression for module '%s': %s", moduleName, diags)
				}
				moduleInfo["version"] = versionValue.AsString()
			} else if _, ok := moduleInfo["version"]; !ok {
				// If the module doesn't have a "version" attribute, set it to an empty string
				moduleInfo["version"] = ""
			}

			// Store the module information in the map
			modules[moduleName] = moduleInfo

			log.Printf("Module name: %s, source: %s, version: %s\n", moduleName, moduleInfo["source"], moduleInfo["version"])

			// Get the latest version of the module (if needed)
			if moduleInfo["source"] != "" {
				latestVersion, err := GetLatestModuleVersion(moduleInfo["source"])
				if err != nil {
					log.Printf("Warning: Failed to retrieve latest version for module '%s': %v\n", moduleName, err)
				} else {
					fmt.Printf("Latest version of %s: %s\n", moduleInfo["source"], latestVersion)
				}
			}
		}
	}

	return modules, nil
}

// UpdateModuleVersions updates the module versions in the HCL content and writes it back to the file.
// It updates both the version attribute and the ref parameter in the source attribute.
func UpdateModuleVersions(filename string, latestVersions map[string]string) error {
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

	// Iterate over the blocks to find module blocks
	body := file.Body()
	for _, block := range body.Blocks() {
		if block.Type() == "module" {
			// Get the module name
			moduleName := block.Labels()[0]

			// Check if the module has a latest version
			if latestVersion, exists := latestVersions[moduleName]; exists {
				// Update the version attribute if it exists
				if attr := block.Body().GetAttribute("version"); attr != nil {
					log.Printf("Updating module '%s' to version '%s'", moduleName, latestVersion)
					block.Body().SetAttributeValue("version", cty.StringVal(latestVersion))
				}

				// Update the ref parameter in the source attribute if it exists
				sourceAttr := block.Body().GetAttribute("source")
				if sourceAttr != nil {
					// Get the source value as a string
					sourceValue := string(sourceAttr.Expr().BuildTokens(nil).Bytes())

					if strings.Contains(sourceValue, "?ref=") {
						// Update the ref in the source attribute
						newSource := strings.Split(sourceValue, "?ref=")[0] + "?ref=" + latestVersion
						block.Body().SetAttributeValue("source", cty.StringVal(newSource))
						log.Printf("Updated source attribute for module '%s' to version '%s'", moduleName, latestVersion)
					} else if strings.HasPrefix(sourceValue, "git@") || strings.HasPrefix(sourceValue, "ssh://") {
						// If the source is a Git URL without a ref, add the ref parameter
						newSource := sourceValue + "?ref=" + latestVersion
						block.Body().SetAttributeValue("source", cty.StringVal(newSource))
						log.Printf("Added ref to source attribute for module '%s': %s", moduleName, newSource)
					}
				}
			}
		}
	}

	// Write the updated content back to the file
	if err := ioutil.WriteFile(filename, file.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	log.Printf("Successfully updated module versions in file: %s", filename)
	return nil
}
