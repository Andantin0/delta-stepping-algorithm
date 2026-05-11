package main

import (
	"Delta-stepping-algorithm/generator"
	"Delta-stepping-algorithm/graph"
	"Delta-stepping-algorithm/parallel"
	"Delta-stepping-algorithm/sequential"
	"fmt"
	"log"
	"sort"
	"time"
)

const (
	delta      = 10.0
	minWeight  = 1.0
	maxWeight  = 100.0
	seed       = 42
	source     = 0
	runs       = 20
	fixedEdges = 6_000_000
)

var vertexCounts = []int{
	5_000,
	10_000,
	15_000,
	20_000,
	25_000,
	30_000,
	40_000,
	50_000,
	60_000,
	70_000,
}

var threadCounts = []int{2, 4, 6, 8, 9, 10, 11, 12}

func median(s []float64) float64 {
	sort.Float64s(s)
	mid := len(s) / 2
	return (s[mid-1] + s[mid]) / 2
}

func measure(fn func()) float64 {
	samples := make([]float64, runs)
	for r := 0; r < runs; r++ {
		t := time.Now()
		fn()
		samples[r] = float64(time.Since(t).Nanoseconds()) / 1e6
	}
	return median(samples)
}

func main() {
	{
		raw, _ := generator.Connected(vertexCounts[0], fixedEdges, minWeight, maxWeight, seed)
		g, _ := graph.NewGraph(vertexCounts[0], raw, delta)
		sequential.Solve(g, source)
		parallel.Solve(g, source, 8)
	}

	graphs := make([]*graph.Graph, len(vertexCounts))
	for i, vertices := range vertexCounts {
		raw, err := generator.Connected(vertices, fixedEdges, minWeight, maxWeight, seed)
		if err != nil {
			log.Fatalf("generator: %v", err)
		}

		g, err := graph.NewGraph(vertices, raw, delta)
		if err != nil {
			log.Fatalf("graph: %v", err)
		}
		graphs[i] = g
	}

	for _, P := range threadCounts {
		fmt.Printf("\nP = %d потоків, E = %d ребер\n", P, fixedEdges)
		fmt.Println("-------------------------------------------------------------------------------------")
		fmt.Printf("| %-10s | %-18s | %-18s | %-11s | %-12s |\n",
			"Вершини", "Посл. час, мс", "Пар. час, мс", "Прискорення", "Ефективність")
		fmt.Println("-------------------------------------------------------------------------------------")

		for i, vertices := range vertexCounts {
			g := graphs[i]

			parMs := measure(func() { parallel.Solve(g, source, P) })
			seqMs := measure(func() { sequential.Solve(g, source) })

			speedup := seqMs / parMs
			efficiency := speedup / float64(P)

			fmt.Printf("| %-10d | %-18.3f | %-18.3f | %-11.3f | %-12.3f |\n",
				vertices, seqMs, parMs, speedup, efficiency)
		}
		fmt.Println("-------------------------------------------------------------------------------------")
	}
}
