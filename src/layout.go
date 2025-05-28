package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
)

func randFloat64() float64 {
	// return a randome float64 between -1 and 1
	return (rand.Float64() * 2) - 1
}

// CreateLayout creates a force-directed layout optimized for 250k+ nodes
func (g *Graph) CreateLayout() {

	fmt.Printf("Creating layout for %d nodes and %d clusters...\n", len(g.Nodes), len(g.Clusters))

	// Calculate the radius and center of the cluster
	for _, cluster := range g.Clusters {
		cluster.Radius = 0
		cluster.CenterX = 1
		cluster.CenterY = 1
		for _, nodeID := range cluster.NodeIDs {
			cluster.Radius += float64(g.Nodes[nodeID].Weight)
			node := g.Nodes[nodeID]
			cluster.CenterX += node.X * randFloat64()
			cluster.CenterY += node.Y * randFloat64()

		}
		cluster.CenterX /= float64(len(cluster.NodeIDs))
		cluster.CenterY /= float64(len(cluster.NodeIDs))
	}

	// Layout all clusters to each other with a force-directed layout
	const (
		padding       = 75
		iterations    = 100
		forceStrength = 0.1
	)

	for range iterations {
		for idA, clusterA := range g.Clusters {
			for idB, clusterB := range g.Clusters {
				if idA >= idB {
					continue // avoid duplicate pairs and self-comparison
				}
				dx := clusterB.CenterX - clusterA.CenterX
				dy := clusterB.CenterY - clusterA.CenterY
				dist := math.Hypot(dx, dy)
				minDist := float64(clusterA.Radius + clusterB.Radius + padding)

				if dist < minDist && dist > 0.001 {
					nx := dx / dist
					ny := dy / dist
					force := (minDist - dist) * forceStrength
					moveX := nx * force
					moveY := ny * force

					clusterA.CenterX -= moveX
					clusterA.CenterY -= moveY
					clusterB.CenterX += moveX
					clusterB.CenterY += moveY
				}
			}
		}
	}

	// Layout all nodes inside the clusters so all nodes spread out evenly inside the radius of the cluster not just around
	const (
		nodeIterations = 50
		repelStrength  = 0.5
		boundStrength  = 0.1
	)

	var wg sync.WaitGroup

	for _, cluster := range g.Clusters {

		if len(cluster.NodeIDs) <= 1 {
			continue
		}

		wg.Add(1)
		go func(cluster *Cluster) {
			defer wg.Done()

			nodes := make([]*Node, len(cluster.NodeIDs))
			radii := make([]float64, len(cluster.NodeIDs))
			for i, id := range cluster.NodeIDs {
				nodes[i] = g.Nodes[id]
				radii[i] = math.Sqrt(float64(nodes[i].Weight))*2 + 1

				// Initial position: random within cluster radius
				angle := randFloat64() * math.Pi * 2
				dist := randFloat64() * cluster.Radius * 0.8 // stay inside
				nodes[i].X = cluster.CenterX + dist*math.Cos(angle)
				nodes[i].Y = cluster.CenterY + dist*math.Sin(angle)
			}

			// Mini force-directed simulation
			for iter := 0; iter < nodeIterations; iter++ {
				for i := 0; i < len(nodes); i++ {
					ni := nodes[i]
					r1 := radii[i]
					fx, fy := 0.0, 0.0

					// Repel from other nodes
					for j := 0; j < len(nodes); j++ {
						if i == j {
							continue
						}
						nj := nodes[j]
						r2 := radii[j]

						dx := ni.X - nj.X
						dy := ni.Y - nj.Y
						dist := math.Hypot(dx, dy)
						minDist := r1 + r2

						if dist < minDist && dist > 0.01 {
							// Normalize
							nx := dx / dist
							ny := dy / dist
							force := (minDist - dist) * repelStrength

							fx += nx * force
							fy += ny * force
						}
					}

					// Pull back inside the cluster circle
					cdx := ni.X - cluster.CenterX
					cdy := ni.Y - cluster.CenterY
					cd := math.Hypot(cdx, cdy)
					maxDist := cluster.Radius - r1

					if cd > maxDist && cd > 0.01 {
						// Outside boundary, pull back in
						nx := cdx / cd
						ny := cdy / cd
						fx -= nx * (cd - maxDist) * boundStrength
						fy -= ny * (cd - maxDist) * boundStrength
					}

					// Apply forces
					ni.X += fx
					ni.Y += fy
				}
			}
		}(cluster)
	}

	wg.Wait()

	//g.PrintClusters()

}
