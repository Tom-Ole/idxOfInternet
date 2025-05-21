package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Node struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Weight int     `json:"weight"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

const loadFromFile = false
const maxDepth = 1

func main() {
	if !loadFromFile {
		startUrl := "https://de.wikipedia.org/wiki/Sex"
		root := GetPage(startUrl, 0)
		pages[startUrl] = root
		GetAllInPages()
		CalculatePageWeights()
		fmt.Print("==========================\n")
		fmt.Print("==========================\n")
		PrintPages()
		fmt.Print("==========================\n")
		fmt.Printf("Page length: %d \n", len(pages))
		fmt.Print("==========================\n")

		fmt.Print("==========================\n")
		fmt.Printf("Preprocess Pages %d \n", len(pages))

		cleaned := PreprocessPages(pages)
		pages = cleaned
		fmt.Print("==========================\n")

		SavePagesToFile("pages.json", pages)
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

		pages = pagesFromFile
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
	nodes := make([]Node, 0, len(pages))
	for _, p := range pages {
		nodes = append(nodes, Node{
			ID:     p.link,
			Title:  p.title,
			Weight: p.weight,
			X:      p.x,
			Y:      p.y,
		})
	}
	writeJSON(w, nodes)
}

func handleEdges(w http.ResponseWriter, r *http.Request) {
	edges := make([]Edge, 0)
	for _, p := range pages {
		for _, out := range p.out {
			edges = append(edges, Edge{From: p.link, To: out.link})
		}
	}
	writeJSON(w, edges)
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func LayoutPages() {
	AssignCoordinatesWeighted(pages)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Erlaube alle Ursprünge – für Produktion kannst du das gezielt setzen!
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// OPTIONS-Anfrage sofort beantworten (für Preflight)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Weiterreichen an nächsten Handler
		next.ServeHTTP(w, r)
	})
}
