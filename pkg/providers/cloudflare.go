package providers

import (
	"os"
	"fmt"
	"context"

	"github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/option"
	"github.com/cloudflare/cloudflare-go/v5/zones"
)

type CloudflareDnsProvider struct {
	client *cloudflare.Client
}

func NewCloudflareDnsProvider(zoneID string) *CloudflareDnsProvider {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	client := cloudflare.NewClient(
		option.WithAPIToken(apiToken), // defaults to os.LookupEnv("CLOUDFLARE_API_TOKEN")
	)

	// zone, err := client.Zones.New(context.TODO(), zones.ZoneNewParams{
	// 	Account: cloudflare.F(zones.ZoneNewParamsAccount{
	// 		ID: cloudflare.F("023e105f4ecef8ad9ca31a8372d0c353"),
	// 	}),
	// 	Name: cloudflare.F("example.com"),
	// 	Type: cloudflare.F(zones.TypeFull),
	// })
	// if err != nil {
	// 	panic(err.Error())
	// }
	// fmt.Printf("%+v\n", zone.ID)

	return &CloudflareDnsProvider{
		client: client,
	}
}

func (cdp *CloudflareDnsProvider) UpdateA(ip, nodeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	zone, err := cdp.client.Zones.Get(ctx, zones.ZoneGetParams{})
	if err != nil {
		return fmt.Errorf("failed to get zone: %w", err)
	}
	record, err := cdp.client.DNS.Records.New(ctx, zone.ID, 
		cdp.client.DNS.Options[zones.RecordNewParams{
		Type:    "A",
		Name:    nodeName,
		Content: ip,
		TTL:     60]
	})
	return nil
}