package client

import (
	"encoding/json"
	"os/exec"
	"fmt"


	"github.com/RedHatInsights/insights-goapi/common"
)

const clientCmd = "insights-client --no-gpg --analyze-mountpoint=%s"

type Scanner interface {
	
	ScanImage(contentPath string, imageId string) (*common.ScanResponse, *[]byte, error)

}

type DefaultScanner struct{}

func NewDefaultScanner()(Scanner){
	return &DefaultScanner{}
}

func (s * DefaultScanner )ScanImage(contentPath string, imageId string) (*common.ScanResponse, *[]byte, error) {
	cmdStr := fmt.Sprintf(clientCmd, contentPath)
	var scanResp common.ScanResponse
	jsonResp, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
    if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(jsonResp, &scanResp)
	return &scanResp, &jsonResp, err
}
