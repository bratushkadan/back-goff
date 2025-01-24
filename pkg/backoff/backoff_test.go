package pkg_test

import (
	"kafka/svcs/pkg"
	"math"
	"slices"
	"testing"
	"time"
)

var (
	compEpsilonSec       = (100 * time.Microsecond).Seconds()
	jitterCompEpsilonSec = (1 * time.Nanosecond).Seconds()
)

func TestBackoffNoJitter(t *testing.T) {
	backoff := &pkg.Backoff{BaseStart: time.Second, BaseMax: 32 * time.Second, Factor: 2.0}
	backoff2 := &pkg.Backoff{BaseStart: 3 * time.Second, BaseMax: 120 * time.Second, Factor: 3.0}
	backoffExpected := []float64{1.0, 2.0, 4.0, 8.0, 16.0, 32.0, 32.0, 32.0}
	backoff2Expected := []float64{3.0, 9.0, 27.0, 81.0, 120.0, 120.0}

	populate := func(b *pkg.Backoff, expected []float64) []float64 {
		actual := make([]float64, 0, len(expected))
		for range len(expected) {
			actual = append(actual, b.GetIncr().Seconds())
		}
		return actual
	}

	backoffActual := populate(backoff, backoffExpected)
	backoff2Actual := populate(backoff2, backoff2Expected)

	compare := func(expected, actual []float64) bool {
		return slices.EqualFunc(expected, actual, func(expected, actual float64) bool {
			return math.Abs(expected-actual) < compEpsilonSec
		})
	}

	if !compare(backoffExpected, backoffActual) {
		t.Errorf("Wrong backoff function calculation results.\nExpected:\n%+v\nActual:\n%+v", backoffExpected, backoffActual)
	}
	if !compare(backoff2Expected, backoff2Actual) {
		t.Errorf("Wrong backoff function calculation results.\nExpected:\n%+v\nActual:\n%+v", backoff2Expected, backoff2Actual)
	}
}

func TestBackoffJitter(t *testing.T) {
	var (
		jitterMin = 200 * time.Millisecond
		jitterMax = 800 * time.Millisecond
	)
	var (
		factor    = 2.0
		baseStart = time.Second
		baseMax   = 32 * time.Second
	)
	backoff := &pkg.Backoff{BaseStart: baseStart, BaseMax: baseMax, Factor: factor, JitterMin: jitterMin, JitterMax: jitterMax}

	populate := func(b *pkg.Backoff, n int) []float64 {
		actual := make([]float64, 0, n)
		for range n {
			actual = append(actual, b.GetIncr().Seconds())
		}
		return actual
	}

	backoffActual := populate(backoff, 10)

	for i, prev := 0, float64(baseStart.Seconds()); i < len(backoffActual)-1; i++ {
		b := math.Min(prev, baseMax.Seconds())
		curMin := jitterMin.Seconds() + b
		curMax := jitterMax.Seconds() + b
		cur := backoffActual[i]

		if (math.Abs(curMax-cur)-jitterCompEpsilonSec < 0) || (math.Abs(cur-curMin)-jitterCompEpsilonSec < 0) {
			// TODO: log calculation steps/generated jitter values
			t.Errorf("Wrong backoff jitter function calculation result.\nExpected f(%d) = %.3f to be in [%.3f; %.3f]", i, cur, curMin, curMax)
		}

		prev = cur
	}
}
