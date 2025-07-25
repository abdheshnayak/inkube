package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/abdheshnayak/inkube/flags"
	"github.com/abdheshnayak/inkube/pkg/egob"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/kube"
	"github.com/abdheshnayak/inkube/pkg/ui/spinner"
	"github.com/abdheshnayak/inkube/pkg/ui/text"
)

type KubeVpnClient struct {
	managerNamespace string
}

func (c *KubeVpnClient) Quit() error {
	return fn.ExecCmd("kubevpn quit", nil, true)
}

func (c *KubeVpnClient) EnsureDependencies() error {
	_, err := exec.LookPath("kubevpn")
	if err != nil {
		return fmt.Errorf("kubevpn not found, please ensure kubevpn is installed")
	}
	return nil
}

func (c *KubeVpnClient) Status() (connected, intercepted bool, err error) {
	var b []byte
	for {
		s := flags.GetCacheDir()
		cachePath := fmt.Sprintf("%s/kubevpn.json", s)
		type KubeVpnStatus struct {
			Data []byte
			Time time.Time
		}

		b, err = os.ReadFile(cachePath)
		if err == nil {
			var status KubeVpnStatus
			if err := egob.Unmarshal(b, &status); err == nil {
				if time.Since(status.Time) < time.Second*10 {
					b = status.Data
					break
				}
			}
		}

		b, err = fn.Exec("kubevpn status -ojson", nil)
		if err != nil {
			return false, false, err
		}

		status := KubeVpnStatus{Data: b, Time: time.Now()}
		if b2, err := egob.Marshal(status); err == nil {
			os.WriteFile(cachePath, b2, 0644)
		}

		break
	}

	var status []struct {
		ClusterId string `json:"ClusterID"`  // k3d-mycluster
		Cluster   string `json:"Cluster"`    // k3d-mycluster
		Mode      string `json:"Mode"`       // full
		KubeConfg string `json:"KubeConfig"` // /Users/abdhesh/.kube/k3d.yaml
		Namespace string `json:"Namespace"`  // test
		Status    string `json:"Status"`     // connected
		Netif     string `json:"Netif"`      // utun4
		ProxyList []struct {
			ClusterID  string `json:"ClusterID"`  // k3d-mycluster
			Cluster    string `json:"Cluster"`    // k3d-mycluster
			Kubeconfig string `json:"Kubeconfig"` // /Users/abdhesh/.kube/k3d.yaml
			Namespace  string `json:"Namespace"`  // test
			Workload   string `json:"Workload"`   // deployments.apps/nginx-deployment
			RuleList   []struct {
				LocalTunIPv4  string `json:"LocalTunIPv4"` // 198.19.0.101
				LocalTunIPv6  string `json:"LocalTunIPv6"` // 2001:2::999a
				CurrentDevice bool   `json:"CurrentDevice"`
				PortMap       struct {
					Port int `json:"80"`
				} `json:"PortMap"`
			} `json:"RuleList"`
		} `json:"ProxyList"`
	}

	if bytes.TrimSpace(b) != nil {
		if err := json.Unmarshal(b, &status); err != nil {
			return false, false, err
		}
	}

	cName, err := kube.Singleton().GetClusterName()
	if err != nil {
		return false, false, err
	}

	connected = false
	intercepted = false

	for _, s := range status {
		if s.Status == "connected" && cName == s.Cluster {
			connected = true
			for _, v := range s.ProxyList {
				if v.Cluster == s.Cluster {
					intercepted = true
					break
				}
			}
			break
		}

	}

	return connected, intercepted, nil
}

func (c *KubeVpnClient) IsConnected() (*string, int, error) {
	b, err := fn.Exec("kubevpn status -ojson", nil)
	if err != nil {
		return nil, 0, err
	}
	var status []struct {
		ClusterId string `json:"ClusterID"`
		Cluster   string `json:"Cluster"`
		Mode      string `json:"Mode"`
		KubeConfg string `json:"KubeConfig"`
		Namespace string `json:"Namespace"`
		Status    string `json:"Status"`
		Netif     string `json:"Netif"`
	}

	if bytes.TrimSpace(b) != nil {
		if err := json.Unmarshal(b, &status); err != nil {
			return nil, 0, err
		}
	}

	if len(status) == 0 {
		return nil, 0, fn.Errorf("no active sessions found")
	}

	currCluster, err := kube.Singleton().GetClusterName()
	if err != nil {
		return nil, 0, err
	}

	for i, s := range status {
		if s.Cluster == currCluster {
			return &s.Cluster, i, nil
		}
	}

	return nil, 0, fn.Errorf("no active sessions found")
}

func (c *KubeVpnClient) Connect(ns string) error {
	// defer spinner.Client.UpdateMessage("connecting to cluster")()
	fn.Log(text.Blue("[#] connecting to cluster"))
	// ensure namespace exists
	if err := kube.Singleton().EnsureNamespace(c.managerNamespace); err != nil {
		return err
	}

	return fn.ExecCmd(fmt.Sprintf("kubevpn connect --manager-namespace=%s", c.managerNamespace), nil, false)
}

func (c *KubeVpnClient) Disconnect() error {
	defer spinner.Client.UpdateMessage("disconnecting from cluster")()
	_, i, err := c.IsConnected()
	if err != nil {
		return err
	}

	return fn.ExecCmd(fmt.Sprintf("kubevpn disconnect %d", i), nil, false)
}

func (c *KubeVpnClient) Intercept(name string, ns string) error {
	defer spinner.Client.UpdateMessage("intercepting pod")()
	return fn.ExecCmd(fmt.Sprintf("kubevpn proxy deployment/%s -n %s --manager-namespace=%s", name, ns, c.managerNamespace), nil, false)
}

func (c *KubeVpnClient) Leave(name string, ns string) error {
	defer spinner.Client.UpdateMessage(fmt.Sprintf("leaving intercept for %s", name))()

	return fn.ExecCmd(fmt.Sprintf("kubevpn leave deployment/%s -n %s", name, ns), nil, false)
}

func NewKubeVpn() ConnectClient {
	return &KubeVpnClient{
		managerNamespace: "kubevpn",
	}
}
