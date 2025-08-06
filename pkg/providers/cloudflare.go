package providers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/dns"
	"github.com/cloudflare/cloudflare-go/v5/option"
	"github.com/cloudflare/cloudflare-go/v5/zones"
)

type CloudflareDnsProvider struct {
	client *cloudflare.Client
	zone   *zones.Zone
}

func NewCloudflareDnsProvider(zoneName string) *CloudflareDnsProvider {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		os.Exit(1)
	}

	client := cloudflare.NewClient(
		option.WithAPIToken(apiToken),
	)

	return &CloudflareDnsProvider{
		client: client,
		zone:   nil,
	}
}

func (cdp *CloudflareDnsProvider) UpdateA(ip, nodeName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// requestOptions = cloudflare.
	recordParams := dns.ARecordParam{
		Name:    cloudflare.F(nodeName),
		TTL:     cloudflare.F(dns.TTL(60)),
		Type:    cloudflare.F(dns.ARecordTypeA),
		Content: cloudflare.F(ip),
		Proxied: cloudflare.F(false),
	}

	newParams := dns.RecordNewParams{
		ZoneID: cloudflare.F(cdp.zone.ID),
		Body:   recordParams,
	}

	options := option.WithMaxRetries(3)

	_, err := cdp.client.DNS.Records.New(ctx, newParams, options)
	if err != nil {
		fmt.Printf("Failed to create A record: %v\n", err)
	}
}
