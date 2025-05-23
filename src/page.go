package main

import "fmt"

func ParsePage(link string, depth int, graph *Graph) {
	if depth == 0 {
		return
	}

	fmt.Printf("Depth: %d -- Parsing page: %s\n", depth, link)

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
