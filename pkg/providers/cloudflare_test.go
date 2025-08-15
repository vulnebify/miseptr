package providers

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/dns"
	"github.com/cloudflare/cloudflare-go/v5/option"
)

func TestCloudflareIntegration(t *testing.T) {
	// AAA testing
	provider := NewCloudflareDnsProvider("scanning.vulnefy.com")
	// Generate a unique node name for testing with random suffix

	// Check if the A record was created successfully
	// Go and get the dns record and verify it matches the expected values
	ctx := context.Background()

	recordType := dns.RecordListParamsTypeA
	recordListParamsMatch := dns.RecordListParamsMatchAny
	recordListParamsName := dns.RecordListParamsName{
		Exact: cloudflare.F("vulnefy-node123.scanning.vulnefy.com"),
	}

	dnsRecordListParams := dns.RecordListParams{
		ZoneID: cloudflare.F(provider.zone.ID),
		Type:   cloudflare.F(recordType),
		Match:  cloudflare.F(recordListParamsMatch),
		Name:   cloudflare.F(recordListParamsName),
	}

	err := provider.UpdateA("192.168.0.2", "vulnefy-node123")

	if err != nil {
		t.Fatalf("failed to assign A record: %v", err)
	}

	createdRecord, err := provider.client.DNS.Records.List(
		ctx, dnsRecordListParams, option.WithMaxRetries(3),
	)
	if err != nil {
		t.Fatalf("failed to list DNS records: %v", err)
	}
	if len(createdRecord.Result) != 1 {
		t.Fatalf("expected 1 DNS record, got %d", len(createdRecord.Result))
	}
	if createdRecord.Result[0].Name != "vulnefy-node123.scanning.vulnefy.com" {
		t.Fatalf("expected DNS 'vulnefy-node123.scanning.vulnefy.com', got '%s'", createdRecord.Result[0].Name)
	}

	t.Cleanup(func() {
		// Cleanup: delete all DNS records
		ctx := context.Background()
		recordsToDelete, err := provider.client.DNS.Records.List(
			ctx, dnsRecordListParams, option.WithMaxRetries(3),
		)
		if err != nil {
			t.Fatalf("failed to list DNS records for cleanup: %v", err)
		}
		for _, record := range recordsToDelete.Result {
			// Print the record being deleted
			t.Logf("Deleting DNS record: %s", record.Name)
			// Delete the record
			provider.client.DNS.Records.Delete(ctx, record.ID, dns.RecordDeleteParams{
				ZoneID: cloudflare.F(provider.zone.ID),
			}, option.WithMaxRetries(3))

			if err != nil {
				t.Fatalf("failed to delete DNS record: %v", err)
			}
		}
	})

}
