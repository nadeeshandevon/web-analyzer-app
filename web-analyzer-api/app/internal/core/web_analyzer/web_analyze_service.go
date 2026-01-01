package webanalyzer

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"web-analyzer-api/app/internal/contract"
	"web-analyzer-api/app/internal/core"
	"web-analyzer-api/app/internal/core/apperror"
	"web-analyzer-api/app/internal/model"
	"web-analyzer-api/app/internal/repository"
	htmlhelper "web-analyzer-api/app/internal/util/html"
	"web-analyzer-api/app/internal/util/logger"

	"golang.org/x/net/html"
)

const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

type webAnalyzerService struct {
	log         *logger.Logger
	repo        repository.WebAnalyzerRepository
	linkChecker core.LinkChecker
}

func NewWebAnalyzerService(logger *logger.Logger, repo repository.WebAnalyzerRepository, linkChecker core.LinkChecker) core.WebAnalyzerService {
	return &webAnalyzerService{
		log:         logger,
		repo:        repo,
		linkChecker: linkChecker,
	}
}

func (s *webAnalyzerService) AnalyzeWebsite(ctx context.Context, baseURL *url.URL) (analysisId string, err error) {
	analysis := model.WebAnalyzer{
		URL:    baseURL.String(),
		Status: StatusPending,
	}

	analysisId, err = s.repo.Save(analysis)
	if err != nil {
		s.log.Error("Failed to save initial analysis: " + err.Error())
		return "", apperror.BadRequest("Failed to save initial analysis data")
	}

	// Call background analysis
	go s.processAnalysisJob(context.Background(), analysisId, baseURL)

	return analysisId, nil
}

func (s *webAnalyzerService) GetAnalyzeData(ctx context.Context, analyzeId string) (*contract.WebAnalyzeResponse, error) {
	result, err := s.repo.GetById(analyzeId)
	if err != nil {
		s.log.Error("Failed to get analysis data: " + err.Error())
		return nil, apperror.InternalServerError("Failed to get analysis data")
	}

	if result == nil {
		s.log.Warn("Analysis result not found")
		return nil, apperror.NotFound("Analysis result not found")
	}

	var inaccessibleDetails []contract.InaccessibleLink
	if result.Links.InaccessibleDetails != nil {
		inaccessibleDetails = make([]contract.InaccessibleLink, len(result.Links.InaccessibleDetails))
		for i, detail := range result.Links.InaccessibleDetails {
			inaccessibleDetails[i] = contract.InaccessibleLink{
				URL:        detail.URL,
				StatusCode: detail.StatusCode,
			}
		}
	}

	var errorDescription string
	if result.ErrorDescription != nil && *result.ErrorDescription != "" {
		errorDescription = *result.ErrorDescription
	}

	analysis := contract.WebAnalyzeResponse{
		URL:         result.URL,
		HTMLVersion: result.HTMLVersion,
		Title:       result.Title,
		Headings:    result.Headings,
		Links: contract.LinkAnalysis{
			Internal:            result.Links.Internal,
			External:            result.Links.External,
			Inaccessible:        result.Links.Inaccessible,
			InaccessibleDetails: inaccessibleDetails,
		},
		HasLoginForm:     result.HasLoginForm,
		Status:           result.Status,
		ErrorDescription: errorDescription,
	}

	return &analysis, nil
}

func (s *webAnalyzerService) UpdateAnalysisStatus(analyzeId string, status string, errorDescription string) {
	analysis, err := s.repo.GetById(analyzeId)
	if err != nil {
		s.log.Error("Failed to get analysis data: " + err.Error())
		return
	}

	if analysis == nil {
		s.log.Warn("Analysis not found for update: analyzeId - " + analyzeId)
		return
	}

	analysis.Status = status

	if errorDescription != "" {
		analysis.ErrorDescription = &errorDescription
	}

	_, err = s.repo.Update(*analysis)
	if err != nil {
		s.log.Error("Failed to update analysis status: " + err.Error())
		return
	}
}

func (s *webAnalyzerService) processAnalysisJob(ctx context.Context, analysisId string, baseURL *url.URL) {
	s.log.Info("Starting background analysis for: " + baseURL.String())
	analysis, err := s.repo.GetById(analysisId)
	if err != nil {
		s.log.Error("Failed to get analysis data: " + err.Error())
		s.UpdateAnalysisStatus(analysisId, StatusFailed, "Failed to get analysis data.")
		return
	}

	resp, err := http.Get(baseURL.String())
	if err != nil {
		s.log.Error("Failed to fetch URL: " + err.Error())
		s.UpdateAnalysisStatus(analysisId, StatusFailed, "URL cannot be accessed. URL is invalid or unreachable.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Error("URL returned error status: " + strconv.Itoa(resp.StatusCode))
		s.UpdateAnalysisStatus(analysisId, StatusFailed, "URL cannot be accessed. URL is invalid or unreachable.")
		return
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		s.log.Error("Failed to parse HTML: " + err.Error())
		s.UpdateAnalysisStatus(analysisId, StatusFailed, "Failed to parse HTML content.")
		return
	}

	//Start fill analysis data

	// 1. Start link analysis from the parsed HTML document
	analysis.Links = s.analyzeLinks(ctx, doc, baseURL)
	// End link analysis from the parsed HTML document

	// 2. Start metadata extraction from the parsed HTML document
	s.analyzeMetadata(doc, analysis)
	// End metadata extraction from the parsed HTML document

	analysis.Status = StatusSuccess
	//End fill analysis data

	_, err = s.repo.Update(*analysis)
	if err != nil {
		s.log.Error("Failed to update analysis result: " + err.Error())
	}
	s.log.Info("Background analysis completed for: " + baseURL.String())
}

func (s *webAnalyzerService) analyzeMetadata(doc *html.Node, analysis *model.WebAnalyzer) {
	analysis.HTMLVersion = htmlhelper.GetHTMLVersion(doc)
	analysis.Title = htmlhelper.GetTitle(doc)
	analysis.Headings = htmlhelper.GetHeadingsCount(doc)
	analysis.HasLoginForm = htmlhelper.HasLoginForm(doc)
}

func (s *webAnalyzerService) analyzeLinks(ctx context.Context, doc *html.Node, baseURL *url.URL) model.LinkAnalysis {
	links := htmlhelper.GetLinks(doc)

	analysis := model.LinkAnalysis{
		InaccessibleDetails: []model.InaccessibleLink{},
	}

	linksChan := make(chan string, len(links))
	resultsChan := make(chan model.LinkCheckResult, len(links))

	numWorkers := 10
	if len(links) < numWorkers {
		numWorkers = len(links)
	}

	wg := &sync.WaitGroup{}
	//Start check links workers in background
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go s.linkChecker.RunWorker(ctx, linksChan, resultsChan, baseURL, wg)
	}
	//End check links workers in background

	for _, link := range links {
		linksChan <- link
	}

	close(linksChan)

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		if !result.IsAccessible {
			analysis.Inaccessible++
			analysis.InaccessibleDetails = append(analysis.InaccessibleDetails, model.InaccessibleLink{
				URL:        result.URL,
				StatusCode: result.StatusCode,
			})
		}
	}

	for _, link := range links {
		if htmlhelper.IsInternalLink(link, baseURL) {
			analysis.Internal++
		} else {
			analysis.External++
		}
	}

	return analysis
}
