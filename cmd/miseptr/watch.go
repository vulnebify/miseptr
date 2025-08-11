package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vulnebify/miseptr/internal/controller"
	"github.com/vulnebify/miseptr/pkg/providers"
)

var (
	provider string
	dns      string
	suffix   string
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Monitor new Kubernetes nodes and update their PTR records",
	Run: func(cmd *cobra.Command, args []string) {
		var hostingProvider providers.HostingProvider
		var dnsProvider providers.DnsProvider

		switch provider {
		case "vultr":
			hostingProvider = providers.NewVultrProvider(suffix)
		default:
			fmt.Printf("Unsupported hosting provider: %s\n", provider)
			os.Exit(1)
		}

		switch dns {
		case "cloudflare":
			dnsProvider = providers.NewCloudflareDnsProvider(suffix)
		case "":
			dnsProvider = nil
		default:
			fmt.Printf("Unsupported DNS provider: %s\n", dns)
			os.Exit(1)
		}

		controller, err := controller.NewController(hostingProvider, dnsProvider)
		if err != nil {
			panic(fmt.Sprintf("Failed to create controller: %v", err))
		}

		controller.StartController()
	},
}

func init() {
	watchCmd.Flags().StringVar(&provider, "provider", "vultr", "Provider for PTR updates (vultr)")
	watchCmd.Flags().StringVar(&dns, "dns", "", "Provider for A updates (optional, options: cloudfalre)")
	watchCmd.Flags().StringVar(&suffix, "suffix", "", "Domain suffix for generated records")

	cobra.CheckErr(watchCmd.MarkFlagRequired("suffix"))
}
