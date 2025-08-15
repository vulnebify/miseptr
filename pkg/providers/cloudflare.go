package providers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
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

	cdp := &CloudflareDnsProvider{
		client: client,
		zone:   nil,
		suffix: suffix,
	}
	err := cdp.getZone()
	if err != nil {
		fmt.Printf("Failed to get zone: %v\n", err)
		fmt.Printf("Failed to create CloudflareDnsProvider: %v\n", err)
		return nil
	}
	return cdp
}

func (cdp *CloudflareDnsProvider) UpdateA(ip, nodeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aRecord, err := cdp.formatARecord(cdp.suffix, nodeName)
	if err != nil {
		fmt.Printf("Failed to format A record: %v\n", err)
		return err
	}

	aRecordParams := dns.ARecordParam{
		Name:    cloudflare.F(aRecord),
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

	_, err = cdp.client.DNS.Records.New(ctx, newRecordParams, options)
	if err != nil {
		if strings.Contains(err.Error(), "An identical record already exists.") {
			fmt.Printf("✅ A already exists: %s -> %s\n", aRecord, ip)
		}
		fmt.Printf("Failed to create A record: %v\n", err)
		return err
	}
	fmt.Printf("✅ A updated: %s -> %s\n", aRecord, ip)
	return nil
}

func (cdp *CloudflareDnsProvider) formatARecord(suffix, nodeName string) (string, error) {
	suffix = strings.TrimSuffix(strings.ToLower(suffix), ".") // normalize
	base, err := publicsuffix.EffectiveTLDPlusOne(suffix)
	if err != nil {
		return "", err
	}

	if suffix == base {
		return nodeName, nil // no subdomain
	}

	subdomain := strings.TrimSuffix(suffix, "."+base)
	return nodeName + "." + subdomain, nil
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

	if err != nil {
		log.Fatalf("failed to list zones: %v", err)
		return err
	}

	// Set the zone in the provider
	cdp.zone = &zone.Result[0]
	if cdp.zone == nil {
		log.Fatalf("Something unpredictable happened. Zone not found for suffix: %s", suffix)
	}
	return nil
}
