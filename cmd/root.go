package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"tfau/lib/hcl"

	"github.com/spf13/cobra"
)

var (
	files     []string
	recursive bool
	upgrades  string
	verbose   bool
	providers = true
	modules   = true
	terraform = true
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
			terraform = true
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
		log.Println("Terraform:", terraform)

		if modules {
			// Iterate over each file and parse modules
			for _, file := range files {
				modules, err := hcl.ParseModules(file)
				if err != nil {
					return fmt.Errorf("error parsing modules in file '%s': %w", file, err)
				}

				// Print the module names and versions
				for name, version := range modules {
					fmt.Printf("Module: %s, Version: %s\n", name, version)
				}
			}
		}

		// Add your logic here to process the files and upgrades
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
			terraform = true
		default:
			return errors.New("unknown upgrade type: " + upgrade)
		}
	}

	// If no valid upgrades are specified, return an error
	if !modules && !providers && !terraform {
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
