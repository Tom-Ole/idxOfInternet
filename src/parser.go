package main

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func GetLinks(link string) ([]string, error) {
	httpClient := &http.Client{Timeout: 5 * time.Second}

	resp, err := httpClient.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	// links pattern: "<a href=\"(.*?)\">"
	// extract links from the response body

	pattern := `<a[^>]*href="([^"]+)"`
	linksMap := make(map[string]struct{})

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := FindAllMatches(string(bodyBytes), pattern)
	for _, match := range matches {
		linksMap[match[1]] = struct{}{}
	}

	var links []string
	for k := range linksMap {
		links = append(links, k)
	}

	// Preprocess the links
	processedLinks, err := PreprocessLinks(links, link)
	if err != nil {
		return nil, err
	}

	PrintInfo("Found %d valid links for %s\n", len(processedLinks), link)

	return processedLinks, nil

}

func FindAllMatches(body string, pattern string) [][]string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(body, -1)
	return matches
}

func ExtractDomain(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}

	return u.Hostname()
}

func GetLinksV2(link string) ([]string, error) {
	fastClient := &http.Client{Timeout: 4 * time.Second}
	resp, err := fastClient.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	base, _ := url.Parse(link)

	links := make(map[string]struct{})
	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				break
			}
			return nil, z.Err()
		}

		if tt != html.StartTagToken {
			continue
		}

		t := z.Token()
		if t.Data != "a" {
			continue
		}

		for i := 0; i < len(t.Attr); i++ {
			a := t.Attr[i]
			if a.Key != "href" {
				continue
			}

			href := strings.TrimSpace(a.Val)
			if href == "" || href[0] == '#' || strings.HasPrefix(href, "javascript:") {
				continue
			}

			absURL, err := base.Parse(href)
			if err != nil || absURL.Scheme == "" || absURL.Host == "" {
				continue
			}

			links[absURL.String()] = struct{}{}
		}
	}

	result := make([]string, 0, len(links))
	for l := range links {
		result = append(result, l)
	}
	return result, nil
}
