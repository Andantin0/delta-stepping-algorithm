package test

import (
	"Delta-stepping-algorithm/generator"
	"Delta-stepping-algorithm/graph"
	"Delta-stepping-algorithm/parallel"
	"math"
	"testing"
)

const parallelDelta = 10.0

func mustSolveParallel(t *testing.T, numV int, raw []graph.RawEdge, src, P int) []float64 {
	t.Helper()
	g, err := graph.NewGraph(numV, raw, parallelDelta)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}
	res, err := parallel.Solve(g, src, P)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	return res.Dist
}

func TestParallelMatchesSequential(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 2},
		{From: 0, To: 2, Weight: 6},
		{From: 1, To: 2, Weight: 3},
		{From: 1, To: 3, Weight: 8},
		{From: 2, To: 3, Weight: 1},
	}
	expected := []float64{0, 2, 5, 6}

	for _, P := range []int{1, 4, 8} {
		dist := mustSolveParallel(t, 4, raw, 0, P)
		for v, want := range expected {
			if !approxEqual(dist[v], want, 1e-9) {
				t.Errorf("P=%d: dist[%d] = %.4f, want %.4f", P, v, dist[v], want)
			}
		}
	}
}

func TestParallelDeterministicAcrossThreadCounts(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 5},
		{From: 0, To: 2, Weight: 3},
		{From: 2, To: 1, Weight: 1},
		{From: 1, To: 3, Weight: 4},
		{From: 2, To: 3, Weight: 9},
	}
	g, _ := graph.NewGraph(4, raw, parallelDelta)
	ref, _ := parallel.Solve(g, 0, 1)

	for _, P := range []int{2, 4, 6, 8, 12} {
		res, err := parallel.Solve(g, 0, P)
		if err != nil {
			t.Fatalf("P=%d: Solve error: %v", P, err)
		}
		for v := range ref.Dist {
			if !approxEqual(ref.Dist[v], res.Dist[v], 1e-9) {
				t.Errorf("P=%d: dist[%d] = %.4f, want %.4f",
					P, v, res.Dist[v], ref.Dist[v])
			}
		}
	}
}

func TestParallelReproducible(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 3},
		{From: 0, To: 2, Weight: 7},
		{From: 1, To: 3, Weight: 2},
		{From: 2, To: 3, Weight: 1},
		{From: 1, To: 2, Weight: 1},
	}
	g, _ := graph.NewGraph(4, raw, parallelDelta)
	ref, _ := parallel.Solve(g, 0, 4)

	for run := 0; run < 5; run++ {
		res, _ := parallel.Solve(g, 0, 4)
		for v := range ref.Dist {
			if !approxEqual(ref.Dist[v], res.Dist[v], 1e-9) {
				t.Errorf("run %d: dist[%d] = %.4f, want %.4f",
					run, v, res.Dist[v], ref.Dist[v])
			}
		}
	}
}

func TestParallelUnreachableVertices(t *testing.T) {
	raw := []graph.RawEdge{{From: 0, To: 1, Weight: 3}}
	for _, P := range []int{1, 4, 8} {
		dist := mustSolveParallel(t, 4, raw, 0, P)
		if dist[2] != math.MaxFloat64 {
			t.Errorf("P=%d: dist[2] = %.4f, want +Inf", P, dist[2])
		}
		if dist[3] != math.MaxFloat64 {
			t.Errorf("P=%d: dist[3] = %.4f, want +Inf", P, dist[3])
		}
	}
}

func TestParallelHeavyEdgesAfterBucket(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 5},
		{From: 0, To: 2, Weight: 50},
		{From: 1, To: 2, Weight: 3},
	}
	for _, P := range []int{1, 4, 8} {
		dist := mustSolveParallel(t, 3, raw, 0, P)
		if !approxEqual(dist[2], 8, 1e-9) {
			t.Errorf("P=%d: dist[2] = %.4f, want 8", P, dist[2])
		}
	}
}

func TestParallelMoreThreadsThanVertices(t *testing.T) {
	raw := []graph.RawEdge{{From: 0, To: 1, Weight: 7}}
	dist := mustSolveParallel(t, 2, raw, 0, 12)

	if dist[0] != 0 {
		t.Errorf("dist[0] = %.4f, want 0", dist[0])
	}
	if !approxEqual(dist[1], 7, 1e-9) {
		t.Errorf("dist[1] = %.4f, want 7", dist[1])
	}
}

func TestParallelNoEdges(t *testing.T) {
	for _, P := range []int{1, 4, 8} {
		dist := mustSolveParallel(t, 4, []graph.RawEdge{}, 0, P)
		if dist[0] != 0 {
			t.Errorf("P=%d: dist[0] = %.4f, want 0", P, dist[0])
		}
		for v := 1; v < 4; v++ {
			if dist[v] != math.MaxFloat64 {
				t.Errorf("P=%d: dist[%d] = %.4f, want +Inf", P, v, dist[v])
			}
		}
	}
}

func TestParallelBarrierCorrectness(t *testing.T) {
	raw := []graph.RawEdge{
		{From: 0, To: 1, Weight: 1},
		{From: 1, To: 2, Weight: 1},
		{From: 2, To: 3, Weight: 1},
		{From: 3, To: 4, Weight: 1},
		{From: 0, To: 4, Weight: 100},
	}
	for _, P := range []int{1, 4, 8} {
		dist := mustSolveParallel(t, 5, raw, 0, P)
		if !approxEqual(dist[4], 4, 1e-9) {
			t.Errorf("P=%d: dist[4] = %.4f, want 4 (шлях 0->1->2->3->4)", P, dist[4])
		}
	}
}

func TestParallelInvalidParams(t *testing.T) {
	raw := []graph.RawEdge{{From: 0, To: 1, Weight: 1}}
	g, _ := graph.NewGraph(2, raw, parallelDelta)

	_, err := parallel.Solve(g, -1, 4)
	if err == nil {
		t.Error("очікується помилка для source=-1")
	}
	_, err = parallel.Solve(g, 5, 4)
	if err == nil {
		t.Error("очікується помилка для source=5 (поза межами)")
	}
	_, err = parallel.Solve(g, 0, 0)
	if err == nil {
		t.Error("очікується помилка для parallelism=0")
	}
}

func TestParallelLargeGraphNoRaceCondition(t *testing.T) {
	const n = 50_000
	rawEdges, err := generator.Connected(n, 500_000, 1, 100, 42)
	if err != nil {
		t.Fatalf("generator: %v", err)
	}
	g, err := graph.NewGraph(n, rawEdges, parallelDelta)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	ref, err := parallel.Solve(g, 0, 1)
	if err != nil {
		t.Fatalf("Solve P=1: %v", err)
	}
	res, err := parallel.Solve(g, 0, 8)
	if err != nil {
		t.Fatalf("Solve P=8: %v", err)
	}

	mismatches := 0
	for v := range ref.Dist {
		if !approxEqual(ref.Dist[v], res.Dist[v], 1e-9) {
			mismatches++
		}
	}
	if mismatches > 0 {
		t.Errorf("знайдено %d розбіжностей між P=1 та P=8 на великому графі", mismatches)
	}
}

func TestParallelLargeGraphNoOverflow(t *testing.T) {
	const n = 10_000
	rawEdges, err := generator.Connected(n, 10_000_000, 1, 100, 42)
	if err != nil {
		t.Fatalf("generator: %v", err)
	}
	g, err := graph.NewGraph(n, rawEdges, parallelDelta)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	res, err := parallel.Solve(g, 0, 8)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}

	if res.Dist[0] != 0 {
		t.Errorf("dist[0] = %.4f, want 0", res.Dist[0])
	}
	for v, d := range res.Dist {
		if d < 0 || (d != math.MaxFloat64 && d >= 1.7e+308/2) {
			t.Errorf("dist[%d] = %.4f виглядає як переповнення", v, d)
		}
	}
}
