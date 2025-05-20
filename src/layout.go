package main

import (
	"math"
	"math/rand"
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
