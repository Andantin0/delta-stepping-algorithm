package sequential

import (
	"Delta-stepping-algorithm/graph"
	"Delta-stepping-algorithm/result"
	"fmt"
	"math"
	"time"
)

func Solve(g *graph.Graph, source int) (*result.Result, error) {
	if source < 0 || source >= g.NumVertices {
		return nil, fmt.Errorf("sequential: source vertex %d out of range [0, %d)", source, g.NumVertices)
	}

	start := time.Now()

	n := g.NumVertices
	dist := make([]float64, n)
	for i := range dist {
		dist[i] = result.Inf
	}
	dist[source] = 0

	buckets := []map[int]struct{}{}

	ensureBucket := func(idx int) {
		for len(buckets) <= idx {
			buckets = append(buckets, make(map[int]struct{}))
		}
	}

	bucketIdx := func(d float64) int {
		return int(d / g.Delta)
	}

	relax := func(v int, newDist float64) bool {
		if newDist >= dist[v] {
			return false
		}
		if dist[v] < result.Inf {
			oldIdx := bucketIdx(dist[v])
			if oldIdx < len(buckets) {
				delete(buckets[oldIdx], v)
			}
		}
		dist[v] = newDist
		idx := bucketIdx(newDist)
		ensureBucket(idx)
		buckets[idx][v] = struct{}{}
		return true
	}

	ensureBucket(0)
	buckets[0][source] = struct{}{}

	for i := 0; i < len(buckets); i++ {
		if len(buckets[i]) == 0 {
			continue
		}

		S := make([]int, 0)

		for len(buckets[i]) > 0 {
			current := make([]int, 0, len(buckets[i]))
			for v := range buckets[i] {
				current = append(current, v)
			}
			buckets[i] = make(map[int]struct{})

			for _, u := range current {
				S = append(S, u)
				for _, e := range g.Light[u] {
					relax(e.To, dist[u]+e.Weight)
				}
			}
		}

		for _, u := range S {
			for _, e := range g.Heavy[u] {
				relax(e.To, dist[u]+e.Weight)
			}
		}
	}

	for i, d := range dist {
		if d >= result.Inf {
			dist[i] = math.MaxFloat64
		}
	}

	return &result.Result{
		Dist:          dist,
		ExecutionTime: time.Since(start).Milliseconds(),
	}, nil
}
