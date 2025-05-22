package main

import (
	"context"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"sync"
	"time"
)

// QuadTree node for Barnes-Hut approximation
type QuadTree struct {
	centerX, centerY float32
	width, height    float32
	mass             float32
	centerOfMassX    float32
	centerOfMassY    float32
	children         [4]*QuadTree
	page             *Page
}

// Create a new QuadTree node
func newQuadTree(x, y, w, h float32) *QuadTree {
	return &QuadTree{
		centerX: x,
		centerY: y,
		width:   w,
		height:  h,
	}
}

// Insert a page into the QuadTree using a non-recursive approach
func (qt *QuadTree) insert(p *Page) {
	const maxDepth = 20
	const minCellSize = 1.0
	current := qt
	depth := 0
	for {
		if current.page == nil && current.children[0] == nil {
			current.page = p
			current.mass = float32(p.weight)
			current.centerOfMassX = p.x
			current.centerOfMassY = p.y
			return
		}

		if current.children[0] == nil {
			// Stop subdividing if max depth or min cell size reached
			if depth >= maxDepth || current.width < minCellSize || current.height < minCellSize {
				log.Printf("[QuadTree] Max depth (%d) or min cell size (%.2f) reached at depth %d, width=%.2f, height=%.2f", maxDepth, minCellSize, depth, current.width, current.height)
				// Place the page here even if it overlaps
				return
			}
			current.subdivide()
			existingPage := current.page
			current.page = nil
			current.insertIntoChild(existingPage)
		}

		// Update center of mass
		totalMass := current.mass + float32(p.weight)
		current.centerOfMassX = (current.centerOfMassX*current.mass + p.x*float32(p.weight)) / totalMass
		current.centerOfMassY = (current.centerOfMassY*current.mass + p.y*float32(p.weight)) / totalMass
		current.mass = totalMass

		// Determine which child to insert into
		quadrant := 0
		if p.x > current.centerX {
			quadrant += 1
		}
		if p.y > current.centerY {
			quadrant += 2
		}

		// Move to child node
		current = current.children[quadrant]
		depth++
	}
}

// Subdivide the current QuadTree node
func (qt *QuadTree) subdivide() {
	halfW := qt.width / 2
	halfH := qt.height / 2
	qt.children = [4]*QuadTree{
		newQuadTree(qt.centerX-halfW, qt.centerY-halfH, halfW, halfH), // NW
		newQuadTree(qt.centerX+halfW, qt.centerY-halfH, halfW, halfH), // NE
		newQuadTree(qt.centerX-halfW, qt.centerY+halfH, halfW, halfH), // SW
		newQuadTree(qt.centerX+halfW, qt.centerY+halfH, halfW, halfH), // SE
	}
}

// Insert a page into the appropriate child node
func (qt *QuadTree) insertIntoChild(p *Page) {
	quadrant := 0
	if p.x > qt.centerX {
		quadrant += 1
	}
	if p.y > qt.centerY {
		quadrant += 2
	}
	qt.children[quadrant].insert(p)
}

// Calculate forces using Barnes-Hut approximation
func (qt *QuadTree) calculateForces(p *Page, theta float32, k float32, forces *[2]float32) {
	if qt.mass == 0 {
		return
	}

	dx := qt.centerOfMassX - p.x
	dy := qt.centerOfMassY - p.y
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if dist == 0 {
		return
	}

	// If the node is far enough or is a leaf, treat it as a single body
	if qt.width/dist < theta || qt.page != nil {
		force := k * k / dist
		forces[0] += dx * force
		forces[1] += dy * force
		return
	}

	// Otherwise, recursively process children
	for _, child := range qt.children {
		child.calculateForces(p, theta, k, forces)
	}
}

// ParallelForceCalculation calculates forces for a subset of pages
func ParallelForceCalculation(pages []*Page, start, end int, qt *QuadTree, theta, k float32, forces [][2]float32, wg *sync.WaitGroup) {
	for i := start; i < end; i++ {
		p := pages[i]
		qt.calculateForces(p, theta, k, &forces[i])
	}
}

// resolveCollisionsSpatial resolves node overlaps using a spatial hash grid (O(n) per iteration)
func resolveCollisionsSpatial(pages []*Page, minNodeRadius, padding, gridCellSize float32) {
	// Build spatial grid
	grid := make(map[[2]int][]*Page)
	for _, p := range pages {
		gx := int(p.x / gridCellSize)
		gy := int(p.y / gridCellSize)
		key := [2]int{gx, gy}
		grid[key] = append(grid[key], p)
	}
	// Check each node only against neighbors in the same and adjacent cells
	for _, p := range pages {
		gx := int(p.x / gridCellSize)
		gy := int(p.y / gridCellSize)
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				key := [2]int{gx + dx, gy + dy}
				for _, q := range grid[key] {
					if p == q {
						continue
					}
					r := minNodeRadius
					dx := q.x - p.x
					dy := q.y - p.y
					dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
					minDist := 2*r + padding
					if dist < minDist && dist > 0 {
						overlap := (minDist - dist) / 2
						if dist > 0 {
							ox := dx / dist * overlap
							oy := dy / dist * overlap
							p.x -= ox
							p.y -= oy
							q.x += ox
							q.y += oy
						}
					}
				}
			}
		}
	}
}

// workerTask represents a chunk of work for the worker pool
type workerTask struct {
	start, end int
}

// OptimizedFruchtermanReingold implements a parallel version of the F-R algorithm with Barnes-Hut approximation
func OptimizedFruchtermanReingold(pages []*Page, width, height float32, iterations int) {
	if len(pages) == 0 {
		return
	}

	iterations = 75
	log.Printf("Starting layout calculation for %d pages", len(pages))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const (
		minNodeWidth  float32 = 200.0
		minNodeHeight float32 = 80.0
		nodePadding   float32 = 100.0
		edgePadding   float32 = 200.0
	)

	effectiveWidth := width - 2*edgePadding
	effectiveHeight := height - 2*edgePadding
	area := effectiveWidth * effectiveHeight
	n := float32(len(pages))
	k := float32(math.Sqrt(float64(area/n))) * 2.0
	temperature := width / 2
	theta := float32(0.5)
	minDist := minNodeWidth + nodePadding

	log.Printf("Layout parameters: k=%.2f, initial temperature=%.2f, minDist=%.2f", k, temperature, minDist)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	centerX := width / 2
	centerY := height / 2
	maxRadius := float32(math.Min(float64(effectiveWidth), float64(effectiveHeight))) / 2

	for _, p := range pages {
		angle := rng.Float32() * 2 * math.Pi
		radius := rng.Float32() * maxRadius
		p.x = centerX + radius*float32(math.Cos(float64(angle)))
		p.y = centerY + radius*float32(math.Sin(float64(angle)))
	}

	numWorkers := runtime.NumCPU()
	chunkSize := (len(pages) + numWorkers - 1) / numWorkers
	log.Printf("Using %d workers for parallel processing, chunk size: %d", numWorkers, chunkSize)

	qt := newQuadTree(width/2, height/2, width, height)
	for i, p := range pages {
		qt.insert(p)
		if i%500 == 0 {
			log.Printf("Initial QuadTree construction: %d/%d pages", i, len(pages))
		}
	}

	for iter := 0; iter < iterations; iter++ {
		select {
		case <-ctx.Done():
			log.Printf("Layout calculation timed out after %d iterations", iter)
			return
		default:
		}

		if iter%3 == 0 {
			log.Printf("Starting iteration %d/%d (temperature: %.2f)", iter+1, iterations, temperature)
		}

		// Calculate repulsive forces in parallel using a worker pool
		forces := make([][2]float32, len(pages))
		tasks := make(chan workerTask, numWorkers)
		var wg sync.WaitGroup

		// Start worker goroutines
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[Worker %d] PANIC: %v", workerID, r)
					}
					wg.Done()
					log.Printf("[Worker %d] finished", workerID)
				}()
				log.Printf("[Worker %d] started", workerID)
				for task := range tasks {
					for i := task.start; i < task.end; i++ {
						p := pages[i]
						qt.calculateForces(p, theta, k, &forces[i])
					}
				}
			}(w)
		}

		// Distribute tasks
		for i := 0; i < len(pages); i += chunkSize {
			end := i + chunkSize
			if end > len(pages) {
				end = len(pages)
			}
			tasks <- workerTask{start: i, end: end}
		}
		close(tasks)
		wg.Wait()
		log.Printf("[Layout] All workers finished for iteration %d", iter)

		if iter%3 == 0 {
			log.Printf("Calculated repulsive forces for all pages")
		}

		// Calculate attractive forces with minimum distance
		attractiveForces := 0
		for i, p := range pages {
			for _, outIdx := range p.out {
				if outIdx >= uint32(len(pages)) {
					continue
				}
				u := pages[outIdx]
				dx := p.x - u.x
				dy := p.y - u.y
				dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

				// Ensure minimum distance based on node dimensions
				if dist < minDist {
					// Strong repulsive force when too close
					force := (minDist - dist) * 3.0 // Increased repulsive force
					dx = dx / dist * force
					dy = dy / dist * force
					forces[i][0] += dx
					forces[i][1] += dy
					forces[outIdx][0] -= dx
					forces[outIdx][1] -= dy
					continue
				}

				// Normal attractive force
				attrForce := (dist*dist - k*k) / k
				if attrForce < 0 {
					attrForce = 0
				}

				forces[i][0] -= dx * attrForce / dist
				forces[i][1] -= dy * attrForce / dist
				forces[outIdx][0] += dx * attrForce / dist
				forces[outIdx][1] += dy * attrForce / dist
				attractiveForces++
			}
		}

		if iter%3 == 0 {
			log.Printf("Calculated %d attractive forces", attractiveForces)
		}

		// Apply forces with boundary constraints
		maxForce := float32(0)
		for i, p := range pages {
			dx, dy := forces[i][0], forces[i][1]

			// Calculate distance from center
			distFromCenter := float32(math.Sqrt(float64((p.x-centerX)*(p.x-centerX) + (p.y-centerY)*(p.y-centerY))))

			// Add radial force to prevent central clustering
			if distFromCenter < maxRadius*0.3 {
				// Strong outward force for nodes too close to center
				radialForce := (maxRadius*0.3 - distFromCenter) * 0.5
				dx += (p.x - centerX) / distFromCenter * radialForce
				dy += (p.y - centerY) / distFromCenter * radialForce
			}

			// Weaker gravity force
			dx += (width/2 - p.x) * 0.1
			dy += (height/2 - p.y) * 0.1

			dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if dist > 0 {
				limited := float32(math.Min(float64(dist), float64(temperature)))
				p.x += dx * limited / dist
				p.y += dy * limited / dist

				// Keep nodes within bounds with proper padding
				p.x = float32(math.Max(float64(edgePadding), math.Min(float64(width-edgePadding-minNodeWidth), float64(p.x))))
				p.y = float32(math.Max(float64(edgePadding), math.Min(float64(height-edgePadding-minNodeHeight), float64(p.y))))

				if dist > maxForce {
					maxForce = dist
				}
			}
		}

		if iter%3 == 0 {
			log.Printf("Applied forces (max force: %.2f)", maxForce)
		}

		// Add this line to resolve collisions efficiently
		for pass := 0; pass < 3; pass++ {
			resolveCollisionsSpatial(pages, minNodeWidth/2, nodePadding, minNodeWidth+nodePadding)
		}

		// Update QuadTree only every 3 iterations
		if iter%3 == 0 {
			qt = newQuadTree(width/2, height/2, width, height)
			for i, p := range pages {
				qt.insert(p)
				if i%500 == 0 {
					log.Printf("Updating QuadTree: %d/%d pages", i, len(pages))
				}
			}
		}

		// Cool down slower to allow more movement
		temperature *= 0.9
	}

	for pass := 0; pass < 20; pass++ {
		resolveCollisionsSpatial(pages, minNodeWidth/2, nodePadding, minNodeWidth+nodePadding)
	}

	log.Printf("Layout calculation completed after %d iterations", iterations)
}

// AssignCoordinatesWeighted assigns coordinates using a hierarchical layout
func AssignCoordinatesWeighted(pages []*Page) {
	log.Printf("Starting weighted layout for %d pages", len(pages))

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Group pages by level
	layerMap := make(map[uint8][]*Page)
	var maxLevel uint8
	for _, p := range pages {
		select {
		case <-ctx.Done():
			log.Printf("Layout calculation timed out")
			return
		default:
		}

		layerMap[p.level] = append(layerMap[p.level], p)
		if p.level > maxLevel {
			maxLevel = p.level
		}
	}

	log.Printf("Found %d levels in the hierarchy", maxLevel+1)

	const layerSpacingX float32 = 200.0
	const baseNodeHeight float32 = 15.0
	const nodeVerticalGap float32 = 10.0
	const nodeHorizontalGap float32 = 30.0

	for level := uint8(0); level <= maxLevel; level++ {
		select {
		case <-ctx.Done():
			log.Printf("Layout calculation timed out")
			return
		default:
		}

		layer := layerMap[level]
		log.Printf("Processing level %d with %d nodes", level, len(layer))

		// Sort nodes by weight
		sort.Slice(layer, func(i, j int) bool {
			return layer[i].weight > layer[j].weight
		})

		// Calculate total height
		var totalHeight float32
		for _, p := range layer {
			totalHeight += float32(p.weight)*baseNodeHeight + nodeVerticalGap
		}

		// Assign positions with improved spacing
		yOffset := -totalHeight / 2
		var currY float32 = yOffset

		for _, p := range layer {
			// Add horizontal offset based on weight
			weightOffset := float32(p.weight) * nodeHorizontalGap
			p.x = float32(level)*layerSpacingX + weightOffset

			// Calculate node height based on weight
			nodeHeight := float32(p.weight)*baseNodeHeight + nodeVerticalGap
			p.y = currY + nodeHeight/2
			currY += nodeHeight
		}

		log.Printf("Completed level %d layout", level)
	}

	log.Printf("Weighted layout completed")
}

// RunSFDP runs the SFDP layout algorithm
func RunSFDP(dotFile string, outputFile string) error {
	cmd := exec.Command("sfdp", "-Tplain", dotFile)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, output, 0644)
}
