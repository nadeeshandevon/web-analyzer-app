package contract

type WebAnalyzeRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type WebAnalyzeResponse struct {
	URL              string         `json:"url"`
	HTMLVersion      string         `json:"html_version"`
	Title            string         `json:"title"`
	Headings         map[string]int `json:"headings"`
	Links            LinkAnalysis   `json:"links"`
	HasLoginForm     bool           `json:"has_login_form"`
	Status           string         `json:"status"`
	ErrorDescription string         `json:"error_description"`
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
