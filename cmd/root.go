package cmd

import (
    "fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"k8s-audit/controller"
    "k8s-audit/utils"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "mycoolname",
	Short: "myamazingshortmessage",
	Long:  `mysuberplongmessage`,
	Run: func(cmd *cobra.Command, args []string) {
		master, err := cmd.Flags().GetString("master")
		cfg := newControllerConfig(master, "")
		c := controller.New(cfg)
        c.Init()
        err = c.Run()
		if err != nil {
            logrus.Fatalf("damn it...: %s", err)
        }
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().StringP("master", "", "", "Apiserver address")
}

func newControllerConfig(masterHost, ns string) controller.Config {
    logrus.Infof("newControllerConfig")
	f := utils.GetFactory()
	kubecli, err := f.Client()
	if err != nil {
		fmt.Errorf("Can not get kubernetes config: %s", err)
	}
	if ns == "" {
		ns, _, err = f.DefaultNamespace()
		if err != nil {
			fmt.Errorf("Can not get kubernetes config: %s", err)
		}
	}
	if masterHost == "" {
		k8sConfig, err := f.ClientConfig()
		if err != nil {
			fmt.Errorf("Can not get kubernetes config: %s", err)
		}
		if k8sConfig == nil {
			fmt.Errorf("Got nil k8sConfig, please check if k8s cluster is available.")
		} else {
			masterHost = k8sConfig.Host
		}
	}
	cfg := controller.Config{
		Namespace:  ns,
		KubeCli:    kubecli,
		MasterHost: masterHost,
	}

	return cfg
}
