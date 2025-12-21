package webanalyzer

import (
	"context"
	"net/http"
	"strings"
	"web-analyzer-api/internal/core"
	"web-analyzer-api/internal/core/apperror"
	"web-analyzer-api/internal/model"
	"web-analyzer-api/internal/repository"
	"web-analyzer-api/internal/util/logger"
	"web-analyzer-api/pkg/contract"

	"golang.org/x/net/html"
)

const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

type webAnalyzerService struct {
	log  *logger.Logger
	repo repository.WebAnalyzerRepository
}

func NewWebAnalyzerService(logger *logger.Logger, repo repository.WebAnalyzerRepository) core.WebAnalyzerService {
	return &webAnalyzerService{
		log:  logger,
		repo: repo,
	}
}

func (s *webAnalyzerService) AnalyzeWebsite(ctx context.Context, url string) (analysisId string, err error) {
	analysis := model.WebAnalyzer{
		URL:    url,
		Status: StatusPending,
	}

	analysisId, err = s.repo.Save(analysis)
	if err != nil {
		s.log.Error("Failed to save initial analysis: " + err.Error())
		return "", apperror.BadRequest("Failed to initialize analysis: " + err.Error())
	}

	// Call background analysis
	go s.processAnalysisJob(analysisId, url)

	return analysisId, nil
}

func (s *webAnalyzerService) processAnalysisJob(analysisId string, url string) {
	s.log.Info("Starting background analysis for: " + url)

	analysis := model.WebAnalyzer{
		ID: analysisId,
	}

	resp, err := http.Get(url)
	if err != nil {
		s.log.Error("Failed to fetch URL: " + err.Error())
		s.UpdateAnalysisStatus(analysisId, StatusFailed)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.UpdateAnalysisStatus(analysisId, StatusFailed)
		s.log.Error("URL returned error status: " + resp.Status)
		return
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		s.UpdateAnalysisStatus(analysisId, StatusFailed)
		s.log.Error("Failed to parse HTML: " + err.Error())
		return
	}

	analysis.HTMLVersion = getHTMLVersion(doc)
	analysis.Status = StatusSuccess

	_, err = s.repo.Update(analysis)
	if err != nil {
		s.log.Error("Failed to update analysis result: " + err.Error())
	}
	s.log.Info("Background analysis completed for: " + url)
}

func (s *webAnalyzerService) GetAnalyzeData(ctx context.Context, analyzeId string) (*contract.WebAnalyzeResponse, error) {

	result, err := s.repo.GetById(analyzeId)
	if err != nil {
		s.log.Error("Failed to get analysis data: " + err.Error())
		return nil, apperror.InternalServerError("Failed to get analysis data")
	}

	if result == nil {
		return nil, apperror.NotFound("Analysis not found")
	}

	analysis := contract.WebAnalyzeResponse{
		URL:         result.URL,
		HTMLVersion: result.HTMLVersion,
	}

	return &analysis, nil
}

func (s *webAnalyzerService) UpdateAnalysisStatus(analyzeId string, status string) {
	analysis := model.WebAnalyzer{
		ID: analyzeId,
	}

	_, err := s.repo.Update(analysis)
	if err != nil {
		s.log.Error("Failed to update analysis status: " + err.Error())
		return
	}
}

func getHTMLVersion(doc *html.Node) string {
	for n := doc.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.DoctypeNode {
			return getDocType(n)
		}
	}

	return "Unknown"
}

func getDocType(n *html.Node) string {
	if strings.EqualFold(n.Data, "html") && len(n.Attr) == 0 {
		return "HTML5"
	}

	for _, a := range n.Attr {
		val := strings.ToLower(a.Val)

		switch {
		case strings.Contains(val, "xhtml"):
			return "XHTML"
		case strings.Contains(val, "html 4.01"):
			return "HTML 4.01"
		case strings.Contains(val, "html 4.0"):
			return "HTML 4.0"
		}
	}

	return "Unknown"
}
