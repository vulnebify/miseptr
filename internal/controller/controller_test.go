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

	mockProvider := &MockProvider{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	controller := WithCustomKubernetesClient(mockProvider, client)

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

	// Validate
	if !mockProvider.Called {
		t.Fatal("expected UpdatePTR to be called, but it wasn't")
	}
	if mockProvider.CalledIP != "1.2.3.4" {
		t.Fatalf("expected IP 1.2.3.4 but got %s", mockProvider.CalledIP)
	}
	if mockProvider.CalledNode != "node-1" {
		t.Fatalf("expected node name 'node-1' but got %s", mockProvider.CalledNode)
	}

	t.Logf("✅ Integration test passed: PTR updated for %s -> %s", mockProvider.CalledIP, mockProvider.CalledNode)
}

func waitForController() {
	time.Sleep(2 * time.Second)
}
