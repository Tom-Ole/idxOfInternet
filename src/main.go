package main

import "fmt"

type NodeID uint32
type EdgeID uint32

type Node struct {
	ID      NodeID
	Link    string
	X       int
	Y       int
	Weight  uint32
	EdgeIDs []EdgeID
}

type Edge struct {
	ID   EdgeID
	From NodeID
	To   NodeID
}

type Cluster struct {
	ID      int
	NodeIDs []NodeID
}

type Graph struct {
	Nodes    map[NodeID]*Node
	Edges    map[EdgeID]*Edge
	Clusters map[int]*Cluster

	nextNodeID NodeID
	nextEdgeID EdgeID
}

func (g *Graph) AddNode(link string, x int, y int, weight uint32) NodeID {
	id := g.nextNodeID
	g.nextNodeID++
	g.Nodes[id] = &Node{
		ID:     id,
		Link:   link,
		X:      x,
		Y:      y,
		Weight: weight,
	}

	return id
}

func (g *Graph) AddEdge(from NodeID, to NodeID) EdgeID {
	id := g.nextEdgeID
	g.nextEdgeID++
	g.Edges[id] = &Edge{
		ID:   id,
		From: from,
		To:   to,
	}

	g.Nodes[from].EdgeIDs = append(g.Nodes[from].EdgeIDs, id)

	return id
}

func (g *Graph) PrintGraph() {
	for _, node := range g.Nodes {
		fmt.Printf("Node %d: %s\n", node.ID, node.Link)
		for _, edgeID := range node.EdgeIDs {
			edge := g.Edges[edgeID]
			fmt.Printf("  Edge to Node %d\n", edge.To)
		}
	}
}

func createGraph() *Graph {
	return &Graph{
		Nodes:    make(map[NodeID]*Node),
		Edges:    make(map[EdgeID]*Edge),
		Clusters: make(map[int]*Cluster),
	}
}

func main() {

	initialLink := "https://go.dev/"
	depth := 2
	graph := createGraph()

	// recursive function to parse the initial Page and there links with given depth.

	ParsePage(initialLink, depth, graph)

	graph.PrintGraph()

}
