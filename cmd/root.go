package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"tfau/lib/hcl"
	"tfau/lib/module"
	"tfau/lib/provider"
	"tfau/lib/terraform"

	"github.com/spf13/cobra"
)

var (
	files     []string
	recursive bool
	upgrades  string
	verbose   bool
	providers = true
	modules   = true
	tf        = true
)

var rootCmd = &cobra.Command{
	Use:   "tfau",
	Short: "A CLI tool to easily upgrade your Terraform modules and providers.",
	Long: `Given a Terraform project and command line parameters,
tfau upgrades each provider, module, and Terraform version in place in your HCL files.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// If no files are specified, set recursive to true
		if len(files) == 0 {
			recursive = true
		}

		// If upgrades are not specified, default to upgrading all (modules, providers, terraform)
		if upgrades == "" {
			providers = true
			modules = true
			tf = true
		} else {
			// Process upgrades if specified
			err := parseUpgradeOption(upgrades)
			if err != nil {
				return err
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Log the configuration
		log.Println("Files:", files)
		log.Println("Recursive:", recursive)
		log.Println("Modules:", modules)
		log.Println("Providers:", providers)
		log.Println("Terraform:", tf)

		// Iterate over each file and parse modules
		for _, file := range files {
			// Parse the .tf file
			file, err := hcl.ParseFile(file)
			if err != nil {
				log.Fatalf("Error parsing file: %s", err)
			}

			// Extract the content based on the schema
			content, diags := file.Body.Content(hcl.Schema)
			if diags.HasErrors() {
				log.Fatalf("Failed to decode body: %s", diags)
			}

			if modules {
				// Extract modules
				modules, err := module.Extract(content)
				if err != nil {
					log.Fatalf("Error extracting modules: %s", err)
				}
				fmt.Println("Modules:", modules)

			}
			if providers {
				// Extract providers
				providers, err := provider.Extract(content)
				if err != nil {
					log.Fatalf("Error extracting providers: %s", err)
				}
				fmt.Println("Providers:", providers)
			}
			if tf {
				// Extract Terraform version
				terraformVersion, err := terraform.Extract(content)
				if err != nil {
					log.Fatalf("Error extracting Terraform version: %s", err)
				}
				fmt.Println("Terraform Version:", terraformVersion)
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
}
