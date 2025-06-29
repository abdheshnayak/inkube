package sw

import (
	"os"

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
	Short: "switch the app, you are working on",
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(); err != nil {
			fn.PrintError(err)
		}
	},
}

func run() error {
	s, ok := os.LookupEnv("INKUBE")
	if ok && s == "true" {
		return fn.Error("you are already in inkube session, please exit the session first")
	}

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

	cfg.Connect = true
	cfg.Namespace = ns.Name
	cfg.Bridge.Intercept = false
	cfg.Bridge.Name = dep.Name

	cfg.LoadEnv.Container = cont.Name
	cfg.LoadEnv.Enabled = true
	return cfg.Write()
}
