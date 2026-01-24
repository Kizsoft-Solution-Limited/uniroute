package tunnel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// DomainManager handles domain and subdomain management
type DomainManager struct {
	baseDomain     string
	subdomainPool  *SubdomainPool
	logger         zerolog.Logger
	dnsValidator   *DNSValidator
}

// NewDomainManager creates a new domain manager
func NewDomainManager(baseDomain string, logger zerolog.Logger) *DomainManager {
	return &DomainManager{
		baseDomain:    baseDomain,
		subdomainPool: NewSubdomainPool(logger),
		logger:        logger,
		dnsValidator:  NewDNSValidator(logger),
	}
}

// AllocateSubdomain allocates a unique subdomain
// Note: It may return a subdomain that already exists in the database
// The CreateTunnel function uses UPSERT to handle this case and update the LocalURL
func (dm *DomainManager) AllocateSubdomain(ctx context.Context, repository *TunnelRepository) (string, error) {
	// Try to allocate from pool
	subdomain := dm.subdomainPool.Allocate()
	
	// Note: We don't check if subdomain already exists here
	// If it does, CreateTunnel will use UPSERT to update the existing tunnel's LocalURL
	// This allows reusing subdomains with different ports
	dm.logger.Debug().Str("subdomain", subdomain).Msg("Allocated subdomain from pool")
	
	return subdomain, nil
}

// ValidateCustomDomain validates a custom domain
func (dm *DomainManager) ValidateCustomDomain(ctx context.Context, domain string) error {
	// Basic validation
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	
	// Check format
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("invalid domain format")
	}
	
	// Check DNS resolution
	ips, err := net.LookupIP(domain)
	if err != nil {
		return fmt.Errorf("domain does not resolve: %w", err)
	}
	
	if len(ips) == 0 {
		return fmt.Errorf("domain has no DNS records")
	}
	
	dm.logger.Info().
		Str("domain", domain).
		Int("ip_count", len(ips)).
		Msg("Custom domain validated")
	
	return nil
}

// ValidateCNAME checks if a domain has the correct CNAME record pointing to the target
func (dm *DomainManager) ValidateCNAME(ctx context.Context, domain, target string) (bool, error) {
	if dm.dnsValidator == nil {
		return false, fmt.Errorf("DNS validator not initialized")
	}
	return dm.dnsValidator.ValidateCNAMERecord(ctx, domain, target)
}

// GetPublicURL generates the public URL for a subdomain
func (dm *DomainManager) GetPublicURL(subdomain string, port int, useHTTPS bool) string {
	scheme := "http"
	if useHTTPS {
		scheme = "https"
	}
	
	if dm.baseDomain != "" {
		return fmt.Sprintf("%s://%s.%s", scheme, subdomain, dm.baseDomain)
	}
	
	localhostDomain := os.Getenv("TUNNEL_LOCALHOST_DOMAIN")
	if localhostDomain == "" {
		localhostDomain = "localhost"
	}
	return fmt.Sprintf("%s://%s.%s:%d", scheme, subdomain, localhostDomain, port)
}

// CheckSubdomainAvailability checks if a subdomain is available
func (dm *DomainManager) CheckSubdomainAvailability(ctx context.Context, repository *TunnelRepository, subdomain string) (bool, error) {
	if repository == nil {
		return true, nil // No repository, assume available
	}
	
	existing, err := repository.GetTunnelBySubdomain(ctx, subdomain)
	if err != nil {
		// Error might mean not found, which is good
		return true, nil
	}
	
	return existing == nil, nil
}

// SubdomainPool manages a pool of available subdomains
type SubdomainPool struct {
	used   map[string]bool
	mu     sync.RWMutex
	logger zerolog.Logger
}

// NewSubdomainPool creates a new subdomain pool
func NewSubdomainPool(logger zerolog.Logger) *SubdomainPool {
	return &SubdomainPool{
		used:   make(map[string]bool),
		logger: logger,
	}
}

// Allocate allocates a new subdomain from the pool
func (sp *SubdomainPool) Allocate() string {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	
	// Generate random subdomain
	subdomain := generateRandomSubdomain()
	
	// Check if already used
	if sp.used[subdomain] {
		// Retry (with max attempts to avoid infinite loop)
		for i := 0; i < 10; i++ {
			subdomain = generateRandomSubdomain()
			if !sp.used[subdomain] {
				break
			}
		}
	}
	
	sp.used[subdomain] = true
	return subdomain
}

// Release releases a subdomain back to the pool
func (sp *SubdomainPool) Release(subdomain string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	delete(sp.used, subdomain)
}

// generateRandomSubdomain generates a random subdomain
func generateRandomSubdomain() string {
	// Use shorter subdomain for better UX
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:8] // 8 character subdomain
}

// DNSValidator validates DNS records
type DNSValidator struct {
	logger zerolog.Logger
}

// NewDNSValidator creates a new DNS validator
func NewDNSValidator(logger zerolog.Logger) *DNSValidator {
	return &DNSValidator{
		logger: logger,
	}
}

// ValidateTXTRecord validates a TXT record for domain ownership
func (dv *DNSValidator) ValidateTXTRecord(ctx context.Context, domain, expectedValue string) (bool, error) {
	records, err := net.LookupTXT(domain)
	if err != nil {
		return false, fmt.Errorf("failed to lookup TXT record: %w", err)
	}
	
	for _, record := range records {
		if record == expectedValue {
			return true, nil
		}
	}
	
	return false, nil
}

// ValidateCNAMERecord validates a CNAME record
func (dv *DNSValidator) ValidateCNAMERecord(ctx context.Context, domain, expectedTarget string) (bool, error) {
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		return false, fmt.Errorf("failed to lookup CNAME: %w", err)
	}
	
	// Remove trailing dot
	cname = strings.TrimSuffix(cname, ".")
	expectedTarget = strings.TrimSuffix(expectedTarget, ".")
	
	return cname == expectedTarget, nil
}

// WaitForDNSPropagation waits for DNS changes to propagate
func (dv *DNSValidator) WaitForDNSPropagation(ctx context.Context, domain string, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, err := net.LookupIP(domain)
			if err == nil {
				return nil // DNS propagated
			}
			
			if time.Now().After(deadline) {
				return fmt.Errorf("DNS propagation timeout after %v", maxWait)
			}
		}
	}
}

