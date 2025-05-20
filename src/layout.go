package main

import (
	"math"
	"math/rand"
	"time"
)

func FruchtermanReingold(graph map[string]*Page, width, height float64, iterations int) {
	area := width * height
	n := float64(len(graph))
	if n == 0 {
		return
	}
	compression := 0.75
	k := math.Sqrt(area/n) * compression
	temperature := width / 10

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Initialize positions randomly
	for _, p := range graph {
		p.x = rng.Float64() * width
		p.y = rng.Float64() * height
	}

	for iter := 0; iter < iterations; iter++ {
		displacements := make(map[*Page][2]float64)

		// Repulsive forces
		for _, v := range graph {
			displacements[v] = [2]float64{0, 0}
			for _, u := range graph {
				if v == u {
					continue
				}
				dx := v.x - u.x
				dy := v.y - u.y
				dist := math.Hypot(dx, dy) + 0.01
				repForce := k * k / dist
				displacements[v] = [2]float64{
					displacements[v][0] + (dx/dist)*repForce,
					displacements[v][1] + (dy/dist)*repForce,
				}
			}
		}

		// Attractive forces (use both in and out links to ensure bi-directional spring forces)
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

		// Apply displacements and limit by temperature
		for _, v := range graph {
			dx, dy := displacements[v][0], displacements[v][1]
			dist := math.Hypot(dx, dy)
			if dist > 0 {
				limited := math.Min(dist, temperature)
				v.x += (dx / dist) * limited
				v.y += (dy / dist) * limited
			}
			// Keep inside bounds
			v.x = math.Min(width, math.Max(0, v.x))
			v.y = math.Min(height, math.Max(0, v.y))
		}

		// Cool temperature
		temperature *= 0.95
	}
}
