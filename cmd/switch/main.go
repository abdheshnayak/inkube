package sw

import (
	"github.com/abdheshnayak/inkube/pkg/config"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/kube"
	"github.com/abdheshnayak/inkube/pkg/ui/fzf"
	"github.com/abdheshnayak/inkube/pkg/ui/spinner"
	"github.com/spf13/cobra"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Cmd = &cobra.Command{
	Use:   "switch",
	Short: "switch",
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(); err != nil {
			fn.PrintError(err)
		}
	},
}

func run() error {
	kubeClient := kube.Singleton()

	cfg := config.Singleton()

	f := spinner.Client.UpdateMessage("fetching namespaces")
	nl, err := kubeClient.Clientset.CoreV1().Namespaces().List(kubeClient.Ctx(), v1.ListOptions{})
	f()
	if err != nil {
		return err
	}

	if len(nl.Items) == 0 {
		return fn.Error("no namespaces found, please create a namespace")
	}

	ns, err := fzf.FindOne(nl.Items, func(item corev1.Namespace) string {
		return item.Name
	}, fzf.WithPrompt("select namespace"))
	if err != nil {
		return err
	}

	dl, err := kubeClient.Clientset.AppsV1().Deployments(ns.Name).List(kubeClient.Ctx(), v1.ListOptions{})
	if err != nil {
		return err
	}

	dep, err := fzf.FindOne(dl.Items, func(item appv1.Deployment) string {
		return item.Name
	}, fzf.WithPrompt("select deployment"))

	if err != nil {
		return err
	}

	cont, err := fzf.FindOne(dep.Spec.Template.Spec.Containers, func(item corev1.Container) string {
		return item.Name
	}, fzf.WithPrompt("select container"))
	if err != nil {
		return err
	}

	cfg.Namespace = ns.Name
	cfg.Tele.Intercept = false
	cfg.Tele.Connect = true
	cfg.Tele.Name = dep.Name

	cfg.LoadEnv.Container = cont.Name
	cfg.LoadEnv.Enabled = true
	return cfg.Write()
}
