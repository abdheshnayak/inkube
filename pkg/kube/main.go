package kube

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/abdheshnayak/inkube/pkg/fn"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	client        *Client
	singletonOnce sync.Once
)

func Singleton() *Client {
	singletonOnce.Do(func() {
		var err error
		client, err = NewClient()
		if err != nil {
			fn.PrintError(err)
			os.Exit(1)
		}
	})
	return client
}

type Client struct {
	*kubernetes.Clientset
}

func (c *Client) Ctx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return ctx
}

func NewClient() (*Client, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

func getClient() (*kubernetes.Clientset, error) {
	config, err := getRestConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func getRestConfig() (*rest.Config, error) {
	// 1. Use in-cluster config if running inside a pod
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// 2. Otherwise, use kubeconfig from KUBECONFIG or default path
	var kubeconfigPath string
	if env := os.Getenv("KUBECONFIG"); env != "" {
		kubeconfigPath = env
	} else {
		if home := homedir.HomeDir(); home != "" {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}
