package init

import (
	"os"
	"path"

	"github.com/abdheshnayak/inkube/pkg/config"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/kube"
	"github.com/abdheshnayak/inkube/pkg/ui/fzf"
	"github.com/abdheshnayak/inkube/pkg/ui/spinner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "init",
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(); err != nil {
			fn.PrintError(err)
		}
	},
}

func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	deboxPath := path.Join(cwd, "devbox.json")
	if _, err := os.Stat(deboxPath); err != nil {
		if os.IsNotExist(err) {
			return fn.Error("devbox.json not found, please run `devbox init` first")
		}
		return err
	}

	kubeClient := kube.Singleton()

	if _, err := config.NewConfig(); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

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

		b, err := yaml.Marshal(config.Config{
			Namespace: ns.Name,
			Name:      dep.Name,
			Container: cont.Name,
			Version:   "v1",
			Intercept: true,

			// not supported yet
			Devbox:  false,
			Loadenv: false,
		})
		if err != nil {
			return err
		}

		cfgPath := path.Join(cwd, "inkube.yaml")
		if err := os.WriteFile(cfgPath, b, 0o644); err != nil {
			return err
		}
	}

	return nil
}
