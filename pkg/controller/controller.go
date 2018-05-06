package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/pflag"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/runtime"

	iclient "github.com/RedHatInsights/insights-goapi/client"
	"github.com/RedHatInsights/insights-goapi/common"
	"github.com/RedHatInsights/insights-goapi/container"
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
		log.Printf("Checking that image exists locally first...")
		if c.imageExists(image.DockerImageMetadata.ID) {
			log.Printf("Image exists.")
			log.Printf("Check in with Master Chief...")
			if c.canScan(image.DockerImageMetadata.ID) {
				log.Printf("Chief check-in successful.")
				log.Printf("Beginning scan.")
				// Scan the thing
				err := c.scanImage(image.DockerImageMetadata.ID,
					string(image.DockerImageReference),
					image.DockerImageMetadata.ID,
					image.GetName())
				// Check back in with the Chief (Dequeue)
				if err == nil {
					log.Printf("Scan completed successfully")
				} else {
					log.Printf("Scan completed with err %s ", err)
				}
				log.Printf("Removing from queue...")
				c.removeFromQueue(image.DockerImageMetadata.ID)
			}
		} else {
			log.Printf("Image does not exist.")
			log.Printf("Aborting scan.")
		}
	}

	return

}

func (c *Controller) removeFromQueue(id string) bool {
	// Setup API Request
	api := "http://" + os.Getenv("SCAN_API") + "/dequeue"
	req, err := http.NewRequest("POST", api+"/"+id, bytes.NewBufferString("{}"))
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
	if len(os.Getenv("MAX_RETRIES")) == 0 {
		maxRetries = 0
	} else {
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
	if len(os.Getenv("RETRY_SECONDS")) == 0 {
		retrySeconds = 60
		retrySecondsDuration = time.Duration(retrySeconds) * time.Second
	} else {
		retrySeconds, retrySecondsErr = strconv.Atoi(os.Getenv("RETRY_SECONDS"))
		retrySecondsDuration = time.Duration(retrySeconds) * time.Second
	}
	if retrySecondsErr != nil {
		log.Printf("Error reading RETRY_SECONDS from environment configuration:")
		log.Printf(retrySecondsErr.Error())
		log.Printf("Defaulting RETRY_SECONDS to 60.")
		retrySeconds = 60
		retrySecondsDuration = time.Duration(retrySeconds) * time.Second
	}

	// Check in with the Chief
	for keepTrying {
		// Only bother incrementing the counter if there is a defined limit
		if retryCounter != 0 {
			log.Printf("Keep trying Dequeue!")
		}
		if maxRetries != 0 {
			retryCounter = retryCounter + 1
			log.Printf("Max retries is %d and counter is at %d.", maxRetries, retryCounter)
		}

		// Make request to Chief
		client := &http.Client{Timeout: time.Second * 30}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Dequeue Client.Do(req) Error:")
			log.Printf(err.Error())
		} else {
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
		if (err == nil) && (resp.StatusCode == 204) {
			log.Printf("Dequeue successful.")
			dequeued = true
			keepTrying = false
			// 412 doesn't exist in queue, error
		} else if (err == nil) && (resp.StatusCode == 412) {
			log.Printf("Dequeue unsuccessful.")
			keepTrying = false
			// Otherwise wait, then retry
		} else {
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
	req, err := http.NewRequest("POST", api+"/"+id, bytes.NewBufferString("{}"))
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
	if len(os.Getenv("MAX_RETRIES")) == 0 {
		maxRetries = 0
	} else {
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
	if len(os.Getenv("RETRY_SECONDS")) == 0 {
		retrySeconds = 60
		retrySecondsDuration = time.Duration(retrySeconds) * time.Second
	} else {
		retrySeconds, retrySecondsErr = strconv.Atoi(os.Getenv("RETRY_SECONDS"))
		retrySecondsDuration = time.Duration(retrySeconds) * time.Second
	}
	if retrySecondsErr != nil {
		log.Printf("Error reading RETRY_SECONDS from environment configuration:")
		log.Printf(retrySecondsErr.Error())
		log.Printf("Defaulting RETRY_SECONDS to 60.")
		retrySeconds = 60
		retrySecondsDuration = time.Duration(retrySeconds) * time.Second
	}

	// Check in with the Chief
	isHalted := false
	for keepTrying {
		// Only bother incrementing the counter if there is a defined limit
		// and we are not in a suspended state
		if retryCounter != 0 {
			log.Printf("Keep trying Queue!")
		}
		if (maxRetries != 0) && (!isHalted) {
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
		} else {
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
		if (err == nil) && (resp.StatusCode == 201) {
			log.Printf("Master Chief says we can scan the image.")
			canScan = true
			keepTrying = false
			// If we get 423 then it is being scanned elsewhere
		} else if (err == nil) && (resp.StatusCode == 423) {
			log.Printf("Master Chief says someone else is scanning this image. Aborting.")
			keepTrying = false
			// If we get 412 then its been scanned in the past 24 hours
		} else if (err == nil) && (resp.StatusCode == 412) {
			log.Printf("Master Chief says this was scanned within the past 24 hours. Aborting.")
			keepTrying = false
			// If we get a 403 then the server has too many scan jobs going, try again after timeout
		} else if (err == nil) && (resp.StatusCode == 403) {
			log.Printf("Master Chief says too many concurrent scan jobs. Wait.")
			// If we get a 409 then HALT all scanning
		} else if (err == nil) && (resp.StatusCode == 409) {
			log.Printf("Master Chief says HALT.")
			log.Printf("Continue checking scan status with Chief for this image.")
			isHalted = true
			retryCounter = 0
			// If we have exceeded the MAX_RETRIES limit then stop
		} else if (retryCounter >= maxRetries) && (maxRetries != 0) {
			log.Printf("MAX_RETRIES exceeded. Stop.")
			keepTrying = false
			// Otherwise wait, then retry
		} else {
			log.Printf("Queue Request made. Waiting to begin next request.")
			time.Sleep(retrySecondsDuration)
		}
	}
	return canScan
}

func (c *Controller) scanImage(id string, imageRef string, imageSha string, openshiftSHA string) error {

	insightsReport, err := c.mountAndScan(id, imageRef, imageSha)
	if err == nil {
		log.Printf("Scan successful")
		c.postResults(insightsReport, openshiftSHA, imageRef)             //TODO handle error
		c.annotateImage(imageSha, openshiftSHA, imageRef, insightsReport) //TODO handle error
	}
	return err
}

func (c *Controller) postResults(results string, openshiftSHA string, imageRef string) {
	api := "http://" + os.Getenv("SCAN_API") + "/reports"
	req, err := http.NewRequest(
		"POST", api+"/"+openshiftSHA+"?name="+imageRef,
		bytes.NewBufferString(results))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
}

func (c *Controller) annotateImage(imageSha string, openshiftSHA string, imageRef string, annotation string) {
	log.Printf("Annotating local docker ID %s", imageSha)
	log.Printf("Annotating Openshift ID %s", openshiftSHA)
	c.updateImageAnnotationInfo(openshiftSHA, annotation)
}

func (c *Controller) updateImageAnnotationInfo(openshiftSha string, newInfo string) bool {

	if c.openshiftClient == nil {
		// if there's no OpenShift client, there can't be any image annotations
		return false
	}

	image, err := c.openshiftClient.Images().Get(openshiftSha)
	if err != nil {
		log.Printf("Job: Error getting image %s: %s\n", openshiftSha, err)
		return false
	}

	oldAnnotations := image.ObjectMeta.Annotations
	if oldAnnotations == nil {
		log.Printf("Image %s has no annotations - creating object.\n", openshiftSha)
		oldAnnotations = make(map[string]string)
	}

	annotator := annotate.NewInsightsAnnotator("0.1", c.getInsightsUILink())
	var res common.ScanResponse
	newInfoBytes := []byte(newInfo)
	json.Unmarshal(newInfoBytes, &res)
	secAnnotations := annotator.CreateSecurityAnnotation(&res, openshiftSha)
	opsAnnotations := annotator.CreateOperationsAnnotation(&res, openshiftSha)

	annotationValues := make(map[string]string)
	annotationValues["quality.images.openshift.io/vulnerability.redhatinsights"] = secAnnotations.ToJSON()
	annotationValues["quality.images.openshift.io/operations.redhatinsights"] = opsAnnotations.ToJSON()
	image.ObjectMeta.Annotations = annotationValues

	log.Printf("Annotate with information %s", annotationValues)

	image, err = c.openshiftClient.Images().Update(image)
	if err != nil {
		log.Printf("Error updating annotations for image: %s. %s\n", openshiftSha, err)
		return false
	}

	log.Println("Image annotated.")

	return true
}

func (c *Controller) mountAndScan(id string, imageRef string, imageSha string) (report string, err error) {

	scanDirectory := "/data/scanDir/scratch"
	//cleanup first
	os.RemoveAll(scanDirectory)
	os.MkdirAll(scanDirectory, os.ModePerm)

	scanOptions := container.NewDefaultImageMounterOptions()
	scanOptions.DstPath = scanDirectory
	scanOptions.Image = imageSha
	mounter := container.NewDefaultImageMounter(*scanOptions)
	_, image, _ := mounter.Mount()
	rhaiDir := path.Join(scanOptions.DstPath, "etc", "redhat-access-insights")
	machineidPath := path.Join(rhaiDir, "machine-id")
	os.MkdirAll(rhaiDir, os.ModePerm)
	err = ioutil.WriteFile(machineidPath, []byte("deeznuts"), 0644)
	if err != nil {
		fmt.Printf("ERROR: Scan failed writing machine ID %s", err)
		return "", err
	}
	scanner := iclient.NewDefaultScanner()
	_, out, err := scanner.ScanImage(scanOptions.DstPath, image.ID)
	if err != nil {
		fmt.Printf("ERROR: Scan failed %s", err)
		return "", err
	}
	report = string(*out)
	log.Printf("Scan results %s", report)
	return report, nil
}

func (c *Controller) getInsightsUILink() string {
	routeName := os.Getenv("SCAN_UI")
	if len(routeName) == 0 {
		routeName = "insights-ocp-ui"
	}
	routeAPI := c.openshiftClient.Routes("insights-scan") //TODO read this via downward api  metadata.namespace
	route, _ := routeAPI.Get(routeName)
	log.Printf("Route HOST is %s ", route.Spec.Host)
	return "https://" + route.Spec.Host

}
