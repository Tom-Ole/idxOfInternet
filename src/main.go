package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

func main() {
	startUrl := "https://go.dev/blog/error-handling-and-go"
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

	// TODO Handle backend: ....
	LayoutPages()

	log.Println("Starting server on http://localhost:8080")
	http.HandleFunc("/nodes", handleNodes)
	http.HandleFunc("/edges", handleEdges)
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

// LayoutPages assigns simple X/Y positions (placeholder for real layout)
func LayoutPages() {
	for _, p := range pages {
		p.x = rand.Float64()*1000 - 500
		p.y = rand.Float64()*1000 - 500
	}
}
