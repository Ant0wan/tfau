package cmd

import (
	"errors"
	"log"
	"os"
	"strings"

	//"tfau/lib/hcl"

	"github.com/spf13/cobra"
)

var (
	files     []string
	recursive bool
	upgrades  string
	verbose   bool
	providers bool
	modules   bool
	terraform bool
	rootCmd   = &cobra.Command{
		Use:   "tfau",
		Short: " A CLI tool to easily upgrade your Terraform modules and providers.",
		Long: `Given a Terraform project and command line parameters,
tfau upgrade each provider, module and terraform version in place in your hcl files.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(files) == 0 { // example
				err := cmd.Help()
				if err != nil {
					log.Fatal(err)
				}
				os.Exit(1)
			}
			return nil
		},
		//	Run: func(cmd *cobra.Command, args []string) {
		//		scrap(addrs, format)
		//	},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := processUpgrades(upgrades)
			if err != nil {
				return err
			}
			return nil
		},
	}
)

func processUpgrades(upgrades string) error {
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

	if !modules && !providers && !terraform {
		return errors.New("at least one upgrade type must be specified")
	}

	log.Println("Modules:", modules)
	log.Println("Providers:", providers)
	log.Println("Terraform:", terraform)

	return nil
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringArrayVarP(&files, "file", "f", []string{}, "hcl file to be updated")
	rootCmd.Flags().StringVar(&upgrades, "upgrades", "", "Comma-separated list of upgrades (modules, providers, terraform)")
}
