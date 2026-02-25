package mcp

import (
	"context"
	"sync"
)

type Service struct {
	serverURLs []string
	clients    map[string]*Client
	mu         sync.RWMutex
}

func NewService(serverURLs []string) *Service {
	urls := make([]string, 0, len(serverURLs))
	seen := make(map[string]bool)
	for _, u := range serverURLs {
		if u != "" && !seen[u] && ValidateServerURL(u) == nil {
			seen[u] = true
			urls = append(urls, u)
		}
	}
	return &Service{
		serverURLs: urls,
		clients:    make(map[string]*Client),
	}
}

func (s *Service) client(serverURL string) *Client {
	s.mu.RLock()
	c := s.clients[serverURL]
	s.mu.RUnlock()
	if c != nil {
		return c
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if c = s.clients[serverURL]; c == nil {
		c = NewClient(serverURL)
		s.clients[serverURL] = c
	}
	return c
}

func (s *Service) ServerURLs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.serverURLs...)
}

func (s *Service) ListTools(ctx context.Context, serverURL string) ([]Tool, error) {
	if err := ValidateServerURL(serverURL); err != nil {
		return nil, err
	}
	client := s.client(serverURL)
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	return client.ListTools(ctx)
}

func (s *Service) CallTool(ctx context.Context, serverURL, name string, arguments map[string]interface{}) ([]byte, error) {
	if err := ValidateServerURL(serverURL); err != nil {
		return nil, err
	}
	client := s.client(serverURL)
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	result, err := client.CallTool(ctx, name, arguments)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ListAllTools(ctx context.Context) (map[string][]Tool, error) {
	out := make(map[string][]Tool)
	for _, url := range s.ServerURLs() {
		tools, err := s.ListTools(ctx, url)
		if err != nil {
			out[url] = nil
			continue
		}
		out[url] = tools
	}
	return out, nil
}
