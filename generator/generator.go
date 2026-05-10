package generator

import (
	"Delta-stepping-algorithm/graph"
	"fmt"
	"math/rand"
)

func Connected(
	numVertices, extraEdges int,
	minWeight, maxWeight float64,
	seed int64,
) ([]graph.RawEdge, error) {
	if numVertices < 2 {
		return nil, fmt.Errorf("generator: numVertices must be >= 2, got %d", numVertices)
	}
	if extraEdges < 0 {
		return nil, fmt.Errorf("generator: extraEdges must be >= 0, got %d", extraEdges)
	}
	if minWeight < 0 {
		return nil, fmt.Errorf("generator: minWeight must be >= 0 (delta-stepping requires non-negative weights), got %f", minWeight)
	}
	if minWeight > maxWeight {
		return nil, fmt.Errorf("generator: minWeight (%.4f) must be <= maxWeight (%.4f)", minWeight, maxWeight)
	}

	rng := rand.New(rand.NewSource(seed))
	edges := make([]graph.RawEdge, 0, numVertices-1+extraEdges)

	perm := rng.Perm(numVertices)

	for i := 1; i < numVertices; i++ {
		w := minWeight + rng.Float64()*(maxWeight-minWeight)
		edges = append(edges, graph.RawEdge{
			From: perm[i-1], To: perm[i], Weight: w,
		})
	}

	for added := 0; added < extraEdges; added++ {
		u := rng.Intn(numVertices)
		v := rng.Intn(numVertices)
		if u == v {
			v = (v + 1) % numVertices
		}
		w := minWeight + rng.Float64()*(maxWeight-minWeight)
		edges = append(edges, graph.RawEdge{From: u, To: v, Weight: w})
	}

	return edges, nil
}

func Random(
	numVertices int,
	edgeProb float64,
	minWeight, maxWeight float64,
	seed int64,
) ([]graph.RawEdge, error) {
	if numVertices < 1 {
		return nil, fmt.Errorf("generator: numVertices must be >= 1, got %d", numVertices)
	}
	if edgeProb < 0 || edgeProb > 1 {
		return nil, fmt.Errorf("generator: edgeProb must be in [0, 1], got %f", edgeProb)
	}
	if minWeight < 0 {
		return nil, fmt.Errorf("generator: minWeight must be >= 0, got %f", minWeight)
	}
	if minWeight > maxWeight {
		return nil, fmt.Errorf("generator: minWeight (%.4f) must be <= maxWeight (%.4f)", minWeight, maxWeight)
	}

	rng := rand.New(rand.NewSource(seed))
	edges := make([]graph.RawEdge, 0)

	for u := 0; u < numVertices; u++ {
		for v := 0; v < numVertices; v++ {
			if u != v && rng.Float64() < edgeProb {
				w := minWeight + rng.Float64()*(maxWeight-minWeight)
				edges = append(edges, graph.RawEdge{From: u, To: v, Weight: w})
			}
		}
	}
	return edges, nil
}
