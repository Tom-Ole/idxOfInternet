package main

import (
	"math"
	"runtime"
	"sync"
)

// CreateLayout creates a force-directed layout optimized for 250k+ nodes
func (g *Graph) CreateLayout() {
	const (
		iterations     = 50         // Reduced for performance
		width, height  = 4000, 4000 // Larger space for more nodes
		repulsionConst = 100000.0
		centeringForce = 0.0005
		damping        = 0.8
		timestep       = 1.0
		gridSize       = 200   // Spatial partitioning grid
		maxInfluence   = 400.0 // Max distance for force calculation
	)

	nodeCount := len(g.Nodes)
	if nodeCount == 0 {
		return
	}

	// Initialize positions in a circle to avoid edge clustering
	nodeList := make([]*Node, 0, nodeCount)
	velocities := make([][2]float64, nodeCount)

	i := 0
	for _, node := range g.Nodes {
		// Distribute nodes in concentric circles
		layer := int(math.Sqrt(float64(i)))
		angle := float64(i-layer*layer) * 2.0 * math.Pi / float64(2*layer+1)
		if layer == 0 {
			angle = 0
		}
		radius := float64(layer)*30.0 + 50.0

		node.X = int(radius * math.Cos(angle))
		node.Y = int(radius * math.Sin(angle))
		nodeList = append(nodeList, node)
		velocities[i] = [2]float64{0, 0}
		i++
	}

	// Spatial grid for optimization
	type GridCell struct {
		nodes   []int // indices into nodeList
		centerX float64
		centerY float64
		mass    float64
	}

	gridCols := int(width / gridSize)
	gridRows := int(height / gridSize)

	// Determine number of workers based on CPU cores
	numWorkers := runtime.NumCPU()
	if numWorkers > 8 {
		numWorkers = 8 // Diminishing returns beyond 8 cores for this workload
	}

	for iter := 0; iter < iterations; iter++ {
		// Create spatial grid
		grid := make([][]GridCell, gridRows)
		for row := range grid {
			grid[row] = make([]GridCell, gridCols)
		}

		// Populate grid
		for idx, node := range nodeList {
			gridX := int((float64(node.X) + width/2) / gridSize)
			gridY := int((float64(node.Y) + height/2) / gridSize)

			// Clamp to grid bounds
			if gridX < 0 {
				gridX = 0
			}
			if gridX >= gridCols {
				gridX = gridCols - 1
			}
			if gridY < 0 {
				gridY = 0
			}
			if gridY >= gridRows {
				gridY = gridRows - 1
			}

			cell := &grid[gridY][gridX]
			cell.nodes = append(cell.nodes, idx)
			cell.centerX += float64(node.X)
			cell.centerY += float64(node.Y)
			cell.mass += 1.0
		}

		// Calculate cell centers
		for row := range grid {
			for col := range grid[row] {
				cell := &grid[row][col]
				if cell.mass > 0 {
					cell.centerX /= cell.mass
					cell.centerY /= cell.mass
				}
			}
		}

		// Calculate forces in parallel
		forces := make([][2]float64, nodeCount)
		var wg sync.WaitGroup

		chunkSize := nodeCount / numWorkers
		if chunkSize < 100 {
			chunkSize = 100
		}

		for start := 0; start < nodeCount; start += chunkSize {
			end := start + chunkSize
			if end > nodeCount {
				end = nodeCount
			}

			wg.Add(1)
			go func(startIdx, endIdx int) {
				defer wg.Done()

				for i := startIdx; i < endIdx; i++ {
					node := nodeList[i]
					x1 := float64(node.X)
					y1 := float64(node.Y)
					fx, fy := 0.0, 0.0

					// Get grid position
					gridX := int((x1 + width/2) / gridSize)
					gridY := int((y1 + height/2) / gridSize)

					// Check surrounding cells (3x3 neighborhood)
					for dy := -1; dy <= 1; dy++ {
						for dx := -1; dx <= 1; dx++ {
							checkX := gridX + dx
							checkY := gridY + dy

							if checkX < 0 || checkX >= gridCols || checkY < 0 || checkY >= gridRows {
								continue
							}

							cell := &grid[checkY][checkX]
							if len(cell.nodes) == 0 {
								continue
							}

							if dx == 0 && dy == 0 {
								// Same cell - calculate individual repulsions
								for _, otherIdx := range cell.nodes {
									if otherIdx == i {
										continue
									}

									other := nodeList[otherIdx]
									x2 := float64(other.X)
									y2 := float64(other.Y)
									dx := x1 - x2
									dy := y1 - y2
									distSq := dx*dx + dy*dy + 1.0

									if distSq > maxInfluence*maxInfluence {
										continue
									}

									force := repulsionConst / distSq
									dist := math.Sqrt(distSq)
									fx += (dx / dist) * force
									fy += (dy / dist) * force
								}
							} else {
								// Distant cell - use center of mass approximation
								if cell.mass > 0 {
									dx := x1 - cell.centerX
									dy := y1 - cell.centerY
									distSq := dx*dx + dy*dy + 1.0

									if distSq > maxInfluence*maxInfluence {
										continue
									}

									force := repulsionConst * cell.mass / distSq
									dist := math.Sqrt(distSq)
									fx += (dx / dist) * force
									fy += (dy / dist) * force
								}
							}
						}
					}

					// Centering force
					centerDist := math.Sqrt(x1*x1 + y1*y1)
					if centerDist > 1.0 {
						centerForceStrength := centeringForce * centerDist
						fx -= (x1 / centerDist) * centerForceStrength
						fy -= (y1 / centerDist) * centerForceStrength
					}

					forces[i] = [2]float64{fx, fy}
				}
			}(start, end)
		}

		wg.Wait()

		// Update positions
		for i, node := range nodeList {
			force := forces[i]
			vel := &velocities[i]

			// Update velocity with damping
			vel[0] = (vel[0] + force[0]*timestep) * damping
			vel[1] = (vel[1] + force[1]*timestep) * damping

			// Velocity limiting
			maxVel := 50.0
			if math.Abs(vel[0]) > maxVel {
				vel[0] = maxVel * math.Copysign(1, vel[0])
			}
			if math.Abs(vel[1]) > maxVel {
				vel[1] = maxVel * math.Copysign(1, vel[1])
			}

			// Update position
			newX := float64(node.X) + vel[0]
			newY := float64(node.Y) + vel[1]

			// Soft circular boundary
			maxRadius := float64(width/2 - 200)
			currentRadius := math.Sqrt(newX*newX + newY*newY)
			if currentRadius > maxRadius {
				scale := maxRadius / currentRadius * 0.98
				newX *= scale
				newY *= scale
				// Reduce velocity when hitting boundary
				vel[0] *= 0.5
				vel[1] *= 0.5
			}

			node.X = int(newX)
			node.Y = int(newY)
		}

		// Progress indication for long runs
		if iter%10 == 0 && iter > 0 {
			// You could add logging here if needed
			// fmt.Printf("Layout iteration %d/%d completed\n", iter, iterations)
		}
	}
}
