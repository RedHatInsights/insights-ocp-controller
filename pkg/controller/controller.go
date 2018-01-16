package controller

import (
	"os"
	"log"
	"sync"
	"io"
	"time"
	"bufio"
	"strings"

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

type ScanResult struct {
	completed bool
	scanId    string
}

func getScanArgs(imageID string, mountPoint string) []string {

	args := []string{}
	args = append(args, "/insights-scanner")
	args = append(args, "-image")
	args = append(args, imageID)
	args = append(args, "-mount_path")
	args = append(args, mountPoint)

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
	c.scanImage("image.DockerImageMetadata.ID", getScanArgs("registry.access.redhat.com/rhscl/postgresql-94-rhel7", "/tmp/image-content8"))
	return

}

func (c *Controller) scanImage(id string, args []string) error {
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewVersionedClient(endpoint, "1.22")
	binds := []string{}
	binds = append(binds, "/var/run/docker.sock:/var/run/docker.sock")
	scanner := "redhatinsights/insights-scanner"

	container, err := client.CreateContainer(
		docker.CreateContainerOptions{
			Config: &docker.Config{
				Image:        scanner,
				AttachStdout: true,
				AttachStderr: true,
				Tty:          true,
				Entrypoint:   args,
				Env: []string{"SCAN_API=" + os.Getenv("SCAN_API"),
							  "INSIGHTS_USERNAME=" + os.Getenv("INSIGHTS_USERNAME"),
							  "INSIGHTS_PASSWORD=" + os.Getenv("INSIGHTS_PASSWORD"),
							  "INSIGHTS_AUTHMETHOD=" + os.Getenv("INSIGHTS_AUTHMETHOD")},
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

	done := make(chan ScanResult)
	abort := make(chan bool, 1)
	r, w := io.Pipe()

    monitorOptions := docker.AttachToContainerOptions{
        Container:    container.ID,
        OutputStream: w,
        ErrorStream:  w,
        Stream:       true,
        Stdout:       true,
        Stderr:       true,
        Logs:         true,
        RawTerminal:  true,
    }

    go client.AttachToContainer(monitorOptions) // will block so isolate

    go func(reader *io.PipeReader, a chan bool) {

		for {
			time.Sleep(time.Second)
			select {
			case _ = <-a:
				log.Printf("Received IO shutdown for scanner.\n")
				reader.Close()
				return

			default:
			}

		}

	}(r, abort)

	go func(reader io.Reader, c chan ScanResult) {
		scanner := bufio.NewScanner(reader)
		scan := ScanResult{completed: false, scanId: ""}

		for scanner.Scan() {
			out := scanner.Text()
			if strings.Contains(out, "Post Scan...") {
				log.Printf("Found completed scan with result.\n")
				scan.completed = true
			}
			log.Printf("%s\n", out)

			if strings.Contains(out, "ScanContainerView{scanId=") {
				cmd := strings.Split(out, "ScanContainerView{scanId=")
				eos := strings.Index(cmd[1], ",")
				scan.scanId = cmd[1][:eos]
				log.Printf("Found scan ID %s with result.\n", scan.scanId)
			}
		}

		log.Printf("Placing scan result %t from scanId %s into channel.\n", scan.completed, scan.scanId)
		c <- scan

	}(r, done)

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

	abort <- true


	return err
}
