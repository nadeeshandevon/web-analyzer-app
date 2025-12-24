package webanalyzer

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"web-analyzer-api/app/internal/core"
	"web-analyzer-api/app/internal/core/apperror"
	"web-analyzer-api/app/internal/model"
	"web-analyzer-api/app/internal/repository"
	htmlhelper "web-analyzer-api/app/internal/util/html"
	"web-analyzer-api/app/internal/util/logger"
	"web-analyzer-api/pkg/contract"

	"golang.org/x/net/html"
)

const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

type linkCheckResult struct {
	url          string
	statusCode   int
	isAccessible bool
}

type webAnalyzerService struct {
	log  *logger.Logger
	repo repository.WebAnalyzerRepository
	wg   *sync.WaitGroup
}

func NewWebAnalyzerService(logger *logger.Logger, repo repository.WebAnalyzerRepository) core.WebAnalyzerService {
	return &webAnalyzerService{
		log:  logger,
		repo: repo,
		wg:   &sync.WaitGroup{},
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
		return "", apperror.BadRequest("Failed to initialize analysis: " + err.Error())
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
		s.log.Warn("Analysis not found for update: " + analyzeId)
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
		s.UpdateAnalysisStatus(analysisId, StatusFailed, err.Error())
		return
	}

	resp, err := http.Get(baseURL.String())
	if err != nil {
		s.log.Error("Failed to fetch URL: " + err.Error())
		s.UpdateAnalysisStatus(analysisId, StatusFailed, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Error("URL returned error status: " + strconv.Itoa(resp.StatusCode))
		s.UpdateAnalysisStatus(analysisId, StatusFailed, "URL returned error status: "+strconv.Itoa(resp.StatusCode))
		return
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		s.log.Error("Failed to parse HTML: " + err.Error())
		s.UpdateAnalysisStatus(analysisId, StatusFailed, "Failed to parse HTML: "+err.Error())
		return
	}

	//Start fill analysis data
	analysis.HTMLVersion = htmlhelper.GetHTMLVersion(doc)
	analysis.Title = htmlhelper.GetTitle(doc)
	analysis.Headings = htmlhelper.GetHeadingsCount(doc)
	analysis.Links = s.analyzeLinks(ctx, doc, baseURL)
	analysis.HasLoginForm = htmlhelper.HasLoginForm(doc)
	analysis.Status = StatusSuccess
	//End fill analysis data

	_, err = s.repo.Update(*analysis)
	if err != nil {
		s.log.Error("Failed to update analysis result: " + err.Error())
	}
	s.log.Info("Background analysis completed for: " + baseURL.String())
}

func (s *webAnalyzerService) analyzeLinks(ctx context.Context, doc *html.Node, baseURL *url.URL) model.LinkAnalysis {
	links := htmlhelper.GetLinks(doc)

	analysis := model.LinkAnalysis{
		InaccessibleDetails: []model.InaccessibleLink{},
	}

	linksChan := make(chan string, len(links))
	resultsChan := make(chan linkCheckResult, len(links))

	numWorkers := 10
	if len(links) < numWorkers {
		numWorkers = len(links)
	}

	for i := 0; i < numWorkers; i++ {
		s.wg.Add(1)
		go s.linkCheckWorker(ctx, linksChan, resultsChan, baseURL)
	}

	for _, link := range links {
		linksChan <- link
	}

	close(linksChan)

	go func() {
		s.wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		if !result.isAccessible {
			analysis.Inaccessible++
			analysis.InaccessibleDetails = append(analysis.InaccessibleDetails, model.InaccessibleLink{
				URL:        result.url,
				StatusCode: result.statusCode,
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

func (s *webAnalyzerService) linkCheckWorker(ctx context.Context, linksChan <-chan string, resultsChan chan<- linkCheckResult, baseURL *url.URL) {
	defer s.wg.Done()
	defer s.log.Info("Link check worker stopped")

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for link := range linksChan {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result := s.checkLink(ctx, client, link, baseURL)
		if result != nil {
			select {
			case resultsChan <- *result:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (s *webAnalyzerService) checkLink(ctx context.Context, client *http.Client, link string, baseURL *url.URL) *linkCheckResult {
	if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "mailto:") {
		s.log.Warn("Invalid link: " + link)
		return nil
	}

	absoluteURL := link
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		parsedLink, err := url.Parse(link)
		if err != nil {
			s.log.Warn("Invalid link: " + link)
			return nil
		}
		absoluteURL = baseURL.ResolveReference(parsedLink).String()
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", absoluteURL, nil)
	if err != nil {
		s.log.Warn("Invalid link: " + link)
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		s.log.Debug("Inaccessible link: " + link)

		req, err = http.NewRequestWithContext(ctx, "GET", absoluteURL, nil)
		if err != nil {
			s.log.Debug("Inaccessible link: " + link)
			return nil
		}
		resp, err = client.Do(req)
		if err != nil {
			s.log.Debug("Inaccessible link: " + link)
			return &linkCheckResult{
				url:          absoluteURL,
				statusCode:   0,
				isAccessible: false,
			}
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		s.log.Debug("Inaccessible link: " + link + " with status code: " + strconv.Itoa(resp.StatusCode))
		return &linkCheckResult{
			url:          absoluteURL,
			statusCode:   resp.StatusCode,
			isAccessible: false,
		}
	}

	s.log.Debug("Accessible link: " + link + " with status code: " + strconv.Itoa(resp.StatusCode))
	return &linkCheckResult{
		url:          absoluteURL,
		statusCode:   resp.StatusCode,
		isAccessible: true,
	}
}
