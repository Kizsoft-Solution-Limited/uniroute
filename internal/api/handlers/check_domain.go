package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)


type CheckDomainHandler struct {
	domainRepo *storage.CustomDomainRepository
	tunnelRepo *tunnel.TunnelRepository
	logger     zerolog.Logger
}

func NewCheckDomainHandler(domainRepo *storage.CustomDomainRepository, tunnelRepo *tunnel.TunnelRepository, logger zerolog.Logger) *CheckDomainHandler {
	return &CheckDomainHandler{
		domainRepo: domainRepo,
		tunnelRepo: tunnelRepo,
		logger:     logger,
	}
}

func (h *CheckDomainHandler) HandleCheckDomain(c *gin.Context) {
	domain := strings.TrimSpace(c.Query("domain"))
	if domain == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// Caddy on-demand TLS calls this before issuing a certificate. Allowing every
	// *.tunnel.uniroute.co hostname caused bots to trigger thousands of LE orders and
	// hit the 50 certs/week limit for uniroute.co — breaking legitimate subdomains.
	const tunnelHostSuffix = ".tunnel.uniroute.co"
	normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
	if strings.HasSuffix(normalized, tunnelHostSuffix) {
		sub := strings.TrimSuffix(normalized, tunnelHostSuffix)
		if sub == "" || strings.Contains(sub, ".") {
			h.logger.Debug().Str("domain", domain).Msg("check-domain: denied (multi-label tunnel hostname)")
			c.Status(http.StatusNotFound)
			return
		}
		if h.tunnelRepo == nil {
			h.logger.Warn().Str("domain", domain).Msg("check-domain: tunnel repository unavailable")
			c.Status(http.StatusInternalServerError)
			return
		}
		_, err := h.tunnelRepo.GetTunnelBySubdomain(c.Request.Context(), sub)
		if err != nil {
			if errors.Is(err, tunnel.ErrTunnelNotFound) {
				h.logger.Debug().Str("domain", domain).Str("subdomain", sub).Msg("check-domain: denied (unknown tunnel subdomain)")
				c.Status(http.StatusNotFound)
				return
			}
			h.logger.Warn().Err(err).Str("domain", domain).Msg("check-domain: tunnel lookup failed")
			c.Status(http.StatusInternalServerError)
			return
		}
		h.logger.Debug().Str("domain", domain).Msg("check-domain: allowed (registered tunnel subdomain)")
		c.Status(http.StatusOK)
		return
	}

	if h.domainRepo != nil {
		exists, err := h.domainRepo.ExistsByDomain(c.Request.Context(), domain)
		if err != nil {
			h.logger.Warn().Err(err).Str("domain", domain).Msg("check-domain: custom_domains lookup failed")
			c.Status(http.StatusInternalServerError)
			return
		}
		
		if exists {
			h.logger.Debug().Str("domain", domain).Msg("check-domain: allowed (in custom_domains)")
			c.Status(http.StatusOK)
			return
		}
	}

	if h.tunnelRepo != nil {
		_, err := h.tunnelRepo.GetTunnelByCustomDomain(c.Request.Context(), domain)
		if err == nil {
			h.logger.Debug().Str("domain", domain).Msg("check-domain: allowed (tunnel custom_domain)")
			c.Status(http.StatusOK)
			return
		}
	}

	h.logger.Debug().Str("domain", domain).Msg("check-domain: denied")
	c.Status(http.StatusNotFound)
}
