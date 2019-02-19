package main

import (
	"log"
	"os"
	// "time"

	// "github.com/RedHatInsights/insights-ocp-controller/pkg/controller"
	// _ "github.com/openshift/origin/pkg/api/install"
	// osclient "github.com/openshift/origin/pkg/client"
	// "github.com/openshift/origin/pkg/cmd/util/clientcmd"

	// "github.com/spf13/pflag"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating cluster config: %s", err)
		os.Exit(1)
	}
	// openshiftClient, err := osclient.New(config)
	// if err != nil {
	// 	log.Printf("Error creating OpenShift client: %s", err)
	// 	os.Exit(2)
	// }
	imageList, err := kubeClient.CoreV1().Pods("").List(metav1.ListOptions{})

	if err != nil {
		log.Println(err)
		return
	}

	if imageList == nil {
		log.Println("No images")
		return
	}

	log.Println(imageList.Items)

	// c := controller.NewController(kubeClient)
	// for true {
	// 	c.ScanImages()
	// 	time.Sleep(time.Hour)
	// }

}
