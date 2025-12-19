package contract

type WebAnalyzeRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type WebAnalyzeResponse struct {
	URL          string         `json:"url"`
	HTMLVersion  string         `json:"html_version"`
	Title        string         `json:"title"`
	Headings     map[string]int `json:"headings"`
	Links        LinkAnalysis   `json:"links"`
	HasLoginForm bool           `json:"has_login_form"`
}

type LinkAnalysis struct {
	Internal            int                `json:"internal"`
	External            int                `json:"external"`
	Inaccessible        int                `json:"inaccessible"`
	InaccessibleDetails []InaccessibleLink `json:"inaccessible_details"`
}

type InaccessibleLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
}

type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code,omitempty"`
}
