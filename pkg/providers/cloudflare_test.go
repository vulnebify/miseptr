package providers

import (
	"testing"
)

func TestCloudflareIntegration(t *testing.T) {
	provider := NewCloudflareDnsProvider("scanning.vulnefy.com")

	err := provider.UpdateA("192.168.0.2", "vulnefy-node123")

	if err != nil {
		t.Fatalf("failed to assign A record: %v", err)
	}
}
