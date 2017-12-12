package main

import (
	"log"
	"os"

	"github.com/RedHatInsights/insights-ocp/controller/pkg/controller"
	_ "github.com/openshift/origin/pkg/api/install"
	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/client/restclient"

	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

func main() {

	config, err := restclient.InClusterConfig()
	if err != nil {
		log.Printf("Error getting in cluster config. Fallback to native config. Error message: %s", err)

		config, err = clientcmd.DefaultClientConfig(pflag.NewFlagSet("empty", pflag.ContinueOnError)).ClientConfig()
		if err != nil {
			log.Printf("Error creating default client config: %s", err)
			os.Exit(1)
		}
	}

	kubeClient, err := kclient.New(config)
	if err != nil {
		log.Printf("Error creating cluster config: %s", err)
		os.Exit(1)
	}
	openshiftClient, err := osclient.New(config)
	if err != nil {
		log.Printf("Error creating OpenShift client: %s", err)
		os.Exit(2)
	}

	c := controller.NewController(openshiftClient, kubeClient)

	c.ScanImages()

}
