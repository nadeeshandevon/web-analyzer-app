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
	"web-analyzer-api/app/internal/model"
	"web-analyzer-api/app/internal/util/logger"
)

type linkChecker struct {
	log *logger.Logger
}

func NewLinkChecker(log *logger.Logger) core.LinkChecker {
	return &linkChecker{
		log: log,
	}
}

func (lc *linkChecker) RunWorker(ctx context.Context, linksChan <-chan string, resultsChan chan<- model.LinkCheckResult, baseURL *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	defer lc.log.Info("Link check worker stopped")

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for {
		select {
		case <-ctx.Done():
			return
		case link, ok := <-linksChan:
			if !ok {
				return
			}

			result := lc.CheckLink(ctx, client, link, baseURL)
			if result != nil {
				select {
				case resultsChan <- *result:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (lc *linkChecker) CheckLink(ctx context.Context, client *http.Client, link string, baseURL *url.URL) *model.LinkCheckResult {
	if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "mailto:") {
		lc.log.Warn("Invalid link: " + link)
		return nil
	}

	absoluteURL := link
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		parsedLink, err := url.Parse(link)
		if err != nil {
			lc.log.Warn("Invalid link: " + link)
			return nil
		}
		absoluteURL = baseURL.ResolveReference(parsedLink).String()
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", absoluteURL, nil)
	if err != nil {
		lc.log.Warn("Invalid link: " + link)
		return nil
	}

	resp, err := client.Do(req)
	needsFallback := err != nil || (resp != nil && resp.StatusCode >= 400)

	if needsFallback {
		if resp != nil {
			resp.Body.Close()
		}

		lc.log.Debug("HEAD failed or returned error, trying GET: " + absoluteURL)
		req, err = http.NewRequestWithContext(ctx, "GET", absoluteURL, nil)
		if err != nil {
			lc.log.Debug("Failed to create GET request: " + err.Error())
			return nil
		}
		resp, err = client.Do(req)
		if err != nil {
			lc.log.Debug("Inaccessible link (GET failed): " + absoluteURL)
			return &model.LinkCheckResult{
				URL:          absoluteURL,
				StatusCode:   0,
				IsAccessible: false,
			}
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		lc.log.Debug("Inaccessible link: " + link + " with status code: " + strconv.Itoa(resp.StatusCode))
		return &model.LinkCheckResult{
			URL:          absoluteURL,
			StatusCode:   resp.StatusCode,
			IsAccessible: false,
		}
	}

	lc.log.Debug("Accessible link: " + link + " with status code: " + strconv.Itoa(resp.StatusCode))
	return &model.LinkCheckResult{
		URL:          absoluteURL,
		StatusCode:   resp.StatusCode,
		IsAccessible: true,
	}
}
