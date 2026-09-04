package worker

import (
	"math"
	"math/rand"
	"time"
)

// CalculateBackoff computes an exponential backoff with jitter.
func CalculateBackoff(attempt int, baseDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return baseDelay
	}

	// Max reasonable delay (e.g., 2 hours)
	const maxDelay = 2 * time.Hour

	// exponential growth: baseDelay * 2^attempt
	// to prevent overflow and absurdly large delays, cap the shift at a safe limit (e.g. 15)
	if attempt > 15 {
		attempt = 15
	}
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))

	// Jitter +/- 10%
	jitterAmount := float64(delay) * 0.1
	jitter := (rand.Float64() * 2 * jitterAmount) - jitterAmount

	finalDelay := time.Duration(float64(delay) + jitter)

	if finalDelay > maxDelay {
		finalDelay = maxDelay
	}

	return finalDelay
}
