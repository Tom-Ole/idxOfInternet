package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type NodeID uint32
type EdgeID uint32

type Node struct {
	ID      NodeID   `json:"id"`
	Link    string   `json:"link"`
	X       int      `json:"x"`
	Y       int      `json:"y"`
	Weight  uint32   `json:"weight"`
	EdgeIDs []EdgeID `json:"edge_ids"`
}

type Edge struct {
	ID   EdgeID `json:"id"`
	From NodeID `json:"from"`
	To   NodeID `json:"to"`
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

func (g *Graph) ClusterByDomain() {
	domainClusters := map[string]int{}
	clusterID := 0

	for _, node := range g.Nodes {
		domain := ExtractDomain(node.Link)
		if domain == "" {
			continue
		}
		if _, exists := domainClusters[domain]; !exists {
			domainClusters[domain] = clusterID
			g.Clusters[clusterID] = &Cluster{ID: clusterID}
			clusterID++
		}
		cid := domainClusters[domain]
		g.Clusters[cid].NodeIDs = append(g.Clusters[cid].NodeIDs, node.ID)
	}
}

func (g *Graph) ClusterByConnectivity() {
	visited := map[NodeID]bool{}
	clusterID := 0

	var dfs func(NodeID)
	dfs = func(id NodeID) {
		visited[id] = true
		g.Clusters[clusterID].NodeIDs = append(g.Clusters[clusterID].NodeIDs, id)
		for _, edgeID := range g.Nodes[id].EdgeIDs {
			to := g.Edges[edgeID].To
			if !visited[to] {
				dfs(to)
			}
		}
	}

	for id := range g.Nodes {
		if !visited[id] {
			g.Clusters[clusterID] = &Cluster{ID: clusterID}
			dfs(id)
			clusterID++
		}
	}
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

func (g *Graph) PrintClusters() {
	for _, cluster := range g.Clusters {
		fmt.Printf("Cluster %d: ", cluster.ID)
		for _, nodeID := range cluster.NodeIDs {
			fmt.Printf("%d ", nodeID)
		}
		fmt.Println()
	}
}

func (g *Graph) Count() int {
	return len(g.Nodes)
}

func (g *Graph) CountClusters() int {
	return len(g.Clusters)
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

	// Create clusters based on domain or connectivity
	graph.ClusterByDomain()
	// graph.ClusterByConnectivity()

	fmt.Printf("Total nodes: %d\n", graph.Count())
	fmt.Printf("Total clusters: %d\n", graph.CountClusters())

	// Create layout to visualize the graph on a vanilla HTML/CSS/JS Frontend with DECK.gl
	graph.CreateLayout()

	// Open a server and serve the graph to the frontend
	http.HandleFunc("/graph", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Convert maps to slices
		nodes := make([]*Node, 0, len(graph.Nodes))
		for _, node := range graph.Nodes {
			nodes = append(nodes, node)
		}

		edges := make([]*Edge, 0, len(graph.Edges))
		for _, edge := range graph.Edges {
			edges = append(edges, edge)
		}

		// Encode as JSON
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"nodes": nodes,
			"edges": edges,
		})
		if err != nil {
			http.Error(w, "Failed to encode graph", http.StatusInternalServerError)
		}
	})
	http.ListenAndServe(":8080", nil)
	fmt.Println("Server started at :8080")

}
