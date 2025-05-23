package main

import "fmt"

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
