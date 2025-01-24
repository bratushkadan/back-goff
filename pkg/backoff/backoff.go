package pkg

import (
	"math"
	"math/rand/v2"
	"time"
)

// Backoff calculation utility for retries.
type Backoff struct {
	attempts int
	// Base start duration - backoff can be no shorter than the BaseStart.Add(JitterMin).
	BaseStart time.Duration
	// Base max duration - backoff can be no longer than the BaseMax.Add(JitterMax).
	BaseMax time.Duration
	// Backoff exponential growth factor.
	Factor float64
	// Min random duration to add to the backoff value to prevent potential
	// upstream overloading during the simultaneous retry among clients implementing backoff.
	// Must be less than JitterMax.
	JitterMin time.Duration
	// Max random duration to add to the backoff value to prevent potential
	// upstream overloading during the simultaneous retry among clients implementing backoff.
	// Must be greater than JitterMin.
	JitterMax time.Duration
}

func (b *Backoff) GetRetries() int {
	return b.attempts
}

func (b *Backoff) Reset() {
	b.attempts = 0
}

func (b *Backoff) Incr() int {
	b.attempts++
	return b.GetRetries()
}

func (b *Backoff) Get() time.Duration {
	backoff := float64(b.BaseStart) * math.Pow(b.Factor, float64(b.attempts))
	if backoff > float64(b.BaseMax) {
		backoff = float64(b.BaseMax)
	}

	if b.JitterMax != 0 {
		jitter := ((float64(b.JitterMax) - float64(b.JitterMin)) * rand.Float64()) + float64(b.JitterMin)
		backoff += jitter
	}

	return time.Duration(backoff)
}

func (b *Backoff) GetIncr() time.Duration {
	defer b.Incr()
	return b.Get()
}
