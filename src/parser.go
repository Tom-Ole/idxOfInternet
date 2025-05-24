package main

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
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
