package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"tfau/lib/hcl"
	"tfau/lib/module"
	"tfau/lib/provider"
	"tfau/lib/terraform"

	"github.com/spf13/cobra"
)

var (
	files            []string
	recursive        bool // TBD
	upgrades         string
	verbose          bool // TBD
	providers        = true
	modules          = true
	tf               = true
	terraformVersion string // Desired Terraform version
)

// findTFFiles recursively finds all .tf files in the given directory
func findTFFiles(dir string) ([]string, error) {
	var tfFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			tfFiles = append(tfFiles, path)
		}
		return nil
	})
	return tfFiles, err
}

var rootCmd = &cobra.Command{
	Use:   "tfau",
	Short: "A CLI tool to easily upgrade your Terraform modules and providers.",
	Long: `Given a Terraform project and command line parameters,
tfau upgrades each provider, module, and Terraform version in place in your HCL files.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Files:", files)

		// If no files are specified, find all .tf files recursively
		if len(files) == 0 {
			recursive = true
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %v", err)
			}
			tfFiles, err := findTFFiles(cwd)
			if err != nil {
				return fmt.Errorf("failed to find .tf files: %v", err)
			}
			files = tfFiles
		}
		log.Println("Recursive:", recursive)

		// If upgrades are not specified, default to upgrading all (modules, providers, terraform)
		// otherwise process specified upgrades only
		if upgrades != "" {
			providers, modules, tf = false, false, false
			err := parseUpgradeOption(upgrades)
			if err != nil {
				return err
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Iterate over each file and parse modules
		for _, file := range files {
			log.Printf("Processing file: %s\n", file)

			// Parse the .tf file and extract the content based on the schema
			content, err := hcl.ParseFile(file)
			if err != nil {
				log.Printf("Error parsing file %s: %v. Skipping file.\n", file, err)
				continue // Skip to the next file
			}

			log.Println("Modules:", modules)
			if modules {
				// Extract modules
				modules, err := module.Extract(content)
				if err != nil {
					log.Printf("Error extracting modules from file %s: %v. Skipping modules.\n", file, err)
				} else {
					// Create a map to store the latest versions
					latestVersions := make(map[string]string)

					// Fetch the latest version for each module
					for name, info := range modules {
						source := info["source"]
						if source != "" {
							latestVersion, err := module.GetLatestModuleVersion(source)
							if err != nil {
								log.Printf("Warning: Failed to retrieve latest version for module '%s' in file %s: %v\n", name, file, err)
							} else {
								latestVersions[name] = latestVersion
								fmt.Printf("Module: %s, Current Version: %s, Latest Version: %s\n", name, info["version"], latestVersion)
							}
						}
					}

					log.Printf("Latest versions to update in file %s: %v", file, latestVersions)

					// Update the module versions in the file
					err = module.UpdateModuleVersions(file, latestVersions)
					if err != nil {
						log.Printf("Failed to update module versions in file %s: %v\n", file, err)
					} else {
						log.Println("Updated module versions in the file.")
					}
				}
			}

			log.Println("Providers:", providers)
			if providers {
				// Extract providers and their latest versions
				currentVersions, latestVersions, err := provider.ExtractWithLatestVersions(content)
				if err != nil {
					log.Printf("Error extracting providers from file %s: %v. Skipping providers.\n", file, err)
				} else {
					if len(currentVersions) == 0 {
						log.Println("No provider blocks found in the file.")
					} else {
						// Print current and latest versions
						for name, version := range currentVersions {
							latestVersion := latestVersions[name]
							fmt.Printf("Provider: %s, Current Version: %s, Latest Version: %s\n", name, version, latestVersion)
						}

						// Update the provider versions in the file
						err := provider.UpdateProviderVersions(file, latestVersions)
						if err != nil {
							log.Printf("Failed to update provider versions in file %s: %v\n", file, err)
						} else {
							log.Println("Updated provider versions in the file.")
						}
					}
				}
			}

			log.Println("Terraform:", tf)
			if tf {
				if terraformVersion != "" {
					log.Printf("Terraform version specified: %s\n", terraformVersion)
					// Update the required_version in the file with the specified version
					err := terraform.UpdateRequiredVersion(file, terraformVersion)
					if err != nil {
						log.Printf("Failed to update required_version in file %s: %v\n", file, err)
					} else {
						log.Printf("Updated required_version to %s in the file.\n", terraformVersion)
					}
				} else {
					// Extract Terraform version and fetch the latest version
					currentVersion, latestVersion, err := terraform.ExtractWithLatestVersion(content)
					if err != nil {
						log.Printf("Error extracting Terraform version from file %s: %v. Skipping Terraform version update.\n", file, err)
					} else {
						if currentVersion == "" {
							log.Println("No Terraform version specified in the file.")
						} else {
							fmt.Printf("Terraform Version: %s, Latest Version: %s\n", currentVersion, latestVersion)

							// Update the required_version in the file with the latest version
							err := terraform.UpdateRequiredVersion(file, latestVersion)
							if err != nil {
								log.Printf("Failed to update required_version in file %s: %v\n", file, err)
							} else {
								log.Printf("Updated required_version to %s in the file.\n", latestVersion)
							}
						}
					}
				}
			}
		}

		return nil
	},
}

func parseUpgradeOption(upgrades string) error {
	upgradesList := strings.Split(upgrades, ",")
	for _, upgrade := range upgradesList {
		switch upgrade {
		case "modules":
			modules = true
		case "providers":
			providers = true
		case "terraform":
			tf = true
		default:
			return errors.New("unknown upgrade type: " + upgrade)
		}
	}

	// If no valid upgrades are specified, return an error
	if !modules && !providers && !tf {
		return errors.New("at least one upgrade type must be specified")
	}

	return nil
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Files flag (optional)
	rootCmd.Flags().StringArrayVarP(&files, "file", "f", []string{}, "HCL file(s) to be updated")

	// Upgrades flag (optional)
	rootCmd.Flags().StringVar(&upgrades, "upgrades", "", "Comma-separated list of upgrades (modules, providers, terraform)")

	// Verbose flag (optional)
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Terraform version flag (optional)
	rootCmd.Flags().StringVar(&terraformVersion, "terraform-version", "", "Desired Terraform version to update to (e.g., '~>1.9')")
}
