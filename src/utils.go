package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func (g *Graph) PrintGraph() {
	for _, node := range g.Nodes {
		fmt.Printf("Node %d (x: %f, y: %f, w: %d): %s\n", node.ID, node.X, node.Y, node.Weight, node.Link)
		for _, edgeID := range node.EdgeIDs {
			edge := g.Edges[edgeID]
			fmt.Printf("  Edge to Node %d\n", edge.To)
		}
	}
}

func (g *Graph) PrintClusters() {
	for _, cluster := range g.Clusters {
		fmt.Printf("Cluster %d (x: %f, y: %f, r: %d): ", cluster.ID, cluster.CenterX, cluster.CenterY, cluster.Radius)
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

func SaveGraphToFile(g *Graph, filename string) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write graph to file: %w", err)
	}

	fmt.Printf("Graph saved to file %s\n", filename)

	return nil
}

func LoadGraphFromFile(filename string) (*Graph, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read graph from file: %w", err)
	}
	var g Graph
	err = json.Unmarshal(data, &g)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal graph: %w", err)
	}

	fmt.Printf("Loaded graph from file: %s\n", filename)

	return &g, nil
}

func PrintInfo(format string, a ...any) {
	shouldPrintInfo := false
	if shouldPrintInfo {
		fmt.Printf(format, a...)
	}
}
