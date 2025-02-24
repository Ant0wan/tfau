// Main module logic and entry point

package module

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/hcl/v2"
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
				latestVersion, err := getLatestModuleVersion(moduleInfo["source"])
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
