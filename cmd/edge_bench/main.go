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
	delta         = 10.0
	minWeight     = 1.0
	maxWeight     = 100.0
	seed          = 42
	source        = 0
	runs          = 20
	fixedVertices = 30_000
)

var edgeCounts = []int{
	200_000,
	500_000,
	1_000_000,
	2_000_000,
	3_000_000,
	4_000_000,
	6_000_000,
	8_000_000,
	10_000_000,
	12_000_000,
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
		raw, _ := generator.Connected(fixedVertices, edgeCounts[0], minWeight, maxWeight, seed)
		g, _ := graph.NewGraph(fixedVertices, raw, delta)
		sequential.Solve(g, source)
		parallel.Solve(g, source, 8)
	}

	graphs := make([]*graph.Graph, len(edgeCounts))
	realEdges := make([]int, len(edgeCounts))
	for i, edges := range edgeCounts {
		raw, err := generator.Connected(fixedVertices, edges, minWeight, maxWeight, seed)
		if err != nil {
			log.Fatalf("generator: %v", err)
		}

		g, err := graph.NewGraph(fixedVertices, raw, delta)
		if err != nil {
			log.Fatalf("graph: %v", err)
		}

		graphs[i] = g
		realEdges[i] = len(raw)
	}

	for _, P := range threadCounts {
		fmt.Printf("\nP = %d потоків, V = %d вершин\n", P, fixedVertices)
		fmt.Println("---------------------------------------------------------------------------------------")
		fmt.Printf("| %-12s | %-18s | %-18s | %-11s | %-12s |\n",
			"Ребра", "Посл. час, мс", "Пар. час, мс", "Прискорення", "Ефективність")
		fmt.Println("---------------------------------------------------------------------------------------")

		for i := range edgeCounts {
			g := graphs[i]

			parMs := measure(func() { parallel.Solve(g, source, P) })
			seqMs := measure(func() { sequential.Solve(g, source) })

			speedup := seqMs / parMs
			efficiency := speedup / float64(P)

			fmt.Printf("| %-12d | %-18.3f | %-18.3f | %-11.3f | %-12.3f |\n",
				realEdges[i], seqMs, parMs, speedup, efficiency)
		}
		fmt.Println("---------------------------------------------------------------------------------------")
	}
}
