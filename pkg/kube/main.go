package kube

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/abdheshnayak/inkube/flags"
	"github.com/abdheshnayak/inkube/pkg/egob"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/ui/spinner"
	"github.com/abdheshnayak/inkube/pkg/ui/text"
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
	cancel context.CancelFunc
}

func (c *Client) Ctx() context.Context {
	ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
	c.cancel = cf
	return ctx
}

func NewClient() (*Client, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}
	return &Client{
		Clientset: client,
		cancel:    nil,
	}, nil
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

func (c *Client) GetEnvs(namespace, name, contname string, refetch bool) (map[string]string, error) {
	defer spinner.Client.UpdateMessage("Getting environment variables")()
	cacheDir := flags.GetCacheDir()
	fileNamePath := path.Join(cacheDir, fmt.Sprintf("%s-%s-%s.secret.cache", namespace, name, contname))

	if !refetch {
		if evs, err := func() (map[string]string, error) {
			b, err := os.ReadFile(fileNamePath)
			if err != nil {
				return nil, err
			}
			resp := make(map[string]string)
			if err := egob.Unmarshal(b, &resp); err != nil {
				return nil, err
			}
			return resp, nil
		}(); err == nil {
			fn.Log(text.Blue("[#] using cached env vars"))
			return evs, nil
		}
	}

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
			maps.Copy(envs, cm.Data)
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

	b, err := egob.Marshal(envs)
	if err != nil {
		fn.Log(text.Yellow("[!] failed to marshal env vars"))
		return envs, err
	}

	if err := os.WriteFile(fileNamePath, b, 0o644); err != nil {
		fn.Log(text.Yellow("[!] failed to env vars to cache"))
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
