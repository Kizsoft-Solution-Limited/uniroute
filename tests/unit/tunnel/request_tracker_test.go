package tunnel_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
)

var (
	NewRequestTracker = tunnel.NewRequestTracker
	ErrDuplicateRequestID = tunnel.ErrDuplicateRequestID
	ErrRequestNotFound    = tunnel.ErrRequestNotFound
	ErrRequestTimeout     = tunnel.ErrRequestTimeout
)

type HTTPResponse = tunnel.HTTPResponse

func TestNewRequestTracker(t *testing.T) {
	logger := zerolog.Nop()
	tracker := NewRequestTracker(logger)

	assert.NotNil(t, tracker)
	// Note: defaultTimeout and pendingRequests are unexported, so we can't test them directly
	// We test the exported behavior instead
}

func TestRequestTracker_RegisterRequest(t *testing.T) {
	logger := zerolog.Nop()
	tracker := NewRequestTracker(logger)

	req, err := tracker.RegisterRequest("req-1")
	require.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "req-1", req.RequestID)
	assert.Equal(t, 30*time.Second, req.Timeout)

	// Test duplicate
	_, err = tracker.RegisterRequest("req-1")
	assert.Error(t, err)
	assert.Equal(t, ErrDuplicateRequestID, err)
}

func TestRequestTracker_CompleteRequest(t *testing.T) {
	logger := zerolog.Nop()
	tracker := NewRequestTracker(logger)

	req, _ := tracker.RegisterRequest("req-1")

	response := &HTTPResponse{
		RequestID: "req-1",
		Status:    200,
		Headers:   make(map[string]string),
		Body:      []byte("test"),
	}

	err := tracker.CompleteRequest("req-1", response)
	assert.NoError(t, err)

	// Verify response received
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := req.WaitForResponse(ctx)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.Status)
	assert.Equal(t, []byte("test"), resp.Body)

	// Test non-existent request
	err = tracker.CompleteRequest("non-existent", response)
	assert.Error(t, err)
	assert.Equal(t, ErrRequestNotFound, err)
}

func TestRequestTracker_FailRequest(t *testing.T) {
	logger := zerolog.Nop()
	tracker := NewRequestTracker(logger)

	req, _ := tracker.RegisterRequest("req-1")

	err := tracker.FailRequest("req-1", ErrRequestTimeout)
	assert.NoError(t, err)

	// Verify error received
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = req.WaitForResponse(ctx)
	assert.Error(t, err)
	assert.Equal(t, ErrRequestTimeout, err)
}

func TestRequestTracker_Timeout(t *testing.T) {
	logger := zerolog.Nop()
	tracker := NewRequestTracker(logger)

	req, _ := tracker.RegisterRequestWithTimeout("req-1", 100*time.Millisecond)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := req.WaitForResponse(ctx)
	assert.Error(t, err)
	assert.Equal(t, ErrRequestTimeout, err)
}

func TestRequestTracker_GetPendingCount(t *testing.T) {
	logger := zerolog.Nop()
	tracker := NewRequestTracker(logger)

	assert.Equal(t, 0, tracker.GetPendingCount())

	tracker.RegisterRequest("req-1")
	tracker.RegisterRequest("req-2")

	assert.Equal(t, 2, tracker.GetPendingCount())

	tracker.CompleteRequest("req-1", &HTTPResponse{RequestID: "req-1", Status: 200})

	assert.Equal(t, 1, tracker.GetPendingCount())
}
