package controller

import (
	"os"
	"log"
	"sync"
	"io"
	"io/ioutil"
	"time"
	"bufio"
	"strings"
	"encoding/json"
	"net/http"
	"bytes"
	"strconv"

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

	// Get the list of images to scan
	for _, image := range imageList.Items {
		log.Printf("Scanning image %s %s", image.DockerImageMetadata.ID, image.DockerImageReference)

		// Check in to schedule the scan
		log.Printf("Checking in with Master Chief...")
		if c.canScan(image.DockerImageMetadata.ID) {
			log.Printf("Check in successful.");
			if c.imageExists(image.DockerImageMetadata.ID) {
				log.Printf("Beginning scan.");
				// Scan the thing
				c.scanImage(image.DockerImageMetadata.ID,
				getScanArgs(string(image.DockerImageReference), "/tmp/image-content8"),
				string(image.DockerImageReference),
				image.DockerImageMetadata.ID)
				// Check back in with the Chief (Dequeue)
				log.Printf("Removing from queue...")
				c.removeFromQueue(image.DockerImageMetadata.ID)
			}
		} else {
			log.Printf("Check in not succesful.");
			log.Printf("Aborting scan.");
		}
	}

	return

}

func (c *Controller) removeFromQueue(id string) bool {
	// Setup API Request
	api := "http://" + os.Getenv("SCAN_API") + "/dequeue"
	req, err := http.NewRequest("POST", api + "/" + id, bytes.NewBufferString("{}"))
	if err != nil {
		log.Printf("Error setting up new request to Master Chief:")
		log.Fatalf(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	dequeued := false

	// Flag to stop trying to communicate with the Chief
	// We always keep trying to check in with the Chief
	// If MAX_RETRIES is 0, then try forever, otherwise X number of times
	// Set RETRY_SECONDS for number of seconds to wat between Chief calls
	keepTrying := true
	retryCounter := 0
	var maxRetries int
	var maxRetriesErr error
	if len(os.Getenv("MAX_RETRIES")) == 0{
		maxRetries = 0
	}else{
		maxRetries, maxRetriesErr = strconv.Atoi(os.Getenv("MAX_RETRIES"))
	}
	if maxRetriesErr != nil {
		log.Printf("Error reading MAX_RETRIES from environment configuration:")
		log.Printf(maxRetriesErr.Error())
		log.Printf("Defaulting MAX_RETRIES to 0, infinite.")
		maxRetries = 0
	}
	var retrySeconds int
	var retrySecondsErr error
	var retrySecondsDuration time.Duration
	if len(os.Getenv("RETRY_SECONDS")) == 0{
		retrySeconds = 60
		retrySecondsDuration = time.Duration(retrySeconds)*time.Second
	}else{
		retrySeconds, retrySecondsErr = strconv.Atoi(os.Getenv("RETRY_SECONDS"))
		retrySecondsDuration = time.Duration(retrySeconds)*time.Second
	}
	if retrySecondsErr != nil {
		log.Printf("Error reading RETRY_SECONDS from environment configuration:")
		log.Printf(retrySecondsErr.Error())
		log.Printf("Defaulting RETRY_SECONDS to 60.")
		retrySeconds = 60
		retrySecondsDuration = time.Duration(retrySeconds)*time.Second
	}

	// Check in with the Chief
	for keepTrying {
		// Only bother incrementing the counter if there is a defined limit
		if ( retryCounter != 0 ){
			log.Printf("Keep trying Dequeue!")
		}
		if ( maxRetries != 0 ){
			retryCounter = retryCounter + 1
			log.Printf("Max retries is %s and counter is at %s.", maxRetries, retryCounter)
		}

		// Make request to Chief
		client := &http.Client{Timeout: time.Second * 30}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Dequeue Client.Do(req) Error:")
			log.Printf(err.Error())
		}else{
			defer resp.Body.Close()
			body, readAllErr := ioutil.ReadAll(resp.Body)
			if readAllErr != nil {
				log.Printf("Dequeue Client ioutil.ReadAll Error:")
				log.Printf(readAllErr.Error())
			}
			log.Printf("Master Chief Dequeue Status: %s", resp.Status)
			log.Printf("Master Chief Dequeue Body: %s", body)
		}
		
		// 204 successful dequeue
		if (err == nil) && (resp.StatusCode == 204){
			log.Printf("Dequeue successful.")	
			dequeued = true
			keepTrying = false
		// 412 doesn't exist in queue, error
		}else if (err == nil) && (resp.StatusCode == 412){
			log.Printf("Dequeue unsuccessful.")
			keepTrying = false
		// Otherwise wait, then retry
		}else{
			log.Printf("Dequeue Request made. Waiting to begin next request.")
			time.Sleep(retrySecondsDuration)
		}
	}
	return dequeued
}

func (c *Controller) imageExists(id string) bool {
    endpoint := "unix:///var/run/docker.sock"
    client, dockerErr := docker.NewVersionedClient(endpoint, "1.22")
    if dockerErr != nil {
        log.Printf("Error creating docker client: %s\n", dockerErr)
        return false
    }

    _, inspectErr := client.InspectImage(id)
    if inspectErr != nil {
        log.Printf("Error testing if image %s exists: %s\n", id, inspectErr)
        return false
    }
    return true
}

func (c *Controller) canScan(id string) bool {
	// Setup API Request
	api := "http://" + os.Getenv("SCAN_API") + "/queue"
	req, err := http.NewRequest("POST", api + "/" + id, bytes.NewBufferString("{}"))
	if err != nil {
		log.Printf("Error setting up new request to Master Chief:")
		log.Fatalf(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	canScan := false

	// Flag to stop trying to communicate with the Chief
	// We always keep trying to check in with the Chief
	// If MAX_RETRIES is 0, then try forever, otherwise X number of times
	// Set RETRY_SECONDS for number of seconds to wat between Chief calls
	keepTrying := true
	retryCounter := 0
	var maxRetries int
	var maxRetriesErr error
	if len(os.Getenv("MAX_RETRIES")) == 0{
		maxRetries = 0
	}else{
		maxRetries, maxRetriesErr = strconv.Atoi(os.Getenv("MAX_RETRIES"))
	}
	if maxRetriesErr != nil {
		log.Printf("Error reading MAX_RETRIES from environment configuration:")
		log.Printf(maxRetriesErr.Error())
		log.Printf("Defaulting MAX_RETRIES to 0, infinite.")
		maxRetries = 0
	}
	var retrySeconds int
	var retrySecondsErr error
	var retrySecondsDuration time.Duration
	if len(os.Getenv("RETRY_SECONDS")) == 0{
		retrySeconds = 60
		retrySecondsDuration = time.Duration(retrySeconds)*time.Second
	}else{
		retrySeconds, retrySecondsErr = strconv.Atoi(os.Getenv("RETRY_SECONDS"))
		retrySecondsDuration = time.Duration(retrySeconds)*time.Second
	}
	if retrySecondsErr != nil {
		log.Printf("Error reading RETRY_SECONDS from environment configuration:")
		log.Printf(retrySecondsErr.Error())
		log.Printf("Defaulting RETRY_SECONDS to 60.")
		retrySeconds = 60
		retrySecondsDuration = time.Duration(retrySeconds)*time.Second
	}

	// Check in with the Chief
	isHalted := false
	for keepTrying {
		// Only bother incrementing the counter if there is a defined limit
		// and we are not in a suspended state
		if ( retryCounter != 0 ){
			log.Printf("Keep trying Queue!")
		}
		if ( maxRetries != 0 ) && ( !isHalted ){
			retryCounter = retryCounter + 1
			log.Printf("Max retries is %s and counter is at %s.", maxRetries, retryCounter)
		}

		// Reset isHalted
		isHalted = false
		
		// Make request to Chief
		client := &http.Client{Timeout: time.Second * 30}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Queue Client.Do(req) Error:")
			log.Printf(err.Error())
		}else{
			defer resp.Body.Close()
			body, readAllErr := ioutil.ReadAll(resp.Body)
			if readAllErr != nil {
				log.Printf("Queue Client ioutil.ReadAll Error:")
				log.Printf(readAllErr.Error())
			}
			log.Printf("Master Chief Queue Status: %s", resp.Status)
			log.Printf("Master Chief Queue Body: %s", body)
		}
		
		// If we get 201 then were good to go
		if (err == nil) && (resp.StatusCode == 201){
			log.Printf("Master Chief says we can scan the image.")
			canScan = true
			keepTrying = false
		// If we get 423 then it is being scanned elsewhere
		}else if (err == nil) && (resp.StatusCode == 423){
			log.Printf("Master Chief says someone else is scanning this image. Aborting.")
			keepTrying = false
		// If we get 412 then its been scanned in the past 24 hours
		}else if (err == nil) && (resp.StatusCode == 412){
			log.Printf("Master Chief says this was scanned within the past 24 hours. Aborting.")
			keepTrying = false
		// If we get a 403 then the server has too many scan jobs going, try again after timeout
		} else if (err == nil) && (resp.StatusCode == 403){
			log.Printf("Master Chief says too many concurrent scan jobs. Wait.")
		// If we get a 409 then HALT all scanning
		} else if (err == nil) && (resp.StatusCode == 409){
			log.Printf("Master Chief says HALT.")
			log.Printf("Continue checking scan status with Chief for this image.")
			isHalted = true
			retryCounter = 0
		// If we have exceeded the MAX_RETRIES limit then stop
		}else if(retryCounter >= maxRetries) && (maxRetries != 0){
			log.Printf("MAX_RETRIES exceeded. Stop.")
			keepTrying = false
		// Otherwise wait, then retry
		}else{
			log.Printf("Queue Request made. Waiting to begin next request.")
			time.Sleep(retrySecondsDuration)
		}
	}
	return canScan
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
							  "INSIGHTS_AUTHMETHOD=" + os.Getenv("INSIGHTS_AUTHMETHOD"),
							  "INSIGHTS_PROXY=" + os.Getenv("INSIGHTS_PROXY")},
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

	annotator := annotate.NewInsightsAnnotator("0.1", os.Getenv("SCAN_UI"))
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