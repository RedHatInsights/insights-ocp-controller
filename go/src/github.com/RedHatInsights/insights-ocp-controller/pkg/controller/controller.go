package controller

import (
	"os"
	"log"
	"sync"
	"io"
	"time"
	"bufio"
	"strings"
	"encoding/json"
	"net/http"
	"bytes"

	"github.com/fsouza/go-dockerclient"
	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/pflag"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/RedHatInsights/insights-goapi/common"
	"github.com/RedHatInsights/insights-goapi/openshift"
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
	args = append(args, "./insights-scanner")
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
		log.Printf("--")
		log.Printf("%+v\n", image)
		log.Printf("--")
		log.Printf("%+v\n", image.DockerImageMetadata)
		log.Printf("--")
		log.Printf("Scanning image %s %s", image.DockerImageMetadata.ID, image.DockerImageReference)
		c.scanImage(image.DockerImageMetadata.ID,
			getScanArgs(string(image.DockerImageReference), "/tmp/image-content8"),
			string(image.DockerImageReference),
			image.DockerImageMetadata.ID)
	}

	// Force known image scan
	// c.scanImage("image.DockerImageMetadata.ID",
	// 		getScanArgs("openshift/wildfly-100-centos7", "/tmp/image-content8"),
	// 		"openshift/wildfly-100-centos7",
	// 		"sha256:01fde7095217610427a3fb133e0ff6003cc5958f65e956fa58aecde3f57d45ff")
	return

}

func (c *Controller) scanImage(id string, args []string, imageRef string, imageSha string) error {
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewVersionedClient(endpoint, "1.22")
	binds := []string{}
	binds = append(binds, "/var/run/docker.sock:/var/run/docker.sock")

	container, err := client.CreateContainer(
		docker.CreateContainerOptions{
			Config: &docker.Config{
				Image:        os.Getenv("SCANNER_IMAGE"),
				AttachStdout: true,
				AttachStderr: true,
				Tty:          true,
				Entrypoint:   args,
				Env: []string{"SCAN_API=" + os.Getenv("INSIGHTS_OCP_API_SERVICE_HOST") + ":8080",
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

	var insightsReport string
	go func(reader io.Reader, sr chan ScanResult) {
		scanner := bufio.NewScanner(reader)
		scan := ScanResult{completed: false, scanId: ""}

		for scanner.Scan() {
			out := scanner.Text()
			if strings.Contains(out, "Post Scan...") {
				log.Printf("Found completed scan with result.\n")
				scan.completed = true
			}
			log.Printf("%s\n", out)
			insightsReport = out

			if strings.Contains(out, "ScanContainerView{scanId=") {
				cmd := strings.Split(out, "ScanContainerView{scanId=")
				eos := strings.Index(cmd[1], ",")
				scan.scanId = cmd[1][:eos]
				log.Printf("Found scan ID %s with result.\n", scan.scanId)
			}
		}

		log.Printf("Placing scan result %t from scanId %s into channel.\n", scan.completed, scan.scanId)
		sr <- scan

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

	if (len(insightsReport) > 0 && !strings.HasPrefix(insightsReport, "ERROR:")) {
		c.postResults(insightsReport, imageSha)
		c.annotateImage(imageRef, imageSha, insightsReport)
	}

	options := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
	}

	err = client.RemoveContainer(options)

	abort <- true
	return err
}

func (c *Controller) postResults(results string, imageSha string) {
	api := "http://" + os.Getenv("SCAN_API") + "/reports"
	req, err := http.NewRequest("POST", api + "/" + imageSha, bytes.NewBufferString(results))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
}

func (c *Controller) annotateImage(imageRef string, imageSha string, annotation string){
	log.Printf("Annotating %s", imageRef)
	log.Printf("Annotating %s", imageSha)
	c.UpdateImageAnnotationInfo(imageSha, annotation)
}

func (c *Controller) UpdateImageAnnotationInfo(imageSha string, newInfo string) bool {

	if c.openshiftClient == nil {
		// if there's no OpenShift client, there can't be any image annotations
		return false
	}

	image, err := c.openshiftClient.Images().Get(imageSha)
	if err != nil {
		log.Printf("Job: Error getting image %s: %s\n", imageSha, err)
		return false
	}

	oldAnnotations := image.ObjectMeta.Annotations
	if oldAnnotations == nil {
		log.Printf("Image %s has no annotations - creating object.\n", imageSha)
		oldAnnotations = make(map[string]string)
	}

	annotator := annotate.NewInsightsAnnotator("0.1", "https://openshift.com/insights")
	var res common.ScanResponse
	newInfoBytes := []byte(newInfo)
	json.Unmarshal(newInfoBytes, &res)
	secAnnotations := annotator.CreateSecurityAnnotation(&res, imageSha)
	opsAnnotations := annotator.CreateOperationsAnnotation(&res, imageSha)

	annotationValues := make(map[string]string)
	annotationValues["quality.images.openshift.io/vulnerability.redhatinsights"] = secAnnotations.ToJSON()
	annotationValues["quality.images.openshift.io/operations.redhatinsights"] = opsAnnotations.ToJSON()
	image.ObjectMeta.Annotations = annotationValues

	log.Println("Annotate with information %s", annotationValues)

	image, err = c.openshiftClient.Images().Update(image)
	if err != nil {
		log.Printf("Error updating annotations for image: %s. %s\n", imageSha, err)
		return false
	}

	log.Println("Image annotated.")

	return true
}
