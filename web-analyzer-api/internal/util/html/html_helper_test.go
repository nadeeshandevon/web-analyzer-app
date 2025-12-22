package htmlhelper

import (
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestGetTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Title in Head",
			html:     `<html><head><title>Title in Head</title></head><body><h1>Body</h1></body></html>`,
			expected: "Title in Head",
		},
		{
			name:     "Title in Body (Should be ignored)",
			html:     `<html><head></head><body><title>Title in Body</title><h1>Body</h1></body></html>`,
			expected: "Title in Body",
		},
		{
			name:     "Multiple Titles",
			html:     `<html><head><title>First Title</title><title>Second Title</title></head><body></body></html>`,
			expected: "First Title",
		},
		{
			name:     "No Title",
			html:     `<html><head></head><body><h1>Body</h1></body></html>`,
			expected: "",
		},
		{
			name:     "Title with Attributes",
			html:     `<html><head><title id="main-title">Attr Title</title></head><body></body></html>`,
			expected: "Attr Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			if got := GetTitle(doc); got != tt.expected {
				t.Errorf("GetTitle() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "HTML5",
			html:     `<!DOCTYPE html><html><body></body></html>`,
			expected: "HTML5",
		},
		{
			name:     "HTML 4.01 Strict",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd"><html><body></body></html>`,
			expected: "HTML 4.01",
		},
		{
			name:     "XHTML 1.0 Strict",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"><html><body></body></html>`,
			expected: "XHTML",
		},
		{
			name:     "Unknown",
			html:     `<html><body></body></html>`,
			expected: "Unknown",
		},
		{
			name:     "InvalidHtml",
			html:     `<div><body></body></div>`,
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			if got := GetHTMLVersion(doc); got != tt.expected {
				t.Errorf("GetHTMLVersion() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetHeadingsCount(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected map[string]int
	}{
		{
			name: "All heading levels",
			html: `<h1>H1</h1><h2>H2</h2><h3>H3</h3><h4>H4</h4><h5>H5</h5><h6>H6</h6>`,
			expected: map[string]int{
				"h1": 1, "h2": 1, "h3": 1, "h4": 1, "h5": 1, "h6": 1,
			},
		},
		{
			name: "Multiple heading level html",
			html: `<h1>H1-1</h1><h1>H1-2</h1><h2>H2</h2>`,
			expected: map[string]int{
				"h1": 2, "h2": 1, "h3": 0, "h4": 0, "h5": 0, "h6": 0,
			},
		},
		{
			name: "Nested headings html",
			html: `<h1>H1<div><h2>H2</h2></div></h1>`,
			expected: map[string]int{
				"h1": 1, "h2": 1, "h3": 0, "h4": 0, "h5": 0, "h6": 0,
			},
		},
		{
			name: "No headings",
			html: `<div>Sample description</div>`,
			expected: map[string]int{
				"h1": 0, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			got := GetHeadingsCount(doc)
			for tag, count := range tt.expected {
				if got[tag] != count {
					t.Errorf("GetHeadingsCount() for %s = %v, want %v", tag, got[tag], count)
				}
			}
		})
	}
}

func TestGetLinks(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name:     "Simple links html",
			html:     `<a href="http://myapp.test/1">L1</a><a href="/relative">L2</a>`,
			expected: []string{"/relative", "http://myapp.test/1"},
		},
		{
			name:     "Mixed elements html",
			html:     `<div><a href="http://myapp.test/1">L1</a><span>Text</span><a href="http://myapp.test/2">L2</a></div>`,
			expected: []string{"http://myapp.test/2", "http://myapp.test/1"},
		},
		{
			name:     "Link without href",
			html:     `<a>No href</a><a href="http://myapp.test/3">L3</a>`,
			expected: []string{"http://myapp.test/3"},
		},
		{
			name:     "No links html",
			html:     `<p>Empty links</p>`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			got := GetLinks(doc)
			if len(got) != len(tt.expected) {
				t.Errorf("GetLinks() returned %d links, want %d", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("GetLinks()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestIsInternalLink(t *testing.T) {
	baseURL, _ := url.Parse("http://myapp.test")

	tests := []struct {
		name     string
		link     string
		expected bool
	}{
		{
			name:     "Relative path",
			link:     "/about",
			expected: true,
		},
		{
			name:     "Same host",
			link:     "http://myapp.test/contact",
			expected: true,
		},
		{
			name:     "Different host",
			link:     "http://external-site.test",
			expected: false,
		},
		{
			name:     "Empty link",
			link:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInternalLink(tt.link, baseURL); got != tt.expected {
				t.Errorf("IsInternalLink(%v) = %v, want %v", tt.link, got, tt.expected)
			}
		})
	}
}

func TestHasLoginForm(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name:     "Valid login form",
			html:     `<form><input type="text" name="username"><input type="password" name="pwd"></form>`,
			expected: true,
		},
		{
			name:     "Email and password",
			html:     `<form><input type="email" name="user_email"><input type="password"></form>`,
			expected: true,
		},
		{
			name:     "Only password",
			html:     `<form><input type="password"></form>`,
			expected: false,
		},
		{
			name:     "Only name",
			html:     `<form><input type="text" name="login"></form>`,
			expected: false,
		},
		{
			name:     "No form",
			html:     `<div>Just a page</div>`,
			expected: false,
		},
		{
			name:     "Input outside form",
			html:     `<div><input type="text" name="username"><input type="password"></div>`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			if got := HasLoginForm(doc); got != tt.expected {
				t.Errorf("HasLoginForm() = %v, want %v (html: %s)", got, tt.expected, tt.html)
			}
		})
	}
}
