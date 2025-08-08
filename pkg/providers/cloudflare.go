package providers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/dns"
	"github.com/cloudflare/cloudflare-go/v5/option"
	"github.com/cloudflare/cloudflare-go/v5/zones"
	"golang.org/x/net/publicsuffix"
)

type CloudflareDnsProvider struct {
	client *cloudflare.Client
	zone   *zones.Zone
	suffix string
}

func NewCloudflareDnsProvider(suffix string) *CloudflareDnsProvider {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		fmt.Println("CLOUDFLARE_API_TOKEN environment variable is not set")
		os.Exit(1)
	}

	client := cloudflare.NewClient(
		option.WithAPIToken(apiToken),
	)

	return &CloudflareDnsProvider{
		client: client,
		zone:   nil,
		suffix: suffix,
	}
}

func (cdp *CloudflareDnsProvider) UpdateA(ip, nodeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := cdp.getZone()
	if err != nil {
		return fmt.Errorf("failed to get zone: %w", err)
	}
	aRecordParams := dns.ARecordParam{
		Name:    cloudflare.F(nodeName),
		TTL:     cloudflare.F(dns.TTL(60)),
		Type:    cloudflare.F(dns.ARecordTypeA),
		Content: cloudflare.F(ip),
		Proxied: cloudflare.F(false),
	}

	newRecordParams := dns.RecordNewParams{
		ZoneID: cloudflare.F(cdp.zone.ID),
		Body:   aRecordParams,
	}

	options := option.WithMaxRetries(3)

	recordResponse, err := cdp.client.DNS.Records.New(ctx, newRecordParams, options)
	if err != nil {
		fmt.Printf("Failed to create A record: %v\n", err)
	}
	fmt.Printf("✅ A updated: %s -> %s\n", nodeName, ip)
	fmt.Print(recordResponse)
	return nil
}

func (cdp *CloudflareDnsProvider) getZone() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Extract the zone name from the suffix, which is expected to be a domain
	// for example: "someone.vuln.com" becomes "vuln.com"
	suffix, err := publicsuffix.EffectiveTLDPlusOne(cdp.suffix)
	if err != nil {
		log.Fatalf("failed to extract zone: %v", err)
		return err
	}

	// List zones to find the one matching the suffix
	zone, err := cdp.client.Zones.List(ctx, zones.ZoneListParams{
		Name: cloudflare.F(suffix),
	})

	if zone == nil {
		fmt.Printf("Zone not found for suffix: %s\n", suffix)
		os.Exit(1)
	}
	// make sure we have exactly one zone matching the query
	if len(zone.Result) != 1 {
		fmt.Printf("Expected one zone for suffix %s, found %d\n", suffix, len(zone.Result))
		os.Exit(1)
	}

	// Set the zone in the provider
	cdp.zone = &zone.Result[0]
	if cdp.zone == nil {
		log.Fatalf("failed to extract zone: %v", err)
		os.Exit(1)
	}
	return nil
}

// // Form new DNS record parameter to use below
// recordNewParams := dns.RecordNewParams{
// 	ZoneID: cloudflare.F(cdp.zone.ID),
// }

// // Create a new DNS record and get a response
// recordResponse, err := cdp.client.DNS.Records.New(ctx, recordNewParams, option.WithMaxRetries(3))
// if err != nil {
// 	return fmt.Errorf("failed to create DNS record: %w", err)
// }

// fmt.Printf("✅ Record created: %s -> %s\n", recordResponse.Name, recordResponse.Content)
// return nil
