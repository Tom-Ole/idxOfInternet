package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Page represents a node in the graph with optimized memory usage
type Page struct {
	title  string
	link   string
	in     []uint32 // Store indices instead of pointers
	out    []uint32 // Store indices instead of pointers
	x      float32  // Use float32 instead of float64 for memory
	y      float32  // Use float32 instead of float64 for memory
	weight uint16   // Use uint16 instead of int for memory
	level  uint8    // Use uint8 instead of int for memory
}

// Global state with mutex protection
var (
	pages     = make(map[string]uint32)  // Map URL to index
	pageList  = make([]Page, 0, 1000000) // Pre-allocate for 1M nodes
	pageMutex sync.RWMutex
	pageWg    sync.WaitGroup // WaitGroup to track page processing

	// Global HTTP client with improved connection pooling
	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	// Rate limiter: 20 requests per second with burst of 10
	limiter = rate.NewLimiter(rate.Limit(20), 10)
)

// getOutPages returns indices of outbound pages
func getOutPages(links map[string]int, depth int) []uint32 {
	outPages := make([]uint32, 0, len(links))

	pageMutex.RLock()
	defer pageMutex.RUnlock()

	for link := range links {
		if idx, ok := pages[link]; ok {
			outPages = append(outPages, idx)
			continue
		}

		// Create new page
		idx := uint32(len(pageList))
		pageList = append(pageList, Page{link: link})
		pages[link] = idx
		outPages = append(outPages, idx)

		// Process new page in background
		pageWg.Add(1)
		go func(link string, idx uint32) {
			defer pageWg.Done()
			processPage(link, idx, depth+1)
		}(link, idx)
	}

	return outPages
}

// processPage handles the actual page processing
func processPage(url string, idx uint32, depth int) {
	if depth > maxDepth {
		return
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Increased timeout
	defer cancel()

	// Wait for rate limiter with context
	if err := limiter.Wait(ctx); err != nil {
		log.Printf("Rate limiter error for %s: %s\n", url, err.Error())
		return
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Error creating request for %s: %s\n", url, err.Error())
		return
	}

	// Add headers to prevent blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Error loading Page %s: %s\n", url, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Skipping page %s: Status code %d\n", url, resp.StatusCode)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		log.Printf("Skipping page %s: Not HTML content\n", url)
		return
	}

	// Read body with timeout
	content, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Limit to 1MB
	if err != nil {
		log.Printf("Error reading content from %s: %s\n", url, err.Error())
		return
	}

	pageInfo := ParsePage(content)

	pageMutex.Lock()
	defer pageMutex.Unlock()

	page := &pageList[idx]
	page.title = pageInfo.title

	// Process links in smaller batches
	const batchSize = 3 // Reduced batch size
	links := make([]string, 0, len(pageInfo.links))
	for link := range pageInfo.links {
		links = append(links, link)
	}

	// Process links in smaller batches
	for i := 0; i < len(links); i += batchSize {
		end := i + batchSize
		if end > len(links) {
			end = len(links)
		}

		batch := links[i:end]
		outPages := make([]uint32, 0, len(batch))

		for _, link := range batch {
			if idx, ok := pages[link]; ok {
				outPages = append(outPages, idx)
				continue
			}

			// Create new page
			newIdx := uint32(len(pageList))
			pageList = append(pageList, Page{link: link})
			pages[link] = newIdx
			outPages = append(outPages, newIdx)

			// Process new page in background
			pageWg.Add(1)
			go func(link string, idx uint32) {
				defer pageWg.Done()
				processPage(link, idx, depth+1)
			}(link, newIdx)
		}

		page.out = append(page.out, outPages...)
	}

	page.weight = 1
	log.Printf("Processed page %s (depth %d): found %d links\n", url, depth, len(pageInfo.links))
}

// GetAllInPages processes incoming links
func GetAllInPages() {
	fmt.Print("==========================\n")
	fmt.Printf("GetAllInPages %d\n", len(pageList))
	fmt.Print("==========================\n")
	pageMutex.RLock()
	defer pageMutex.RUnlock()

	for i, page := range pageList {
		for _, outIdx := range page.out {
			if outIdx < uint32(len(pageList)) {
				pageList[outIdx].in = append(pageList[outIdx].in, uint32(i))
			}
		}
	}
}

// CalculatePageWeights updates page weights
func CalculatePageWeights() {
	fmt.Print("==========================\n")
	fmt.Printf("CalculatePageWeights %d\n", len(pageList))
	fmt.Print("==========================\n")
	pageMutex.RLock()
	defer pageMutex.RUnlock()

	for i := range pageList {
		weight := len(pageList[i].in) + len(pageList[i].out)
		if weight > 65535 {
			weight = 65535
		}
		pageList[i].weight = uint16(weight * 2)
	}
}

// GetPage returns a page by URL
func GetPage(url string, depth int) uint32 {
	pageMutex.RLock()
	if idx, ok := pages[url]; ok {
		pageMutex.RUnlock()
		return idx
	}
	pageMutex.RUnlock()

	pageMutex.Lock()
	defer pageMutex.Unlock()

	// Double check after acquiring write lock
	if idx, ok := pages[url]; ok {
		return idx
	}

	idx := uint32(len(pageList))
	pageList = append(pageList, Page{link: url})
	pages[url] = idx

	pageWg.Add(1)
	go func() {
		defer pageWg.Done()
		processPage(url, idx, depth)
	}()
	return idx
}

// WaitForPages waits for all page processing to complete
func WaitForPages() {
	fmt.Printf("Waiting for %d pages to be processed...\n", len(pageList))
	pageWg.Wait()
	fmt.Printf("All pages processed successfully\n")
}

// PrintPages prints page information
func PrintPages() {
	pageMutex.RLock()
	defer pageMutex.RUnlock()

	for i, page := range pageList {
		log.Printf("Page %s\n", page.link)
		log.Printf("Title: %s\n", page.title)
		log.Printf("In len: %d\n", len(page.in))
		log.Printf("Out len: %d\n", len(page.out))
		log.Printf("Weight: %d\n", page.weight)
		log.Printf("Index: %d\n", i)
	}
}
