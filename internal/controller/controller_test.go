package controller

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// MockProvider mocks the Provider interface for testing
type MockProvider struct {
	Called     bool
	CalledIP   string
	CalledNode string
}

func (m *MockProvider) UpdatePTR(ip, nodeName string) error {
	m.Called = true
	m.CalledIP = ip
	m.CalledNode = nodeName
	return nil
}

func (m *MockProvider) UpdateA(ip, nodeName string) error {
	m.Called = true
	m.CalledIP = ip
	m.CalledNode = nodeName
	return nil
}

func TestControllerIntegration(t *testing.T) {
	testEnv := &envtest.Environment{}
	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatalf("failed to start envtest: %v", err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	mockHostingProvider := &MockProvider{}
	mockDnsProvider := &MockProvider{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	controller := WithCustomKubernetesClient(mockHostingProvider, mockDnsProvider, client)

	go func() {
		controller.StartController()
	}()

	// Create a fake Node
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-1",
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{
					Type:    v1.NodeExternalIP,
					Address: "1.2.3.4",
				},
			},
		},
	}

	_, err = client.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create node: %v", err)
	}

	waitForController()

	// Validate Hosting Provider
	if !mockHostingProvider.Called {
		t.Fatal("expected UpdatePTR to be called, but it wasn't")
	}
	if mockHostingProvider.CalledIP != "1.2.3.4" {
		t.Fatalf("expected IP 1.2.3.4 but got %s", mockHostingProvider.CalledIP)
	}
	if mockHostingProvider.CalledNode != "node-1" {
		t.Fatalf("expected node name 'node-1' but got %s", mockHostingProvider.CalledNode)
	}

	t.Logf("✅ Integration test passed: PTR updated for %s -> %s", mockHostingProvider.CalledIP, mockHostingProvider.CalledNode)

	// Validate DNS Provider
	if !mockDnsProvider.Called {
		t.Fatal("expected UpdatePTR to be called, but it wasn't")
	}
	if mockDnsProvider.CalledIP != "1.2.3.4" {
		t.Fatalf("expected IP 1.2.3.4 but got %s", mockHostingProvider.CalledIP)
	}
	if mockDnsProvider.CalledNode != "node-1" {
		t.Fatalf("expected node name 'node-1' but got %s", mockHostingProvider.CalledNode)
	}

	t.Logf("✅ Integration test passed: A updated for %s -> %s", mockDnsProvider.CalledIP, mockDnsProvider.CalledNode)
}

func waitForController() {
	time.Sleep(2 * time.Second)
}
