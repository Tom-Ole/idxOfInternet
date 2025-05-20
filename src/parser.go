package main

import (
	"fmt"
	"regexp"
)

type PageContent struct {
	links map[string]int
	title string
}

// TODO preprocess url to have consistent format like https://www.example.com and https://example.com are the same
func getLinks(strContent string) map[string]int {
	// Schema: <a ... > ... </a>
	// Regex: <a\s+(?:[^>]*?\s+)?href="([^"]*)"
	regex := regexp.MustCompile(`<a\s+(?:[^>]*?\s+)?href="([^"]*)"`)
	matches := regex.FindAllStringSubmatch(strContent, -1)
	links := make(map[string]int)

	fmt.Printf("Found %d total link matches\n", len(matches))

	for _, match := range matches {
		tmpLink := match[1]
		if IsValidURL(tmpLink) {
			fmt.Printf("Valid URL found: %s\n", tmpLink)
			if _, ok := links[tmpLink]; !ok {
				links[tmpLink] = 1
			} else {
				links[tmpLink]++
			}
		}
	}

	fmt.Printf("Total valid links found: %d\n", len(links))
	return links
}

func getTitle(strContent string) string {
	// Schema: <title> ... </title>
	// Regex: <title>([^<]*)</title>
	regex := regexp.MustCompile(`<title>([^<]*)</title>`)
	matches := regex.FindAllStringSubmatch(strContent, -1)

	if len(matches) > 0 {
		return matches[0][1]
	}

	return ""

}

func ParsePage(content []byte) PageContent {

	strContent := string(content)

	links := getLinks(strContent)
	title := getTitle(strContent)

	return PageContent{
		links: links,
		title: title,
	}
}
