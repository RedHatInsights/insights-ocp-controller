package main

import (
	"log"
	"os"
	"time"

	"github.com/RedHatInsights/insights-ocp-controller/pkg/controller"
	// _ "github.com/openshift/origin/pkg/api/install"
	// osclient "github.com/openshift/origin/pkg/client"
	// "github.com/openshift/origin/pkg/cmd/util/clientcmd"

	// "github.com/spf13/pflag"
	restclient "k8s.io/client-go/rest"
	kclient "k8s.io/client-go/kubernetes"
)

func main() {

	config, err := restclient.InClusterConfig()
	if err != nil {
		log.Printf("Error getting in cluster config. AAAAAHHHHHHH!!!! Error message: %s", err)

		// config, err = clientcmd.DefaultClientConfig(pflag.NewFlagSet("empty", pflag.ContinueOnError)).ClientConfig()
		// if err != nil {
		// 	log.Printf("Error creating default client config: %s", err)
		// 	os.Exit(1)
		// }
	}

	kubeClient, err := kclient.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating cluster config: %s", err)
		os.Exit(1)
	}
	// openshiftClient, err := osclient.New(config)
	// if err != nil {
	// 	log.Printf("Error creating OpenShift client: %s", err)
	// 	os.Exit(2)
	// }

	c := controller.NewController(kubeClient)
	for true {
		c.ScanImages()
		time.Sleep(time.Hour)
	}

}
