package test

import (
	"Delta-stepping-algorithm/graph"
	"Delta-stepping-algorithm/result"
	"Delta-stepping-algorithm/sequential"
	"math"
	"testing"
)

const delta = 10.0

func mustSolve(t *testing.T, numV int, raw []graph.RawEdge, src int) []float64 {
	t.Helper()
	g, err := graph.NewGraph(numV, raw, delta)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}
	res, err := sequential.Solve(g, src)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	return res.Dist
}

func approxEqual(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

func TestBasicCorrectness(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 2},
		{From: 0, To: 2, Weight: 6},
		{From: 1, To: 2, Weight: 3},
		{From: 1, To: 3, Weight: 8},
		{From: 2, To: 3, Weight: 1},
	}
	dist := mustSolve(t, 4, raw, 0)

	expected := []float64{0, 2, 5, 6}
	for v, want := range expected {
		if !approxEqual(dist[v], want, 1e-9) {
			t.Errorf("dist[%d] = %.4f, want %.4f", v, dist[v], want)
		}
	}
}

func TestUnreachableVertices(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 3},
	}
	dist := mustSolve(t, 4, raw, 0)

	if dist[0] != 0 {
		t.Errorf("dist[0] = %.4f, want 0", dist[0])
	}
	if !approxEqual(dist[1], 3, 1e-9) {
		t.Errorf("dist[1] = %.4f, want 3", dist[1])
	}
	if dist[2] != math.MaxFloat64 {
		t.Errorf("dist[2] = %.4f, want +Inf", dist[2])
	}
	if dist[3] != math.MaxFloat64 {
		t.Errorf("dist[3] = %.4f, want +Inf", dist[3])
	}
}

func TestShortestAmongMultiplePaths(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 10},
		{From: 0, To: 2, Weight: 3},
		{From: 2, To: 1, Weight: 2},
	}
	dist := mustSolve(t, 3, raw, 0)

	if !approxEqual(dist[1], 5, 1e-9) {
		t.Errorf("dist[1] = %.4f, want 5 (шлях 0->2->1)", dist[1])
	}
}

func TestChainGraph(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 1},
		{From: 1, To: 2, Weight: 1},
		{From: 2, To: 3, Weight: 1},
		{From: 3, To: 4, Weight: 1},
	}
	dist := mustSolve(t, 5, raw, 0)

	expected := []float64{0, 1, 2, 3, 4}
	for v, want := range expected {
		if !approxEqual(dist[v], want, 1e-9) {
			t.Errorf("dist[%d] = %.4f, want %.4f", v, dist[v], want)
		}
	}
}

func TestStarGraph(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 3},
		{From: 0, To: 2, Weight: 7},
		{From: 0, To: 3, Weight: 1},
		{From: 0, To: 4, Weight: 9},
	}
	dist := mustSolve(t, 5, raw, 0)

	expected := map[int]float64{0: 0, 1: 3, 2: 7, 3: 1, 4: 9}
	for v, want := range expected {
		if !approxEqual(dist[v], want, 1e-9) {
			t.Errorf("dist[%d] = %.4f, want %.4f", v, dist[v], want)
		}
	}
}

func TestNoEdges(t *testing.T) {
	dist := mustSolve(t, 4, []graph.RawEdge{}, 0)

	if dist[0] != 0 {
		t.Errorf("dist[0] = %.4f, want 0", dist[0])
	}
	for v := 1; v < 4; v++ {
		if dist[v] != math.MaxFloat64 {
			t.Errorf("dist[%d] = %.4f, want +Inf", v, dist[v])
		}
	}
}

func TestMinimalGraph(t *testing.T) {
	raw := []graph.RawEdge{{From: 0, To: 1, Weight: 7}}
	dist := mustSolve(t, 2, raw, 0)

	if dist[0] != 0 {
		t.Errorf("dist[0] = %.4f, want 0", dist[0])
	}
	if !approxEqual(dist[1], 7, 1e-9) {
		t.Errorf("dist[1] = %.4f, want 7", dist[1])
	}
}

func TestHeavyEdgesProcessedAfterBucket(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 5},
		{From: 0, To: 2, Weight: 50},
		{From: 1, To: 2, Weight: 3},
	}
	dist := mustSolve(t, 3, raw, 0)

	if !approxEqual(dist[1], 5, 1e-9) {
		t.Errorf("dist[1] = %.4f, want 5", dist[1])
	}
	if !approxEqual(dist[2], 8, 1e-9) {
		t.Errorf("dist[2] = %.4f, want 8 (шлях 0->1->2)", dist[2])
	}
}

func TestInvalidSource(t *testing.T) {
	raw := []graph.RawEdge{{From: 0, To: 1, Weight: 1}}
	g, _ := graph.NewGraph(2, raw, delta)

	_, err := sequential.Solve(g, -1)
	if err == nil {
		t.Error("очікується помилка для source=-1, отримано nil")
	}
	_, err = sequential.Solve(g, 5)
	if err == nil {
		t.Error("очікується помилка для source=5 (поза межами), отримано nil")
	}
}

func TestInvalidGraphParams(t *testing.T) {
	raw := []graph.RawEdge{{From: 0, To: 1, Weight: 1}}

	_, err := graph.NewGraph(0, raw, delta)
	if err == nil {
		t.Error("очікується помилка для numVertices=0")
	}
	_, err = graph.NewGraph(2, raw, -1)
	if err == nil {
		t.Error("очікується помилка для delta=-1")
	}
	negRaw := []graph.RawEdge{{From: 0, To: 1, Weight: -5}}
	_, err = graph.NewGraph(2, negRaw, delta)
	if err == nil {
		t.Error("очікується помилка для від'ємної ваги ребра")
	}
}

func TestLargeGraphNoOverflow(t *testing.T) {
	const n = 1000
	raw := make([]graph.RawEdge, 0, n-1)
	for i := 0; i < n-1; i++ {
		raw = append(raw, graph.RawEdge{From: i, To: i + 1, Weight: 1})
	}
	dist := mustSolve(t, n, raw, 0)

	if dist[0] != 0 {
		t.Errorf("dist[0] = %.4f, want 0", dist[0])
	}
	for v, d := range dist {
		if d < 0 || (d != math.MaxFloat64 && d >= result.Inf) {
			t.Errorf("dist[%d] = %.4f виглядає як переповнення", v, d)
		}
	}
}
