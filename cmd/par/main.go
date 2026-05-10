package main

import (
	"Delta-stepping-algorithm/generator"
	"Delta-stepping-algorithm/graph"
	"Delta-stepping-algorithm/parallel"
	"fmt"
	"log"
	"math"
	"runtime"
	"sort"
	"time"
)

const (
	minWeight = 1.0
	maxWeight = 100.0
	seed      = 42
	source    = 0
	runs      = 20
)

func median(s []float64) float64 {
	sort.Float64s(s)
	mid := len(s) / 2
	return (s[mid-1] + s[mid]) / 2
}

func measure(g *graph.Graph, P int) float64 {
	samples := make([]float64, runs)
	for r := 0; r < runs; r++ {
		t := time.Now()
		parallel.Solve(g, source, P)
		samples[r] = float64(time.Since(t).Nanoseconds()) / 1e6
	}
	return median(samples)
}

func main() {
	P := runtime.NumCPU()
	fmt.Println()
	fmt.Printf("Логічних процесорів: %d\n", P)

	fmt.Println()
	fmt.Println("Дослідження оптимального значення Δ")
	fmt.Printf("Граф: 60 000 вершин, 12M ребер, P=%d\n\n", P)
	fmt.Println("----------------------------------------")
	fmt.Printf("| %-12s | %-20s |\n", "Δ", "Медіанний час, мс")
	fmt.Println("----------------------------------------")

	const researchVertices = 60_000
	const researchEdges = 12_000_000

	deltaValues := []float64{1, 5, 10, 20, 30, 50, 75, 100, 150, 200, 300, 500}

	rawEdges, err := generator.Connected(researchVertices, researchEdges, minWeight, maxWeight, seed)
	if err != nil {
		log.Fatalf("generator: %v", err)
	}

	bestDelta := 0.0
	bestTime := math.MaxFloat64

	for _, delta := range deltaValues {
		g, err := graph.NewGraph(researchVertices, rawEdges, delta)
		if err != nil {
			log.Fatalf("graph delta=%.0f: %v", delta, err)
		}
		parallel.Solve(g, source, P)

		ms := measure(g, P)
		fmt.Printf("| %-12.0f | %-20.3f |\n", delta, ms)

		if ms < bestTime {
			bestTime = ms
			bestDelta = delta
		}
	}
	fmt.Println("-------------------------------------------")
	fmt.Printf("\nОптимальне Δ = %.0f (час: %.3f мс)\n", bestDelta, bestTime)

	fmt.Println()
	fmt.Printf("Паралельний алгоритм Δ-степінгу (Δ=%.0f, P=%d)\n", bestDelta, P)
	fmt.Println("---------------------------------------------------")
	fmt.Printf("| %-10s | %-12s | %-19s |\n", "Вершини", "Ребра", "Медіанний час, мс")
	fmt.Println("---------------------------------------------------")

	sizes := []struct {
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

	{
		raw, _ := generator.Connected(sizes[0].vertices, sizes[0].extraEdges, minWeight, maxWeight, seed)
		g, _ := graph.NewGraph(sizes[0].vertices, raw, bestDelta)
		parallel.Solve(g, source, P)
	}

	for _, s := range sizes {
		raw, err := generator.Connected(s.vertices, s.extraEdges, minWeight, maxWeight, seed)
		if err != nil {
			log.Fatalf("generator: %v", err)
		}
		g, err := graph.NewGraph(s.vertices, raw, bestDelta)
		if err != nil {
			log.Fatalf("graph: %v", err)
		}
		ms := measure(g, P)
		fmt.Printf("| %-10d | %-12d | %-19.3f |\n", s.vertices, len(raw), ms)
	}
	fmt.Println("---------------------------------------------------")
}
