package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vulnebify/miseptr/internal/controller"
	"github.com/vulnebify/miseptr/pkg/providers"
)

var (
	providerName string
	ptrSuffix    string
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Monitor new Kubernetes nodes and update their PTR records",
	Run: func(cmd *cobra.Command, args []string) {
		var provider providers.HostingProvider

		switch providerName {
		case "vultr":
			hostingProvider = providers.NewVultrProvider(ptrSuffix)
		default:
			fmt.Printf("Unsupported provider: %s\n", providerName)
			os.Exit(1)
		}
		switch dnsProviderName {
		case "cloudflare":
			dnsProvider
		}

		controller, err := controller.NewController(provider)
		if err != nil {
			panic(fmt.Sprintf("Failed to create controller: %v", err))
		}

		controller.StartController()
	},
}

func init() {
	watchCmd.Flags().StringVar(&dnsProviderName, "suffix", "", "Domain suffix for generated PTR records")
	watchCmd.Flags().StringVar(&hostingProviderName, "hosting-provider", "vultr", "Provider for PTR updates (vultr)")
	watchCmd.Flags().StringVar(&ptrSuffix, "suffix", "", "Domain suffix for generated PTR records")

	cobra.CheckErr(watchCmd.MarkFlagRequired("suffix"))
}
