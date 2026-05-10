package graph

import "fmt"

type Edge struct {
	To     int
	Weight float64
}

type Graph struct {
	NumVertices int
	Light       [][]Edge
	Heavy       [][]Edge
	Delta       float64
}

func NewGraph(numVertices int, edges []RawEdge, delta float64) (*Graph, error) {
	if numVertices < 1 {
		return nil, fmt.Errorf("graph: numVertices must be >= 1, got %d", numVertices)
	}
	if delta <= 0 {
		return nil, fmt.Errorf("graph: delta must be > 0, got %f", delta)
	}

	g := &Graph{
		NumVertices: numVertices,
		Light:       make([][]Edge, numVertices),
		Heavy:       make([][]Edge, numVertices),
		Delta:       delta,
	}
	for i := range g.Light {
		g.Light[i] = []Edge{}
		g.Heavy[i] = []Edge{}
	}

	for _, e := range edges {
		if e.From < 0 || e.From >= numVertices {
			return nil, fmt.Errorf("graph: edge source vertex %d out of range [0, %d)", e.From, numVertices)
		}
		if e.To < 0 || e.To >= numVertices {
			return nil, fmt.Errorf("graph: edge target vertex %d out of range [0, %d)", e.To, numVertices)
		}
		if e.Weight < 0 {
			return nil, fmt.Errorf("graph: negative edge weight %.4f on edge (%d->%d); delta-stepping requires non-negative weights", e.Weight, e.From, e.To)
		}

		edge := Edge{To: e.To, Weight: e.Weight}
		if e.Weight <= delta {
			g.Light[e.From] = append(g.Light[e.From], edge)
		} else {
			g.Heavy[e.From] = append(g.Heavy[e.From], edge)
		}
	}
	return g, nil
}

type RawEdge struct {
	From   int
	To     int
	Weight float64
}
