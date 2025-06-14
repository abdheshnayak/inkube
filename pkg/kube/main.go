package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/ui/spinner"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (c *Client) GetEnvs(namespace, name, contname string) (map[string]string, error) {
	defer spinner.Client.UpdateMessage("Getting environment variables")()

	envs := make(map[string]string)

	// Get the Deployment
	deploy, err := c.AppsV1().Deployments(namespace).Get(c.Ctx(), name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var container *corev1.Container
	for _, cont := range deploy.Spec.Template.Spec.Containers {
		if cont.Name == contname {
			container = &cont
			break
		}
	}
	if container == nil {
		return nil, fmt.Errorf("container %s not found", contname)
	}

	// Handle Env
	for _, env := range container.Env {
		if env.Value != "" {
			envs[env.Name] = env.Value
		} else if env.ValueFrom != nil {
			switch {
			case env.ValueFrom.ConfigMapKeyRef != nil:
				cm, err := c.CoreV1().ConfigMaps(namespace).Get(c.Ctx(), env.ValueFrom.ConfigMapKeyRef.Name, v1.GetOptions{})
				if err != nil {
					return nil, err
				}
				val := cm.Data[env.ValueFrom.ConfigMapKeyRef.Key]
				envs[env.Name] = val

			case env.ValueFrom.SecretKeyRef != nil:
				secret, err := c.CoreV1().Secrets(namespace).Get(c.Ctx(), env.ValueFrom.SecretKeyRef.Name, v1.GetOptions{})
				if err != nil {
					return nil, err
				}
				val := string(secret.Data[env.ValueFrom.SecretKeyRef.Key])
				envs[env.Name] = val

			case env.ValueFrom.FieldRef != nil:
				envs[env.Name] = fmt.Sprintf("fieldRef: %s", env.ValueFrom.FieldRef.FieldPath)

			case env.ValueFrom.ResourceFieldRef != nil:
				envs[env.Name] = fmt.Sprintf("resourceFieldRef: %s", env.ValueFrom.ResourceFieldRef.Resource)
			}
		}
	}

	// Handle EnvFrom
	for _, envFrom := range container.EnvFrom {
		if envFrom.ConfigMapRef != nil {
			cm, err := c.CoreV1().ConfigMaps(namespace).Get(c.Ctx(), envFrom.ConfigMapRef.Name, v1.GetOptions{})
			if err != nil {
				return nil, err
			}
			for k, v := range cm.Data {
				envs[k] = v
			}
		}
		if envFrom.SecretRef != nil {
			secret, err := c.CoreV1().Secrets(namespace).Get(c.Ctx(), envFrom.SecretRef.Name, v1.GetOptions{})
			if err != nil {
				return nil, err
			}
			for k, v := range secret.Data {
				envs[k] = string(v)
			}
		}
	}

	return envs, nil
}

// func (c *Client) GetEnvs(namespace string, name string, contname string) (map[string]string, error) {
// 	envs := make(map[string]string)
// 	pod, err := c.AppsV1().Deployments(namespace).Get(c.Ctx(), name, v1.GetOptions{})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	for _, cont := range pod.Spec.Template.Spec.Containers {
// 		if cont.Name == contname {
// 			for _, env := range cont.Env {
// 				envs[env.Name] = env.Value
// 			}
// 			break
// 		}
// 	}
//
// 	return envs, nil
// }
