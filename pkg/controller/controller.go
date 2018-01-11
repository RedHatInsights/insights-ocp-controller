package controller

import (
	"os"
	"log"
	"sync"

	"github.com/fsouza/go-dockerclient"
	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/pflag"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

type Controller struct {
	openshiftClient *osclient.Client
	kubeClient      *kclient.Client
	mapper          meta.RESTMapper
	typer           runtime.ObjectTyper
	f               *clientcmd.Factory
	wait            sync.WaitGroup
}

func getScanArgs(imageID string) []string {

	args := []string{}
	args = append(args, "/insights-scanner")
	args = append(args, "-image")
	args = append(args, imageID)

	return args
}

func NewController(os *osclient.Client, kc *kclient.Client) *Controller {

	f := clientcmd.New(pflag.NewFlagSet("empty", pflag.ContinueOnError))
	mapper, typer := f.Object(false)

	return &Controller{
		openshiftClient: os,
		kubeClient:      kc,
		mapper:          mapper,
		typer:           typer,
		f:               f,
	}
}

func (c *Controller) ScanImages() {

	imageList, err := c.openshiftClient.Images().List(kapi.ListOptions{})

	if err != nil {
		log.Println(err)
		return
	}

	if imageList == nil {
		log.Println("No images")
		return
	}
	for _, image := range imageList.Items {
		log.Printf("Scanning image %s %s", image.DockerImageMetadata.ID, image.DockerImageReference)

	}
	c.scanImage("image.DockerImageMetadata.ID", getScanArgs("registry.access.redhat.com/rhscl/postgresql-94-rhel7"))
	return

}

func (c *Controller) scanImage(id string, args []string) error {
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewVersionedClient(endpoint, "1.22")
	binds := []string{}
	binds = append(binds, "/var/run/docker.sock:/var/run/docker.sock")
	scanner := "registry.access.redhat.com/insights-scanner"

	container, err := client.CreateContainer(
		docker.CreateContainerOptions{
			Config: &docker.Config{
				Image:        scanner,
				AttachStdout: true,
				AttachStderr: true,
				Tty:          true,
				Entrypoint:   args,
				Env: [
					"SCAN_API=" + os.Getenv("SCAN_API")
				]
			},
			HostConfig: &docker.HostConfig{
				Privileged: true,
				Binds:      binds,
			},
		})
	if err != nil {
		log.Println("FAIL")
		log.Println(err.Error())
		return err
	}
	log.Println(container.ID)
	err = client.StartContainer(container.ID, &docker.HostConfig{Privileged: true})
	if err != nil {
		log.Println("FAIL to start")
		log.Println(err.Error())
		return err
	}
	log.Println("Waiting")
	status, err := client.WaitContainer(container.ID)
	if err != nil {
		log.Println("FAIL to wait")
		log.Println(err.Error())
		return err
	}
	log.Printf("Done waiting %d", status)

	options := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
	}

	err = client.RemoveContainer(options)
	return err
}
