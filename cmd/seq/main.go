package main

import (
	"Delta-stepping-algorithm/generator"
	"Delta-stepping-algorithm/graph"
	"Delta-stepping-algorithm/sequential"
	"fmt"
	"log"
	"sort"
	"time"
)

const (
	delta     = 50.0
	minWeight = 1.0
	maxWeight = 100.0
	seed      = 42
	source    = 0
	runs      = 20
)

var sizes = []struct {
	vertices   int
	extraEdges int
}{
	{5_000, 1_000_000},
	{10_000, 2_000_000},
	{15_000, 3_000_000},
	{20_000, 4_000_000},
	{30_000, 6_000_000},
	{40_000, 8_000_000},
	{50_000, 10_000_000},
	{60_000, 12_000_000},
	{70_000, 14_000_000},
}

func median(s []float64) float64 {
	sort.Float64s(s)
	mid := len(s) / 2
	return (s[mid-1] + s[mid]) / 2
}

func main() {
	{
		raw, _ := generator.Connected(sizes[0].vertices, sizes[0].extraEdges, minWeight, maxWeight, seed)
		g, _ := graph.NewGraph(sizes[0].vertices, raw, delta)
		sequential.Solve(g, source)
	}

	fmt.Println()
	fmt.Println("Послідовний алгоритм Δ-степінгу")
	fmt.Println("---------------------------------------------------")
	fmt.Printf("| %-10s | %-12s | %-19s |\n", "Вершини", "Ребра", "Медіанний час, мс")
	fmt.Println("---------------------------------------------------")

	for _, s := range sizes {
		raw, err := generator.Connected(s.vertices, s.extraEdges, minWeight, maxWeight, seed)
		if err != nil {
			log.Fatalf("generator: %v", err)
		}
		g, err := graph.NewGraph(s.vertices, raw, delta)
		if err != nil {
			log.Fatalf("graph: %v", err)
		}

		samples := make([]float64, runs)
		for r := 0; r < runs; r++ {
			t := time.Now()
			sequential.Solve(g, source)
			samples[r] = float64(time.Since(t).Nanoseconds()) / 1e6
		}
		fmt.Printf("| %-10d | %-12d | %-19.3f |\n", s.vertices, len(raw), median(samples))
	}
	fmt.Println("---------------------------------------------------")
}
