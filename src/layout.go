package main

import (
	"math"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"time"
)

func FruchtermanReingold(graph map[string]*Page, width float64, height float64, iterations int) {
	area := width * height
	n := float64(len(graph))
	if n == 0 {
		return
	}
	compression := 0.7
	k := math.Sqrt(area/n) * compression
	temperature := width / 10
	gravity := 0.05

	centerX := width / 2
	centerY := height / 2

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Estimate frontend radius from weight
	getRadius := func(p *Page) float64 {
		return (5 + math.Sqrt(float64(max(p.weight, 1)))*3) * 2
	}

	// Random initial positions
	for _, p := range graph {
		p.x = rng.Float64() * width
		p.y = rng.Float64() * height
	}

	for iter := 0; iter < iterations; iter++ {
		displacements := make(map[*Page][2]float64)

		// Repulsive forces with collision-aware radius
		for _, v := range graph {
			displacements[v] = [2]float64{0, 0}
			rv := getRadius(v)
			for _, u := range graph {
				if v == u {
					continue
				}
				ru := getRadius(u)
				dx := v.x - u.x
				dy := v.y - u.y
				dist := math.Hypot(dx, dy) + 0.01

				minDist := (rv + ru) / 2 // average radii

				// If overlapping or too close, exaggerate repulsion
				if dist < minDist {
					dist = minDist
				}

				repForce := k * k / dist
				displacements[v] = [2]float64{
					displacements[v][0] + (dx/dist)*repForce,
					displacements[v][1] + (dy/dist)*repForce,
				}
			}
		}

		// Attractive forces
		for _, v := range graph {
			for _, u := range v.out {
				dx := v.x - u.x
				dy := v.y - u.y
				dist := math.Hypot(dx, dy) + 0.01
				attrForce := dist * dist / k
				displacements[v] = [2]float64{
					displacements[v][0] - (dx/dist)*attrForce,
					displacements[v][1] - (dy/dist)*attrForce,
				}
				displacements[u] = [2]float64{
					displacements[u][0] + (dx/dist)*attrForce,
					displacements[u][1] + (dy/dist)*attrForce,
				}
			}
		}

		// Apply displacement + gravity
		for _, v := range graph {
			dx, dy := displacements[v][0], displacements[v][1]
			dx += (centerX - v.x) * gravity
			dy += (centerY - v.y) * gravity

			dist := math.Hypot(dx, dy)
			if dist > 0 {
				limited := math.Min(dist, temperature)
				v.x += (dx / dist) * limited
				v.y += (dy / dist) * limited
			}
		}

		temperature *= 0.95
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func RunSFDP(dotFile string, outputFile string) error {
	cmd := exec.Command("sfdp", "-Tplain", dotFile)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, output, 0644)
}

func AssignCoordinatesWeighted(pages map[string]*Page) {
	// Group pages by layer
	layerMap := make(map[int][]*Page)
	maxLevel := 0
	for _, p := range pages {
		layerMap[p.level] = append(layerMap[p.level], p)
		if p.level > maxLevel {
			maxLevel = p.level
		}
	}

	const layerSpacingX = 150.0
	const baseNodeHeight = 10.0
	const nodeVerticalGap = 5.0

	for level := 0; level <= maxLevel; level++ {
		layer := layerMap[level]

		// Sort nodes by weight descending (or customize your own heuristic)
		sort.Slice(layer, func(i, j int) bool {
			return layer[i].weight > layer[j].weight
		})

		// Total "height" occupied by all nodes (weight-based)
		totalHeight := 0.0
		for _, p := range layer {
			// space required by node = base height * weight + gap
			totalHeight += float64(p.weight)*baseNodeHeight + nodeVerticalGap
		}

		// Start y offset so layer is vertically centered at y=0
		yOffset := -totalHeight / 2

		// Assign positions
		currY := yOffset
		for i, p := range layer {
			// Position X is layer * spacing plus jitter based on index to spread horizontally a bit
			p.x = float64(level)*layerSpacingX + float64(i%5)*10 // jitter up to 50 px horizontally

			// Position Y is current y + half node height to center
			nodeHeight := float64(p.weight)*baseNodeHeight + nodeVerticalGap
			p.y = currY + nodeHeight/2
			currY += nodeHeight
		}
	}
}
