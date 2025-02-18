package cmd

import (
	"log"
	"os"
	"sync"

	"tfau/lib"

	"github.com/spf13/cobra"
)

var (
)

func Execute() {
	err := rootCmd.Execute()
	if err != nill {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringArrayVarP(&addr, )
}
