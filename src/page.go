package main

import (
	"fmt"
	"sync"
)

func ParsePage(link string, depth int, graph *Graph) {
	if depth == 0 {
		return
	}

	PrintInfo("Depth: %d -- Parsing page: %s\n", depth, link)

	links, err := GetLinks(link)
	if err != nil {
		fmt.Println("Error fetching links:", err)
		return
	}

	nodeID := graph.AddNode(link, 1, 1, 1)

	for _, l := range links {
		childNodeID := graph.AddNode(l, 1, 1, 1)
		graph.AddEdge(nodeID, childNodeID)
		ParsePage(l, depth-1, graph)
	}
}

func ParsePageConcurrently(link string, depth int, graph *Graph) {
	var wg sync.WaitGroup
	var visited sync.Map
	var mu sync.Mutex

	var crawl func(string, int)
	crawl = func(url string, d int) {
		defer wg.Done()

		if d == 0 {
			return
		}

		mu.Lock()
		if _, ok := visited.Load(url); ok {
			mu.Unlock()
			return
		}
		visited.Store(url, struct{}{})
		mu.Unlock()

		PrintInfo("Depth: %d -- Parsing: %s\n", d, url)

		links, err := GetLinksV2(url)
		if err != nil {
			return
		}

		fromID := graph.AddNode(url, 1, 1, 1)

		for _, l := range links {
			toID := graph.AddNode(l, 1, 1, 1)
			graph.AddEdge(fromID, toID)

			wg.Add(1)
			go crawl(l, d-1)
		}
	}

	wg.Add(1)
	go crawl(link, depth)
	wg.Wait()
}
