package common

type System struct {
	Hostname string                 `json:"hostname"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Content struct {
	Plain string `json:"plain"`
	Html  string `json:"html"`
}

type Report struct {
	RuleData       map[string]interface{} `json:"rule_data"`
	Title          *Content               `json:"title"`
	Summary        *Content               `json:"summary"`
	Description    *Content               `json:"description"`
	Details        *Content               `json:"details"`
	Reference      *Content               `json:"reference"`
	Resolution     *Content               `json:"resolution"`
	Severity       string                 `json:"severity"`
	Category       string                 `json:"category"`
	Impact         int                    `json:"impact"`
	Likelihood     int                    `json:"likelihood"`
	RebootRequired bool                   `json:"reboot_required"`
	Acks           []interface{}          `json:"acks"`
}

type ScanResponse struct {
	Version string                 `json:"version"`
	System  *System                `json:"system"`
	Reports map[string]Report      `json:"reports"`
	Upload  map[string]interface{} `json:"upload"`
}
