package htmlhelper

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func GetHTMLVersion(doc *html.Node) string {
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

func GetHeadingsCount(doc *html.Node) map[string]int {
	headings := map[string]int{
		"h1": 0,
		"h2": 0,
		"h3": 0,
		"h4": 0,
		"h5": 0,
		"h6": 0,
	}

	nodeList := []*html.Node{doc}

	for len(nodeList) > 0 {
		n := nodeList[len(nodeList)-1]
		nodeList = nodeList[:len(nodeList)-1]

		if n.Type == html.ElementNode {
			if _, ok := headings[n.Data]; ok {
				headings[n.Data]++
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			nodeList = append(nodeList, c)
		}
	}

	return headings
}

func IsInternalLink(link string, baseURL *url.URL) bool {
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		return true
	}

	parsedLink, err := url.Parse(link)
	if err != nil {
		return false
	}
	return parsedLink.Host == baseURL.Host
}

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func GetLinks(doc *html.Node) []string {
	var links []string
	latestNodes := []*html.Node{doc}

	for len(latestNodes) > 0 {
		n := latestNodes[len(latestNodes)-1]
		latestNodes = latestNodes[:len(latestNodes)-1]

		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
					break
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			latestNodes = append(latestNodes, c)
		}
	}

	return links
}

func GetTitle(doc *html.Node) string {
	if isTitleElement(doc) {
		return doc.FirstChild.Data
	}

	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		result := GetTitle(c)
		if result != "" {
			return result
		}
	}

	return ""
}

func HasLoginForm(doc *html.Node) bool {
	hasPassword := false
	hasUserField := false

	stack := []*html.Node{doc}

	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if n.Type == html.ElementNode && n.Data == "input" {
			t := getAttr(n, "type")
			name := getAttr(n, "name")

			switch t {
			case "password":
				hasPassword = true

			case "text", "email":
				hasUserField = true
			}

			if strings.Contains(name, "user") ||
				strings.Contains(name, "email") ||
				strings.Contains(name, "login") {
				hasUserField = true
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			stack = append(stack, c)
		}
	}

	return hasPassword && hasUserField
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return strings.ToLower(a.Val)
		}
	}
	return ""
}
