package model

import "time"

type WebAnalyzer struct {
	ID           string
	URL          string
	HTMLVersion  string
	Title        string
	Headings     map[string]int
	Links        LinkAnalysis
	HasLoginForm bool
	Status       string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type LinkAnalysis struct {
	Internal            int
	External            int
	Inaccessible        int
	InaccessibleDetails []InaccessibleLink
}

type InaccessibleLink struct {
	URL        string
	StatusCode int
}
