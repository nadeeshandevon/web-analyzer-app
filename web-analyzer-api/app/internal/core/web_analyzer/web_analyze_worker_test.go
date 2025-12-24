package webanalyzer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
	"web-analyzer-api/app/internal/core"
	"web-analyzer-api/app/internal/model"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/stretchr/testify/assert"
)

func TestNewLinkChecker(t *testing.T) {
	log := logger.Get("debug")
	lc := NewLinkChecker(log)
	assert.NotNil(t, lc)
}

func setupBaseURL() (*url.URL, core.LinkChecker) {
	log := logger.Get("debug")
	lc := NewLinkChecker(log)
	baseURL, _ := url.Parse("http://base.com")
	return baseURL, lc
}

func TestCheckLink(t *testing.T) {
	baseURL, lc := setupBaseURL()
	client := http.DefaultClient

	t.Run("Invalid link patterns", func(t *testing.T) {
		assert.Nil(t, lc.CheckLink(context.Background(), client, "", baseURL))
		assert.Nil(t, lc.CheckLink(context.Background(), client, "#", baseURL))
		assert.Nil(t, lc.CheckLink(context.Background(), client, "mailto:test@test.com", baseURL))
		assert.Nil(t, lc.CheckLink(context.Background(), client, "javascript:void(0)", baseURL))
	})

	t.Run("Relative link resolution", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/about", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		tsURL, _ := url.Parse(ts.URL)
		result := lc.CheckLink(context.Background(), client, "/about", tsURL)
		assert.NotNil(t, result)
		assert.Equal(t, ts.URL+"/about", result.URL)
		assert.True(t, result.IsAccessible)
	})

	t.Run("HEAD success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodHead, r.Method)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		result := lc.CheckLink(context.Background(), client, ts.URL, baseURL)
		assert.True(t, result.IsAccessible)
		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("HEAD fails, GET fallback success", func(t *testing.T) {
		callCount := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			assert.Equal(t, http.MethodGet, r.Method)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		result := lc.CheckLink(context.Background(), client, ts.URL, baseURL)
		assert.True(t, result.IsAccessible)
		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, 2, callCount)
	})

	t.Run("Both HEAD and GET fail", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		result := lc.CheckLink(context.Background(), client, ts.URL, baseURL)
		assert.False(t, result.IsAccessible)
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
	})

	t.Run("Invalid parsing", func(t *testing.T) {
		result := lc.CheckLink(context.Background(), client, "http://[fe80::%31]/", baseURL)
		assert.Nil(t, result)
	})

	t.Run("Both HEAD and GET fail with network error", func(t *testing.T) {
		result := lc.CheckLink(context.Background(), client, "http://localhost:1", baseURL)
		assert.False(t, result.IsAccessible)
		assert.Equal(t, 0, result.StatusCode)
	})

	t.Run("Redirect handling", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/redirect" {
				http.Redirect(w, r, "/target", http.StatusFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		tsURL, _ := url.Parse(ts.URL)

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		result := lc.CheckLink(context.Background(), client, "/redirect", tsURL)
		assert.True(t, result.IsAccessible)
		assert.Equal(t, http.StatusFound, result.StatusCode)
	})

	t.Run("HEAD Request creation error", func(t *testing.T) {
		result := lc.CheckLink(context.Background(), client, "http://invalid.com:abc", baseURL)
		assert.Nil(t, result)
	})
}

func TestRunWorker(t *testing.T) {
	baseURL, lc := setupBaseURL()

	t.Run("Process links and cancel", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		linksChan := make(chan string, 5)
		resultsChan := make(chan model.LinkCheckResult, 5)
		var wg sync.WaitGroup

		linksChan <- ts.URL
		linksChan <- ts.URL

		ctx, cancel := context.WithCancel(context.Background())

		wg.Add(1)
		go lc.RunWorker(ctx, linksChan, resultsChan, baseURL, &wg)

		<-resultsChan

		cancel()

		wg.Wait()
	})

	t.Run("Results channel block and cancel", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		linksChan := make(chan string, 1)
		resultsChan := make(chan model.LinkCheckResult) // No buffer
		var wg sync.WaitGroup

		linksChan <- ts.URL

		wg.Add(1)
		go lc.RunWorker(ctx, linksChan, resultsChan, baseURL, &wg)

		time.Sleep(100 * time.Millisecond)
		cancel()

		wg.Wait()
	})
}
