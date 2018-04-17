package annotate

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/RedHatInsights/insights-goapi/common"
)

func TestSecurityAnnotationCreate(t *testing.T) {
	resp, err := ioutil.ReadFile("testdata/securityhits.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	annotator := &InsightsAnnotator{version: "0.1", refBaseURL: "https://openshift.com/insights"}
	var res common.ScanResponse
	json.Unmarshal(resp, &res)
	secAnnotation := annotator.CreateSecurityAnnotation(&res, "SHA123456")
	t.Log(secAnnotation.ToJSON())

}
func TestOpsAnnotationCreate(t *testing.T) {
	resp, err := ioutil.ReadFile("testdata/operationhits.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	annotator := &InsightsAnnotator{version: "0.1", refBaseURL: "https://openshift.com/insights"}
	var res common.ScanResponse
	json.Unmarshal(resp, &res)
	opsAnnotation := annotator.CreateOperationsAnnotation(&res, "SHA123456")
	if opsAnnotation.Reference != "https://openshift.com/insights/SHA123456" {
		t.Fatalf("Reference URL b")
	}
	t.Log(opsAnnotation.ToJSON())
}
func TestEmptySummaryAnnotationCreate(t *testing.T) {
	resp, err := ioutil.ReadFile("testdata/emptyhits.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	annotator := &InsightsAnnotator{version: "0.1", refBaseURL: "https://openshift.com/insights"}
	var res common.ScanResponse
	json.Unmarshal(resp, &res)
	a := annotator.CreateOperationsAnnotation(&res, "SHA123456")
	if strings.Contains(a.ToJSON(), "summary") {
		t.Fatalf("Summary set when there are no hits")
	}
}
func TestReferenceURLCreate(t *testing.T) {
	mockResponse := common.ScanResponse{
		Version: "0.0",
		System:  nil,
		Reports: nil,
		Upload:  nil,
	}
	annotator := &InsightsAnnotator{version: "0.1", refBaseURL: "https://openshift.com/insights"}
	a := annotator.CreateOperationsAnnotation(&mockResponse, "OPS1234")
	b := annotator.CreateSecurityAnnotation(&mockResponse, "SEC1234")
	if a.Reference != "https://openshift.com/insights/OPS1234" {
		t.Fatalf("Reference URL is not https://openshift.com/insights/OPS1234")
	}
	if b.Reference != "https://openshift.com/insights/SEC1234" {
		t.Fatalf("Reference URL  is not https://openshift.com/insights/OPS1234")
	}
}
