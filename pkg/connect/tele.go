package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/kube"
	"github.com/abdheshnayak/inkube/pkg/ui/spinner"
)

func (c *TeleClient) EnsureDependencies() error {
	_, err := exec.LookPath("telepresence")
	if err != nil {
		return fmt.Errorf("telepresence not found, please ensure telepresence is installed")
	}
	return nil
}

func (c *TeleClient) Status() (connected, intercept bool, err error) {
	b, err := fn.Exec("telepresence status --output json", nil)
	if err != nil {
		return false, false, err
	}

	var status TeleStatus

	if bytes.TrimSpace(b) != nil {
		if err := json.Unmarshal(b, &status); err != nil {
			return false, false, err
		}
	}

	if status.UserDaemon.Status == "Connected" {
		return true, status.UserDaemon.Name != "", nil
	}

	return false, false, nil
}

type TeleStatus struct {
	TrafficManager struct {
		Name          string `json:"name"`
		TrafficAgent  string `json:"traffic_agent"`
		Version       string
		VirtualSubnet string `json:"virtual_subnet"`
	} `json:"traffic_manager"`
	UserDaemon struct {
		DaemonPort        int    `json:"daemon_port"`
		Executable        string `json:"executable"`
		InDocker          bool   `json:"in_docker"`
		InstallID         string `json:"install_id"`
		KubernetesContext string `json:"kubernetes_context"`
		KubernetesServer  string `json:"kubernetes_server"`
		ManagerNamespace  string `json:"manager_namespace"`
		MappedNamespaces  []string
		Name              string `json:"name"`
		Namespace         string `json:"namespace"`
		Running           bool   `json:"running"`
		Status            string `json:"status"`
		Version           string
	} `json:"user_daemon"`
}

type TeleClient struct {
	managerNamespace string
}

func (c *TeleClient) IsConnected() (*string, int, error) {
	b, err := fn.Exec("telepresence status --output json", nil)
	if err != nil {
		return nil, 0, err
	}

	var status TeleStatus

	if bytes.TrimSpace(b) != nil {
		if err := json.Unmarshal(b, &status); err != nil {
			return nil, 0, err
		}
	}

	if status.UserDaemon.Status == "Connected" {
		return &status.UserDaemon.Name, 0, nil
	}

	return nil, 0, fn.Errorf("no active sessions found")
}

func (c *TeleClient) Connect(ns string) error {
	defer spinner.Client.UpdateMessage("connecting to cluster")()
	// ensure namespace exists
	if err := kube.Singleton().EnsureNamespace(c.managerNamespace); err != nil {
		return err
	}

	return fn.ExecCmd(fmt.Sprintf("telepresence connect -n %s", ns), nil, false)
}

func (c *TeleClient) Disconnect() error {
	defer spinner.Client.UpdateMessage("disconnecting from cluster")()

	return fn.ExecCmd(fmt.Sprintf("telepresence quit"), nil, false)
}

func (c *TeleClient) Intercept(name string, ns string) error {

	defer spinner.Client.UpdateMessage("intercepting pod")()
	return fn.ExecCmd(fmt.Sprintf("telepresence intercept %s/%s", ns, name), nil, true)
}

func (c *TeleClient) Leave(name string, ns string) error {
	defer spinner.Client.UpdateMessage(fmt.Sprintf("leaving intercept for %s", name))()

	return fn.ExecCmd(fmt.Sprintf("telepresence leave %s -n %s", name, ns), nil, true)
}

func NewTele() ConnectClient {
	return &TeleClient{
		managerNamespace: "default",
	}
}
