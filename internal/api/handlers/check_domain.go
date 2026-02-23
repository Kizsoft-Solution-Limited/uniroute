package handlers

import (
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

	if strings.HasSuffix(domain, ".tunnel.uniroute.co") {
		h.logger.Debug().Str("domain", domain).Msg("check-domain: allowed (tunnel subdomain)")
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
