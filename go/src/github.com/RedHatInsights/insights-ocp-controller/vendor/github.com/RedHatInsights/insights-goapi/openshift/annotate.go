package annotate

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/RedHatInsights/insights-goapi/common"
)

//OpenshiftAnnotation representation
type OpenshiftAnnotation struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Timestamp   time.Time           `json:"timestamp"`
	Reference   string              `json:"reference"`
	Compliant   bool                `json:"compliant"`
	Summary     []map[string]string `json:"summary,omitempty"`
}

//InsightsAnnotator creates and manipulate Insights annotations
type InsightsAnnotator struct {
	version    string
	refBaseURL string
}

// NewInsightsAnnotator Create a new annotator instance
func NewInsightsAnnotator(version string, refBaseURL string) *InsightsAnnotator {

	a := &InsightsAnnotator{
		version:    version,
		refBaseURL: refBaseURL,
	}
	return a
}

func createSummary(criticalCount, highCount, mediumCount, lowCount int) []map[string]string {
	return []map[string]string{
		{
			"label":         "critical",
			"data":          fmt.Sprintf("%v", criticalCount),
			"severityIndex": fmt.Sprintf("%v", 3),
		},
		{
			"label":         "high",
			"data":          fmt.Sprintf("%v", highCount),
			"severityIndex": fmt.Sprintf("%v", 2),
		},
		{
			"label":         "medium",
			"data":          fmt.Sprintf("%v", mediumCount),
			"severityIndex": fmt.Sprintf("%v", 1),
		}, {
			"label":         "low",
			"data":          fmt.Sprintf("%v", lowCount),
			"severityIndex": fmt.Sprintf("%v", 0),
		},
	}
}

// CreateSecurityAnnotation result from scan response
func (a *InsightsAnnotator) CreateSecurityAnnotation(scanResp *common.ScanResponse, imageID string) *OpenshiftAnnotation {
	var summary []map[string]string
	var criticalCount, highCount, mediumCount, lowCount int
	for _, report := range scanResp.Reports {
		if report.Category == "Security" {
			switch report.Severity {
			case "CRITICAL":
				criticalCount++
			case "ERROR":
				highCount++
			case "WARN":
				mediumCount++
			case "INFO":
				lowCount++
			}
		}
	}
	if criticalCount+highCount+mediumCount+lowCount > 0 {
		summary = createSummary(criticalCount, highCount, mediumCount, lowCount)
	}
	return &OpenshiftAnnotation{
		"redhatinsights",
		"Security Insights",
		time.Now(),
		a.refBaseURL + "/" + imageID,
		true,
		summary,
	}
}

// CreateOperationsAnnotation from a scan result
func (a *InsightsAnnotator) CreateOperationsAnnotation(scanResp *common.ScanResponse, imageID string) *OpenshiftAnnotation {
	var summary []map[string]string
	var criticalCount, highCount, mediumCount, lowCount int
	for _, report := range scanResp.Reports {
		if report.Category != "Security" {
			switch report.Severity {
			case "CRITICAL":
				criticalCount++
			case "ERROR":
				highCount++
			case "WARN":
				mediumCount++
			case "INFO":
				lowCount++
			}
		}
	}
	if criticalCount+highCount+mediumCount+lowCount > 0 {
		summary = createSummary(criticalCount, highCount, mediumCount, lowCount)
	}
	return &OpenshiftAnnotation{
		"redhatinsights",
		"Stability, Performance and Availability Insights",
		time.Now(),
		a.refBaseURL + "/" + imageID,
		true,
		summary,
	}
}

//ToJSON - return json version of annotation
func (o *OpenshiftAnnotation) ToJSON() string {
	str, err := json.Marshal(o)
	if err != nil {
		//TODO log
		return "{}"
	}
	return string(str)
}
