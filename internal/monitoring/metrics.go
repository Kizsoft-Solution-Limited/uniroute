package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request metrics
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uniroute_requests_total",
			Help: "Total number of requests",
		},
		[]string{"provider", "model", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "uniroute_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "model"},
	)

	// Token metrics
	TokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uniroute_tokens_total",
			Help: "Total number of tokens processed",
		},
		[]string{"provider", "model", "type"}, // type: input, output
	)

	// Cost metrics
	CostTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uniroute_cost_total",
			Help: "Total cost in USD",
		},
		[]string{"provider", "model"},
	)

	// Provider health metrics
	ProviderHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uniroute_provider_health",
			Help: "Provider health status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"provider"},
	)

	// Rate limit metrics
	RateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uniroute_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"api_key_id", "type"}, // type: per_minute, per_day
	)
)

// RecordRequest records a request metric
func RecordRequest(provider, model, status string, duration float64) {
	RequestsTotal.WithLabelValues(provider, model, status).Inc()
	RequestDuration.WithLabelValues(provider, model).Observe(duration)
}

// RecordTokens records token usage
func RecordTokens(provider, model, tokenType string, count int) {
	TokensTotal.WithLabelValues(provider, model, tokenType).Add(float64(count))
}

// RecordCost records cost
func RecordCost(provider, model string, cost float64) {
	CostTotal.WithLabelValues(provider, model).Add(cost)
}

// SetProviderHealth sets provider health status
func SetProviderHealth(provider string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	ProviderHealth.WithLabelValues(provider).Set(value)
}

// RecordRateLimitHit records a rate limit hit
func RecordRateLimitHit(apiKeyID, limitType string) {
	RateLimitHits.WithLabelValues(apiKeyID, limitType).Inc()
}
