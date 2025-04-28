package providers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/vultr/govultr/v3"
	"golang.org/x/oauth2"
)

type VultrProvider struct {
	client      *govultr.Client
	ptrTemplate string
}

func NewVultrProvider(suffix string) *VultrProvider {
	apiKey := os.Getenv("VULTR_API_KEY")

	config := &oauth2.Config{}
	ctx := context.Background()
	ts := config.TokenSource(ctx, &oauth2.Token{AccessToken: apiKey})
	client := govultr.NewClient(oauth2.NewClient(ctx, ts))

	ptrTemplate := "%s." + suffix

	return &VultrProvider{
		client:      client,
		ptrTemplate: ptrTemplate,
	}
}

func (vp *VultrProvider) UpdatePTR(ip, nodeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	instanceID, err := vp.getInstanceIDByIP(ctx, ip)
	if err != nil {
		return fmt.Errorf("get instance id failed: %w", err)
	}

	ptr := fmt.Sprintf(vp.ptrTemplate, nodeName)

	reverse := &govultr.ReverseIP{
		IP:      ip,
		Reverse: ptr,
	}

	err = vp.client.Instance.CreateReverseIPv4(ctx, instanceID, reverse)
	if err != nil {
		return fmt.Errorf("failed to update PTR: %w", err)
	}

	fmt.Printf("✅ PTR updated: %s -> %s\n", ip, ptr)
	return nil
}

func (vp *VultrProvider) getInstanceIDByIP(ctx context.Context, ip string) (string, error) {
	instances, meta, _, err := vp.client.Instance.List(ctx, &govultr.ListOptions{PerPage: 100})
	if err != nil {
		return "", fmt.Errorf("failed to list instances: %w", err)
	}

	for _, inst := range instances {
		if inst.MainIP == ip {
			return inst.ID, nil
		}
	}

	// If there are multiple pages of instances, loop through them (optional)
	for meta.Links.Next != "" {
		nextInstances, nextMeta, _, err := vp.client.Instance.List(ctx, &govultr.ListOptions{
			Cursor:  meta.Links.Next,
			PerPage: 100,
		})
		if err != nil {
			return "", fmt.Errorf("failed to list instances: %w", err)
		}
		for _, inst := range nextInstances {
			if inst.MainIP == ip {
				return inst.ID, nil
			}
		}
		meta = nextMeta
	}

	return "", fmt.Errorf("instance not found for IP %s", ip)
}
