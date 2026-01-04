package gateway

import (
	"testing"
	"time"
)

func TestNewLatencyTracker(t *testing.T) {
	tracker := NewLatencyTracker(100)
	if tracker == nil {
		t.Fatal("NewLatencyTracker returned nil")
	}
}

func TestLatencyTracker_RecordLatency(t *testing.T) {
	tracker := NewLatencyTracker(100)
	
	tracker.RecordLatency("provider1", 100*time.Millisecond)
	tracker.RecordLatency("provider1", 200*time.Millisecond)
	
	avg := tracker.GetAverageLatency("provider1")
	expected := 150 * time.Millisecond
	if avg != expected {
		t.Errorf("Expected average latency %v, got %v", expected, avg)
	}
}

func TestLatencyTracker_GetAverageLatency(t *testing.T) {
	tracker := NewLatencyTracker(100)
	
	// No data should return default
	avg := tracker.GetAverageLatency("nonexistent")
	if avg != 1*time.Second {
		t.Errorf("Expected default latency 1s, got %v", avg)
	}
	
	// With data
	tracker.RecordLatency("provider1", 50*time.Millisecond)
	tracker.RecordLatency("provider1", 150*time.Millisecond)
	
	avg = tracker.GetAverageLatency("provider1")
	expected := 100 * time.Millisecond
	if avg != expected {
		t.Errorf("Expected average latency %v, got %v", expected, avg)
	}
}

func TestLatencyTracker_GetLatencyStats(t *testing.T) {
	tracker := NewLatencyTracker(100)
	
	tracker.RecordLatency("provider1", 50*time.Millisecond)
	tracker.RecordLatency("provider1", 100*time.Millisecond)
	tracker.RecordLatency("provider1", 150*time.Millisecond)
	
	avg, min, max, count := tracker.GetLatencyStats("provider1")
	
	if avg != 100*time.Millisecond {
		t.Errorf("Expected avg %v, got %v", 100*time.Millisecond, avg)
	}
	if min != 50*time.Millisecond {
		t.Errorf("Expected min %v, got %v", 50*time.Millisecond, min)
	}
	if max != 150*time.Millisecond {
		t.Errorf("Expected max %v, got %v", 150*time.Millisecond, max)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestLatencyTracker_MaxSamples(t *testing.T) {
	tracker := NewLatencyTracker(3) // Keep only 3 samples
	
	// Record more than max
	for i := 0; i < 5; i++ {
		tracker.RecordLatency("provider1", time.Duration(i+1)*100*time.Millisecond)
	}
	
	_, _, _, count := tracker.GetLatencyStats("provider1")
	if count != 3 {
		t.Errorf("Expected max 3 samples, got %d", count)
	}
}

