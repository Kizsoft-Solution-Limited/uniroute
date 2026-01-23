package tunnel

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// PendingRequest represents a pending HTTP request waiting for response
type PendingRequest struct {
	RequestID   string
	ResponseChan chan *HTTPResponse
	ErrorChan   chan error
	CreatedAt   time.Time
	Timeout     time.Duration
}

// RequestTracker tracks pending requests for request/response matching
type RequestTracker struct {
	pendingRequests map[string]*PendingRequest
	mu              sync.RWMutex
	logger          zerolog.Logger
	defaultTimeout  time.Duration
}

func NewRequestTracker(logger zerolog.Logger) *RequestTracker {
	return &RequestTracker{
		pendingRequests: make(map[string]*PendingRequest),
		logger:          logger,
		defaultTimeout: 120 * time.Second,
	}
}

// RegisterRequest registers a new pending request
func (rt *RequestTracker) RegisterRequest(requestID string) (*PendingRequest, error) {
	return rt.RegisterRequestWithTimeout(requestID, rt.defaultTimeout)
}

// RegisterRequestWithTimeout registers a new pending request with custom timeout
func (rt *RequestTracker) RegisterRequestWithTimeout(requestID string, timeout time.Duration) (*PendingRequest, error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	// Check if request already exists
	if _, exists := rt.pendingRequests[requestID]; exists {
		return nil, ErrDuplicateRequestID
	}

	req := &PendingRequest{
		RequestID:   requestID,
		ResponseChan: make(chan *HTTPResponse, 1),
		ErrorChan:   make(chan error, 1),
		CreatedAt:   time.Now(),
		Timeout:     timeout,
	}

	rt.pendingRequests[requestID] = req

	// Auto-cleanup after timeout
	go rt.cleanupAfterTimeout(requestID, timeout)

	return req, nil
}

// CompleteRequest completes a pending request with a response
func (rt *RequestTracker) CompleteRequest(requestID string, response *HTTPResponse) error {
	rt.mu.Lock()
	req, exists := rt.pendingRequests[requestID]
	if !exists {
		rt.mu.Unlock()
		return ErrRequestNotFound
	}
	delete(rt.pendingRequests, requestID)
	rt.mu.Unlock()

	select {
	case req.ResponseChan <- response:
		return nil
	case <-time.After(1 * time.Second):
		return ErrResponseChannelTimeout
	}
}

// FailRequest fails a pending request with an error
func (rt *RequestTracker) FailRequest(requestID string, err error) error {
	rt.mu.Lock()
	req, exists := rt.pendingRequests[requestID]
	if !exists {
		rt.mu.Unlock()
		return ErrRequestNotFound
	}
	delete(rt.pendingRequests, requestID)
	rt.mu.Unlock()

	select {
	case req.ErrorChan <- err:
		return nil
	case <-time.After(1 * time.Second):
		return ErrResponseChannelTimeout
	}
}

// GetRequest retrieves a pending request
func (rt *RequestTracker) GetRequest(requestID string) (*PendingRequest, bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	req, exists := rt.pendingRequests[requestID]
	return req, exists
}

// cleanupAfterTimeout removes a request after timeout
func (rt *RequestTracker) cleanupAfterTimeout(requestID string, timeout time.Duration) {
	<-time.After(timeout)

	rt.mu.Lock()
	req, exists := rt.pendingRequests[requestID]
	if exists {
		delete(rt.pendingRequests, requestID)
		rt.mu.Unlock()

		// Send timeout error
		select {
		case req.ErrorChan <- ErrRequestTimeout:
		default:
		}
	} else {
		rt.mu.Unlock()
	}
}

// CleanupStaleRequests removes requests older than maxAge
func (rt *RequestTracker) CleanupStaleRequests(maxAge time.Duration) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	now := time.Now()
	for id, req := range rt.pendingRequests {
		if now.Sub(req.CreatedAt) > maxAge {
			delete(rt.pendingRequests, id)
			select {
			case req.ErrorChan <- ErrRequestTimeout:
			default:
			}
		}
	}
}

// GetPendingCount returns the number of pending requests
func (rt *RequestTracker) GetPendingCount() int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return len(rt.pendingRequests)
}

// WaitForResponse waits for a response to a pending request
func (req *PendingRequest) WaitForResponse(ctx context.Context) (*HTTPResponse, error) {
	timeout := req.Timeout
	if timeout > 120*time.Second {
		timeout = 120 * time.Second
	}
	
	select {
	case response := <-req.ResponseChan:
		return response, nil
	case err := <-req.ErrorChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, ErrRequestTimeout
	}
}

