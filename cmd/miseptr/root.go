package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulnebify/miseptr/internal"
)

const asciiArt = `
 __  __  ___  ____   _____  ____   _____  ____  
|  \/  ||_ _|/ ___| | ____||  _ \ |_   _||  _ \ 
| |\/| | | | \___ \ |  _|  | |_) |  | |  | |_) |
| |  | | | |  ___) || |___ |  __/   | |  |  _ < 
|_|  |_||___||____/ |_____||_|      |_|  |_| \_\
                                                                                                                   
`

var rootCmd = &cobra.Command{
	Use:     "miseptr",
	Short:   "Automated PTR record updater for Kubernetes nodes",
	Version: internal.Version,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(asciiArt)
		_ = cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.AddCommand(watchCmd)
}
