package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// DomainHandler handles custom domain management requests
type DomainHandler struct {
	repository   *storage.CustomDomainRepository
	domainManager *tunnel.DomainManager
	logger       zerolog.Logger
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(repository *storage.CustomDomainRepository, logger zerolog.Logger) *DomainHandler {
	return &DomainHandler{
		repository: repository,
		logger:     logger,
	}
}

// SetDomainManager sets the domain manager for DNS validation
func (h *DomainHandler) SetDomainManager(manager *tunnel.DomainManager) {
	h.domainManager = manager
}

// ListDomains handles GET /auth/domains
func (h *DomainHandler) ListDomains(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	domains, err := h.repository.ListDomainsByUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list domains")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list domains",
		})
		return
	}

	domainList := make([]map[string]interface{}, len(domains))
	for i, d := range domains {
		domainList[i] = map[string]interface{}{
			"id":              d.ID.String(),
			"domain":          d.Domain,
			"verified":       d.Verified,
			"dns_configured":  d.DNSConfigured,
			"created_at":      d.CreatedAt,
			"updated_at":      d.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domainList,
	})
}

// CreateDomain handles POST /auth/domains
func (h *DomainHandler) CreateDomain(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	var req struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Trim and validate domain
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "domain cannot be empty",
		})
		return
	}

	// Basic validation
	if !strings.Contains(domain, ".") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid domain format",
		})
		return
	}

	// Create domain in database (initially unverified and DNS not configured)
	domainObj, err := h.repository.CreateDomain(c.Request.Context(), userID, domain)
	if err != nil {
		if err == storage.ErrDomainAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "domain already exists",
			})
			return
		}
		h.logger.Error().Err(err).Msg("Failed to create domain")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create domain",
		})
		return
	}

	// Return response immediately - DNS validation happens asynchronously
	// This prevents timeouts when DNS lookups are slow
	c.JSON(http.StatusOK, gin.H{
		"message": "domain created successfully",
		"domain": map[string]interface{}{
			"id":             domainObj.ID.String(),
			"domain":         domainObj.Domain,
			"verified":       domainObj.Verified,
			"dns_configured":  domainObj.DNSConfigured,
			"created_at":     domainObj.CreatedAt,
			"updated_at":     domainObj.UpdatedAt,
		},
		"dns_instructions": map[string]string{
			"type":    "CNAME",
			"name":    domain,
			"target":  "tunnel.uniroute.co",
			"message": "Add this CNAME record in your DNS provider to complete setup",
		},
	})

	// Check DNS configuration asynchronously (don't block response)
	// Users can add domain first, then configure DNS later
	if h.domainManager != nil {
		go func() {
			// Use background context with timeout for DNS check
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			configured, err := h.domainManager.ValidateCNAME(ctx, domain, "tunnel.uniroute.co")
			if err == nil && configured {
				// Update domain to mark DNS as configured
				dnsConfigured := true
				if updateErr := h.repository.UpdateDomain(ctx, domainObj.ID, nil, &dnsConfigured); updateErr != nil {
					h.logger.Warn().
						Err(updateErr).
						Str("domain", domain).
						Msg("Failed to update domain DNS status")
				} else {
					h.logger.Info().
						Str("domain", domain).
						Msg("Domain DNS automatically verified")
				}
			} else if err != nil {
				// Log DNS check error but don't fail (domain might not be configured yet)
				h.logger.Debug().
					Err(err).
					Str("domain", domain).
					Msg("DNS not yet configured (this is expected for new domains)")
			}
		}()
	}
}

// DeleteDomain handles DELETE /auth/domains/:id
func (h *DomainHandler) DeleteDomain(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	domainIDStr := c.Param("id")
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid domain ID",
		})
		return
	}

	err = h.repository.DeleteDomain(c.Request.Context(), domainID, userID)
	if err != nil {
		if err == storage.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "domain not found",
			})
			return
		}
		h.logger.Error().Err(err).Msg("Failed to delete domain")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete domain",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "domain deleted successfully",
	})
}

// VerifyDomain handles POST /auth/domains/:id/verify
// Checks if DNS (CNAME) is properly configured for the domain
func (h *DomainHandler) VerifyDomain(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	domainIDStr := c.Param("id")
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid domain ID",
		})
		return
	}

	// Get domain to verify ownership
	domainObj, err := h.repository.GetDomainByID(c.Request.Context(), domainID)
	if err != nil {
		if err == storage.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "domain not found",
			})
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get domain")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get domain",
		})
		return
	}

	// Verify domain belongs to user
	if domainObj.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	// Check CNAME configuration
	dnsConfigured := false
	var dnsError string
	if h.domainManager != nil {
		ctx := c.Request.Context()
		configured, err := h.domainManager.ValidateCNAME(ctx, domainObj.Domain, "tunnel.uniroute.co")
		if err != nil {
			dnsError = err.Error()
		} else if configured {
			dnsConfigured = true
			// Update domain to mark DNS as configured
			_ = h.repository.UpdateDomain(ctx, domainID, nil, &dnsConfigured)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"domain":          domainObj.Domain,
		"dns_configured":  dnsConfigured,
		"dns_error":       dnsError,
		"dns_instructions": map[string]string{
			"type":    "CNAME",
			"name":    domainObj.Domain,
			"target":  "tunnel.uniroute.co",
			"message": "Add this CNAME record in your DNS provider",
		},
	})
}
