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
	recursive bool // TBD
	upgrades  string
	verbose   bool // TBD
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
		log.Println("Files:", files)

		// If no files are specified, set recursive to true
		if len(files) == 0 {
			recursive = true
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
			// Parse the .tf file and extract the content based on the schema
			content, err := hcl.ParseFile(file)
			if err != nil {
				log.Fatalf("Failed to parse file: %s", err)
			}

			log.Println("Modules:", modules)
			if modules {
				// Extract modules
				modules, err := module.Extract(content)
				if err != nil {
					log.Fatalf("Error extracting modules: %s", err)
				}
				fmt.Println("Modules:", modules)

			}

			log.Println("Providers:", providers)
			if providers {
				// Extract providers
				providers, err := provider.Extract(content)
				if err != nil {
					log.Fatalf("Error extracting providers: %s", err)
				}
				fmt.Println("Providers:", providers)
			}

			log.Println("Terraform:", tf)
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
