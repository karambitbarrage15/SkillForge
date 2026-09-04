package worker

import (
	"testing"
	"time"
)

func TestCalculateBackoff(t *testing.T) {
	baseDelay := 1 * time.Second

	// Test deterministic exponential growth roughly (ignoring minor jitter effects on bounds)
	// For attempt=0, delay = 1s
	d0 := CalculateBackoff(0, baseDelay)
	if d0 < 900*time.Millisecond || d0 > 1100*time.Millisecond {
		t.Errorf("expected ~1s, got %v", d0)
	}

	// For attempt=1, delay = 2s
	d1 := CalculateBackoff(1, baseDelay)
	if d1 < 1800*time.Millisecond || d1 > 2200*time.Millisecond {
		t.Errorf("expected ~2s, got %v", d1)
	}

	// For attempt=3, delay = 8s
	d3 := CalculateBackoff(3, baseDelay)
	if d3 < 7200*time.Millisecond || d3 > 8800*time.Millisecond {
		t.Errorf("expected ~8s, got %v", d3)
	}

	// Test max bound
	dMax := CalculateBackoff(20, baseDelay) // 2^20 seconds is very large
	if dMax > 2*time.Hour {
		t.Errorf("expected max delay to be capped at 2h, got %v", dMax)
	}

	// Test negative attempts just in case
	dNeg := CalculateBackoff(-1, baseDelay)
	if dNeg != baseDelay {
		t.Errorf("expected baseDelay for negative attempts, got %v", dNeg)
	}
}

// Test jitter randomness
func TestCalculateBackoff_Jitter(t *testing.T) {
	baseDelay := 10 * time.Second
	attempt := 2 // 40s base

	seen := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		d := CalculateBackoff(attempt, baseDelay)

		// Ensure within bounds: 40s +/- 10% = 36s to 44s
		if d < 36*time.Second || d > 44*time.Second {
			t.Errorf("jitter outside bounds: %v", d)
		}
		seen[d] = true
	}

	// Ensure we got multiple different values (randomness)
	if len(seen) < 10 {
		t.Errorf("expected multiple jitter values, got %d", len(seen))
	}
}
