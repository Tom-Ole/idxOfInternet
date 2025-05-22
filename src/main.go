package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
)

type Node struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Weight int     `json:"weight"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

const loadFromFile = true
const maxDepth = 3
const layoutWidth float32 = 10000.0 * 10 // Increased canvas size for large graphs
const layoutHeight float32 = 10000.0 * 8

func main() {
	// Set GOMAXPROCS to use all available CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	if !loadFromFile {
		startUrl := "https://de.wikipedia.org/wiki/Sex"
		fmt.Print("==========================\n")
		fmt.Printf("Starting page processing from: %s\n", startUrl)
		fmt.Print("==========================\n")

		root := GetPage(startUrl, 0)
		pages[startUrl] = root

		fmt.Print("==========================\n")
		fmt.Println("Waiting for all pages to be processed...")
		WaitForPages()
		fmt.Printf("All pages processed. Total pages: %d\n", len(pages))
		fmt.Print("==========================\n")

		GetAllInPages()
		CalculatePageWeights()
		PrintPages()
		fmt.Print("==========================\n")
		fmt.Printf("Page length: %d \n", len(pages))
		fmt.Print("==========================\n")

		fmt.Print("==========================\n")
		fmt.Printf("Preprocess Pages %d \n", len(pages))

		// Convert pages map to slice for preprocessing
		pagesMap := make(map[string]*Page)
		for url, idx := range pages {
			pagesMap[url] = &pageList[idx]
		}

		cleaned := PreprocessPages(pagesMap)

		// Convert back to map
		pages = make(map[string]uint32)
		pageList = make([]Page, 0, len(cleaned))
		for url, p := range cleaned {
			idx := uint32(len(pageList))
			pageList = append(pageList, *p)
			pages[url] = idx
		}
		fmt.Print("==========================\n")

		// Save pages to file
		pagesToSave := make(map[string]*Page)
		for url, idx := range pages {
			pagesToSave[url] = &pageList[idx]
		}
		SavePagesToFile("pages.json", pagesToSave)
		fmt.Print("==========================\n")

	} else {
		fmt.Print("==========================\n")
		fmt.Printf("Read Page from file: %s \n", "pages.json")
		fmt.Print("==========================\n")

		pagesFromFile, err := ReadPagesFromFile("pages.json")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Loaded %d pages from file\n", len(pagesFromFile))
		fmt.Print("==========================\n")

		// Convert pages to map
		pages = make(map[string]uint32)
		pageList = make([]Page, 0, len(pagesFromFile))
		for url, p := range pagesFromFile {
			idx := uint32(len(pageList))
			pageList = append(pageList, *p)
			pages[url] = idx
		}
	}

	fmt.Print("==========================\n")
	fmt.Printf("Create Layout %d \n", len(pages))
	fmt.Print("==========================\n")

	LayoutPages()

	fmt.Printf("Layout created")

	log.Println("Starting server on http://localhost:8080")
	http.Handle("/nodes", enableCORS(http.HandlerFunc(handleNodes)))
	http.Handle("/edges", enableCORS(http.HandlerFunc(handleEdges)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleNodes(w http.ResponseWriter, r *http.Request) {
	// Convert pages map to slice for efficient iteration
	pagesSlice := make([]*Page, 0, len(pages))
	for _, idx := range pages {
		pagesSlice = append(pagesSlice, &pageList[idx])
	}

	// Process nodes in parallel chunks
	numCPU := runtime.NumCPU()
	chunkSize := len(pagesSlice) / numCPU
	nodes := make([]Node, len(pagesSlice))
	var wg sync.WaitGroup

	for i := 0; i < numCPU; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numCPU-1 {
			end = len(pagesSlice)
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				p := pagesSlice[j]
				nodes[j] = Node{
					ID:     p.link,
					Title:  p.title,
					Weight: int(p.weight),
					X:      p.x,
					Y:      p.y,
				}
			}
		}(start, end)
	}
	wg.Wait()

	writeJSON(w, nodes)
}

func handleEdges(w http.ResponseWriter, r *http.Request) {
	// Convert pages map to slice for efficient iteration
	pagesSlice := make([]*Page, 0, len(pages))
	for _, idx := range pages {
		pagesSlice = append(pagesSlice, &pageList[idx])
	}

	// Process edges in parallel chunks
	numCPU := runtime.NumCPU()
	chunkSize := len(pagesSlice) / numCPU
	edgeChunks := make([][]Edge, numCPU)
	var wg sync.WaitGroup

	for i := 0; i < numCPU; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numCPU-1 {
			end = len(pagesSlice)
		}

		wg.Add(1)
		go func(start, end, chunkIdx int) {
			defer wg.Done()
			edges := make([]Edge, 0)
			for j := start; j < end; j++ {
				p := pagesSlice[j]
				for _, outIdx := range p.out {
					if outIdx < uint32(len(pagesSlice)) {
						edges = append(edges, Edge{
							From: p.link,
							To:   pagesSlice[outIdx].link,
						})
					}
				}
			}
			edgeChunks[chunkIdx] = edges
		}(start, end, i)
	}
	wg.Wait()

	// Combine edge chunks
	var allEdges []Edge
	for _, chunk := range edgeChunks {
		allEdges = append(allEdges, chunk...)
	}

	writeJSON(w, allEdges)
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func LayoutPages() {
	// Convert pages map to slice for efficient layout
	pagesSlice := make([]*Page, 0, len(pages))
	for _, idx := range pages {
		pagesSlice = append(pagesSlice, &pageList[idx])
	}

	// First use hierarchical layout for initial positioning
	AssignCoordinatesWeighted(pagesSlice)

	// Then apply force-directed layout for fine-tuning
	OptimizedFruchtermanReingold(pagesSlice, layoutWidth, layoutHeight, 100)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
