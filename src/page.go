package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Page struct {
	title  string
	link   string
	in     []*Page
	out    []*Page
	x      float64
	y      float64
	weight int
	level  int
}

var pages = make(map[string]*Page)

func getOutPages(links map[string]int, depth int) []*Page {
	outPages := make([]*Page, 0)
	fmt.Printf("Processing %d links in getOutPages\n", len(links))

	for link := range links {
		fmt.Printf("Processing out link: %s\n", link)
		if p, ok := pages[link]; ok {
			fmt.Printf("Found existing page for %s\n", link)
			outPages = append(outPages, p)
			continue
		}

		fmt.Printf("Creating new page for %s\n", link)
		page := GetPage(link, depth+1)
		fmt.Printf("Created page with link: %s, out length: %d\n", page.link, len(page.out))
		pages[link] = page
		outPages = append(outPages, page)
	}
	fmt.Printf("Returning %d out pages\n", len(outPages))
	return outPages
}

func GetAllInPages() {
	for _, page := range pages {
		for _, outPage := range page.out {
			if ref, ok := pages[outPage.link]; ok {
				ref.in = append(ref.in, page)
			}
		}
	}
}

func CalculatePageWeights() {
	// Calculate Page weights based on the number of incoming links
	for _, page := range pages {
		page.weight = (len(page.in) + len(page.out)) * 2
	}
}

func GetPage(url string, depth int) *Page {
	fmt.Printf("GetPage called for %s at depth %d\n", url, depth)

	// If already visited, return immediately
	if existing, ok := pages[url]; ok {
		return existing
	}

	// Insert placeholder to prevent recursive visits
	stub := &Page{link: url}
	pages[url] = stub

	if depth > maxDepth {
		fmt.Printf("Max depth reached, returning stub for %s\n", url)
		return stub
	}

	client := http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		log.Printf("Error loading Page %s: %s\n", url, err.Error())
		return stub
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Page %s: %s\n", url, err.Error())
		return stub
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Skipping %s due to HTTP status %d\n", url, resp.StatusCode)
		return stub
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		log.Printf("Skipping non-HTML content: %s (%s)\n", url, contentType)
		return stub
	}

	pageInfo := ParsePage(content)
	fmt.Printf("Parsed page %s, found %d links\n", url, len(pageInfo.links))

	// Fill in real values
	stub.title = pageInfo.title
	stub.out = getOutPages(pageInfo.links, depth)
	stub.weight = 1

	return stub
}

func PrintPages() {
	for _, page := range pages {
		log.Printf("Page %s\n", page.link)
		log.Printf("Title: %s\n", page.title)
		log.Printf("In len: %d\n", len(page.in))
		log.Printf("Out len: %d\n", len(page.out))
		log.Printf("Weight: %d\n", page.weight)
	}
}
