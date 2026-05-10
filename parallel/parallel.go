package parallel

import (
	"Delta-stepping-algorithm/graph"
	"Delta-stepping-algorithm/result"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

func atomicLoadFloat64(addr *float64) float64 {
	return math.Float64frombits(
		atomic.LoadUint64((*uint64)(unsafe.Pointer(addr))),
	)
}

func atomicCASFloat64(addr *float64, old, new float64) bool {
	return atomic.CompareAndSwapUint64(
		(*uint64)(unsafe.Pointer(addr)),
		math.Float64bits(old),
		math.Float64bits(new),
	)
}

type localUpdate struct {
	v                int
	oldDist, newDist float64
}

type phaseTask struct {
	vertices []int
	lo, hi   int
	light    bool
}

func Solve(g *graph.Graph, source, parallelism int) (*result.Result, error) {
	if source < 0 || source >= g.NumVertices {
		return nil, fmt.Errorf(
			"parallel: source vertex %d out of range [0, %d)",
			source, g.NumVertices,
		)
	}
	if parallelism < 1 {
		return nil, fmt.Errorf(
			"parallel: parallelism must be >= 1, got %d", parallelism,
		)
	}

	startTime := time.Now()
	n := g.NumVertices
	P := parallelism

	dist := make([]float64, n)
	for i := range dist {
		dist[i] = result.Inf
	}
	dist[source] = 0

	var mu sync.Mutex
	buckets := []map[int]struct{}{}

	ensureBucket := func(idx int) {
		for len(buckets) <= idx {
			buckets = append(buckets, make(map[int]struct{}))
		}
	}
	bucketIdx := func(d float64) int { return int(d / g.Delta) }

	applyUpdates := func(upds []localUpdate) {
		if len(upds) == 0 {
			return
		}
		mu.Lock()
		for _, u := range upds {
			if u.oldDist < result.Inf {
				if idx := bucketIdx(u.oldDist); idx < len(buckets) {
					delete(buckets[idx], u.v)
				}
			}
			idx := bucketIdx(u.newDist)
			ensureBucket(idx)
			buckets[idx][u.v] = struct{}{}
		}
		mu.Unlock()
	}

	taskCh := make(chan phaseTask, P*2)
	var wg sync.WaitGroup

	for w := 0; w < P; w++ {
		go func() {
			upds := make([]localUpdate, 0, 512)

			for t := range taskCh {
				upds = upds[:0]

				for k := t.lo; k < t.hi; k++ {
					u := t.vertices[k]
					du := atomicLoadFloat64(&dist[u])
					if du >= result.Inf {
						continue
					}

					var edges []graph.Edge
					if t.light {
						edges = g.Light[u]
					} else {
						edges = g.Heavy[u]
					}

					for _, e := range edges {
						v := e.To
						newDist := du + e.Weight
						for {
							cur := atomicLoadFloat64(&dist[v])
							if newDist >= cur {
								break
							}
							if atomicCASFloat64(&dist[v], cur, newDist) {
								upds = append(upds, localUpdate{
									v: v, oldDist: cur, newDist: newDist,
								})
								break
							}
						}
					}
				}
				applyUpdates(upds)
				wg.Done()
			}
		}()
	}

	dispatch := func(vertices []int, light bool) {
		total := len(vertices)
		if total == 0 {
			return
		}
		numW := P
		if numW > total {
			numW = total
		}
		S := total / numW
		R := total % numW

		wg.Add(numW)
		for t := 0; t < numW; t++ {
			lo := t*S + min(t, R)
			hi := lo + S
			if t < R {
				hi++
			}
			taskCh <- phaseTask{vertices: vertices, lo: lo, hi: hi, light: light}
		}
		wg.Wait()
	}

	ensureBucket(0)
	buckets[0][source] = struct{}{}

	for i := 0; i < len(buckets); i++ {
		if len(buckets[i]) == 0 {
			continue
		}

		visited := make([]int, 0)

		for len(buckets[i]) > 0 {
			current := make([]int, 0, len(buckets[i]))
			for v := range buckets[i] {
				current = append(current, v)
			}
			buckets[i] = make(map[int]struct{})
			visited = append(visited, current...)
			dispatch(current, true)
		}

		dispatch(visited, false)
	}

	close(taskCh)

	for i, d := range dist {
		if d >= result.Inf {
			dist[i] = math.MaxFloat64
		}
	}

	return &result.Result{
		Dist:          dist,
		ExecutionTime: time.Since(startTime).Milliseconds(),
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
