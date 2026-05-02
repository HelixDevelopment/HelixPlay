package network

import (
	"testing"
)

func TestEstimatorRecordAndQuery(t *testing.T) {
	e := NewEstimator(10)
	e.Record(20, 15000, 0.1)
	e.Record(22, 16000, 0.2)
	e.Record(25, 14000, 0.15)

	rtt := e.RTT()
	if rtt <= 0 {
		t.Fatalf("expected positive RTT, got %f", rtt)
	}

	bw := e.Bandwidth()
	if bw <= 0 {
		t.Fatalf("expected positive bandwidth, got %f", bw)
	}

	loss := e.Loss()
	if loss < 0 {
		t.Fatalf("expected non-negative loss, got %f", loss)
	}

	jitter := e.Jitter()
	if jitter < 0 {
		t.Fatalf("expected non-negative jitter, got %f", jitter)
	}
}

func TestEstimatorWindowCapacity(t *testing.T) {
	e := NewEstimator(3)
	e.Record(10, 1000, 0)
	e.Record(20, 2000, 0)
	e.Record(30, 3000, 0)
	e.Record(40, 4000, 0)

	// Window should only keep last 3
	if len(e.window) != 3 {
		t.Fatalf("expected window size 3, got %d", len(e.window))
	}
}

func TestEstimatorRecommendation(t *testing.T) {
	e := NewEstimator(10)

	// Good network
	e.Record(20, 50000, 0.1)
	if rec := e.Recommendation(); rec != "good" {
		t.Errorf("expected good, got %s", rec)
	}

	// Fair network
	e.Record(60, 30000, 1.0)
	if rec := e.Recommendation(); rec != "fair" {
		t.Errorf("expected fair, got %s", rec)
	}

	// Poor network
	e.Record(150, 10000, 5.0)
	if rec := e.Recommendation(); rec != "poor" {
		t.Errorf("expected poor, got %s", rec)
	}
}

func TestEstimatorStable(t *testing.T) {
	e := NewEstimator(10)
	if e.Stable() {
		t.Error("empty estimator should not be stable")
	}

	e.Record(20, 50000, 0.1)
	e.Record(21, 50000, 0.1)
	if !e.Stable() {
		t.Error("expected stable for low jitter and loss")
	}

	e.Record(100, 50000, 5.0)
	if e.Stable() {
		t.Error("expected unstable for high jitter")
	}
}

func TestPercentile(t *testing.T) {
	vals := []float64{1, 2, 3, 4, 5}
	if p := percentile(vals, 0.5); p != 3 {
		t.Errorf("expected p50=3, got %f", p)
	}
	if p := percentile(vals, 0.9); p != 5 {
		t.Errorf("expected p90=5, got %f", p)
	}
}

func TestStdDev(t *testing.T) {
	vals := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	// Mean = 5, variance = 4, stddev = 2
	if s := stdDev(vals); s < 1.9 || s > 2.1 {
		t.Errorf("expected stddev ~2, got %f", s)
	}
}
