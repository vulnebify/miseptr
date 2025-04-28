package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vulnebify/miseptr/pkg/providers"
)

type Controller struct {
	provider  providers.Provider
	clientset kubernetes.Interface
}

func NewController(provider providers.Provider) (*Controller, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// fallback to local kubeconfig
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &Controller{
		provider:  provider,
		clientset: clientset,
	}, nil
}

func WithCustomKubernetesClient(provider providers.Provider, clientset kubernetes.Interface) *Controller {
	return &Controller{
		provider:  provider,
		clientset: clientset,
	}
}

func (c *Controller) StartController() {
	watcher, err := c.clientset.CoreV1().Nodes().Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(fmt.Sprintf("Failed to watch nodes: %v", err))
	}
	fmt.Println("👀 Watching for new nodes...")

	for event := range watcher.ResultChan() {
		if event.Type == watch.Added {
			node, ok := event.Object.(*v1.Node)
			if !ok {
				fmt.Println("Received unexpected object type")
				continue
			}
			c.handleNewNode(node)
		}
	}
}

func (c *Controller) handleNewNode(node *v1.Node) {
	nodeName := node.GetName()
	externalIP := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeExternalIP {
			externalIP = addr.Address
			break
		}
	}
	if externalIP == "" {
		fmt.Printf("No ExternalIP found for node %s\n", nodeName)
		return
	}

	err := c.provider.UpdatePTR(externalIP, nodeName)
	if err != nil {
		fmt.Printf("Failed to update PTR for node %s: %v\n", nodeName, err)
	}
}
