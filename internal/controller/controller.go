package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

	var externalIP string
	for i := 0; i < 6; i++ {
		updatedNode, err := c.clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Failed to fetch updated node %s: %v\n", nodeName, err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, addr := range updatedNode.Status.Addresses {
			if addr.Type == v1.NodeExternalIP {
				externalIP = addr.Address
				break
			}
		}

		if externalIP != "" {
			break
		}

		fmt.Printf("ExternalIP still not assigned for node %s, retrying...\n", nodeName)
		time.Sleep(5 * time.Second)
	}

	if externalIP == "" {
		fmt.Printf("Skipping PTR update for %s: no ExternalIP after timeout.\n", nodeName)
		return
	}

	if err := c.provider.UpdatePTR(externalIP, nodeName); err != nil {
		fmt.Printf("Failed to update PTR for %s (%s): %v\n", nodeName, externalIP, err)
	}
}
