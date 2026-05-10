package result

import "math"

const Inf = math.MaxFloat64 / 2

type Result struct {
	Dist          []float64
	ExecutionTime int64
}
