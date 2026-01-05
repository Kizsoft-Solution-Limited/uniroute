# 💰 UniRoute Pricing Model

## Overview

UniRoute offers flexible pricing options to accommodate different user needs. This document outlines the pricing structure for the **Managed Service** where UniRoute provides and manages AI provider API keys.

---

## 🎯 Pricing Philosophy

### Core Principle
**UniRoute charges users only for AI model usage when using UniRoute's managed service** (because UniRoute pays providers on your behalf). All other features are completely free:
- ✅ Tunneling - Free
- ✅ Webhook Testing - Free
- ✅ Analytics & Monitoring - Free
- ✅ Routing & Load Balancing - Free
- ✅ Security Features - Free
- ✅ API Management - Free

### Pricing Transparency
- UniRoute pricing is based on **provider costs + operational overhead + profit margin**
- Users can see exactly what they're paying for
- No hidden fees or surprise charges
- Clear cost breakdown per request

---

## 💳 Pricing Structure

### Model 1: Pay-as-You-Go (Current)

**How it works:**
- Users pay only for AI model usage
- Pricing = Provider Cost × (1 + Margin)
- Billed per request based on actual token usage
- No monthly minimums, no setup fees

**Pricing Formula:**
```
User Price = Provider Cost × (1 + Margin) × (1 + Overhead)
```

Where:
- **Provider Cost**: What UniRoute pays OpenAI, Anthropic, Google, etc.
- **Margin**: Profit margin (e.g., 10-20%)
- **Overhead**: Operational costs (infrastructure, support, etc.) - typically 5-10%

**Example Calculation:**
```
GPT-4 Request:
- Provider Cost: $0.03 per 1M input tokens
- Margin: 15%
- Overhead: 5%
- User Price: $0.03 × 1.15 × 1.05 = $0.0362 per 1M input tokens
```

---

## 📊 Provider Cost Reference (2024)

### OpenAI Pricing (per 1M tokens)

| Model | Input | Output | UniRoute Price (15% margin) |
|-------|-------|--------|----------------------------|
| GPT-4 | $30.00 | $60.00 | $36.23 / $72.45 |
| GPT-4 Turbo | $10.00 | $30.00 | $12.08 / $36.23 |
| GPT-3.5 Turbo | $0.50 | $1.50 | $0.60 / $1.81 |

### Anthropic Pricing (per 1M tokens)

| Model | Input | Output | UniRoute Price (15% margin) |
|-------|-------|--------|----------------------------|
| Claude 3.5 Sonnet | $3.00 | $15.00 | $3.62 / $18.11 |
| Claude 3 Opus | $15.00 | $75.00 | $18.11 / $90.56 |
| Claude 3 Haiku | $0.25 | $1.25 | $0.30 / $1.51 |

### Google Pricing (per 1M tokens)

| Model | Input | Output | UniRoute Price (15% margin) |
|-------|-------|--------|----------------------------|
| Gemini 1.5 Pro | $1.25 | $5.00 | $1.51 / $6.04 |
| Gemini 1.5 Flash | $0.075 | $0.30 | $0.09 / $0.36 |
| Gemini Pro | $0.00 | $0.00 | $0.00 (Free tier) |

### Local Models (Ollama, vLLM)

| Model | Input | Output | UniRoute Price |
|-------|-------|--------|----------------|
| All Local Models | $0.00 | $0.00 | **FREE** |

**Note:** Local models are always free - no charges for using local LLMs.

---

## 📦 Subscription Plans (Future Implementation)

### Plan Structure

UniRoute will offer subscription plans to provide better value for high-volume users. Plans will include:

1. **Free Tier** (BYOK or Self-Hosted)
   - Unlimited requests
   - All features free
   - You pay providers directly
   - No UniRoute charges

2. **Starter Plan** (Managed Service)
   - Pay-as-you-go pricing
   - No monthly minimum
   - All features included
   - Best for: Low to medium usage

3. **Professional Plan** (Managed Service)
   - Discounted rates (10-15% off pay-as-you-go)
   - Monthly commitment: $100+
   - Priority support
   - Advanced analytics
   - Best for: Medium to high usage

4. **Enterprise Plan** (Managed Service)
   - Custom pricing (volume discounts)
   - Monthly commitment: $1000+
   - Dedicated support
   - SLA guarantees
   - Custom features
   - Best for: High-volume, enterprise needs

### Subscription Benefits

**Volume Discounts:**
- 0-100K tokens/month: Standard pricing
- 100K-1M tokens/month: 5% discount
- 1M-10M tokens/month: 10% discount
- 10M+ tokens/month: 15% discount

**Plan Features:**

| Feature | Free (BYOK) | Starter | Professional | Enterprise |
|---------|-------------|---------|--------------|-----------|
| **Pricing Model** | Free | Pay-as-you-go | Discounted | Custom |
| **Monthly Minimum** | $0 | $0 | $100 | $1000 |
| **Volume Discount** | N/A | No | Yes (10-15%) | Yes (15-25%) |
| **Support** | Community | Email | Priority | Dedicated |
| **SLA** | Best effort | 99% | 99.9% | 99.99% |
| **Custom Features** | No | No | Limited | Yes |
| **Analytics** | Basic | Basic | Advanced | Custom |

---

## 💵 Cost Calculation Examples

### Example 1: Small Project (Starter Plan)

**Usage:**
- 50,000 tokens/month
- Mix: 70% GPT-3.5 Turbo, 30% GPT-4

**Calculation:**
```
GPT-3.5 Turbo (35,000 tokens):
- Input: 20,000 tokens × $0.60/1M = $0.012
- Output: 15,000 tokens × $1.81/1M = $0.027
- Subtotal: $0.039

GPT-4 (15,000 tokens):
- Input: 10,000 tokens × $36.23/1M = $0.362
- Output: 5,000 tokens × $72.45/1M = $0.362
- Subtotal: $0.724

Total: $0.763/month
```

### Example 2: Medium Project (Professional Plan)

**Usage:**
- 5,000,000 tokens/month
- Mix: 50% GPT-4 Turbo, 30% Claude 3.5 Sonnet, 20% GPT-3.5 Turbo
- Professional Plan: 10% discount

**Calculation:**
```
GPT-4 Turbo (2,500,000 tokens):
- Input: 1,500,000 tokens × $12.08/1M = $18.12
- Output: 1,000,000 tokens × $36.23/1M = $36.23
- Subtotal: $54.35

Claude 3.5 Sonnet (1,500,000 tokens):
- Input: 900,000 tokens × $3.62/1M = $3.26
- Output: 600,000 tokens × $18.11/1M = $10.87
- Subtotal: $14.13

GPT-3.5 Turbo (1,000,000 tokens):
- Input: 600,000 tokens × $0.60/1M = $0.36
- Output: 400,000 tokens × $1.81/1M = $0.72
- Subtotal: $1.08

Subtotal: $69.56
Discount (10%): -$6.96
Total: $62.60/month
```

### Example 3: Enterprise Project (Enterprise Plan)

**Usage:**
- 50,000,000 tokens/month
- Mix: 40% GPT-4, 40% Claude 3.5 Sonnet, 20% GPT-4 Turbo
- Enterprise Plan: 20% discount + custom pricing

**Calculation:**
```
GPT-4 (20,000,000 tokens):
- Input: 12,000,000 tokens × $36.23/1M = $434.76
- Output: 8,000,000 tokens × $72.45/1M = $579.60
- Subtotal: $1,014.36

Claude 3.5 Sonnet (20,000,000 tokens):
- Input: 12,000,000 tokens × $3.62/1M = $43.44
- Output: 8,000,000 tokens × $18.11/1M = $144.88
- Subtotal: $188.32

GPT-4 Turbo (10,000,000 tokens):
- Input: 6,000,000 tokens × $12.08/1M = $72.48
- Output: 4,000,000 tokens × $36.23/1M = $144.92
- Subtotal: $217.40

Subtotal: $1,420.08
Discount (20%): -$284.02
Total: $1,136.06/month
```

---

## 🔄 Pricing Updates

### How Pricing Changes

1. **Provider Cost Changes**
   - UniRoute will update pricing when providers change their rates
   - Users will be notified 30 days in advance
   - Existing subscriptions honored at current rates for billing period

2. **Margin Adjustments**
   - Margin may be adjusted based on operational costs
   - Changes will be transparent and communicated
   - Volume discounts may increase with higher usage

3. **New Provider Pricing**
   - New providers added with competitive pricing
   - Pricing based on provider costs + standard margin

---

## 📈 Cost Optimization Tips

### For Users

1. **Use Local Models When Possible**
   - Local models (Ollama, vLLM) are always free
   - Use for development, testing, and privacy-sensitive tasks

2. **Choose the Right Model**
   - Use GPT-3.5 Turbo for simple tasks (cheaper)
   - Use GPT-4 only when needed (better quality, higher cost)

3. **Optimize Token Usage**
   - Shorter prompts = lower costs
   - Use streaming for better UX without extra cost

4. **Monitor Usage**
   - Use analytics to track costs
   - Set up alerts for budget limits

5. **Consider BYOK**
   - If you have existing provider accounts, use BYOK
   - No UniRoute charges, you pay providers directly

---

## 🆚 Pricing Comparison

### UniRoute vs Direct Provider Pricing

| Provider | Direct Price | UniRoute Price | Difference |
|----------|-------------|----------------|------------|
| GPT-4 (Input) | $30.00/1M | $36.23/1M | +20.8% |
| GPT-4 Turbo (Input) | $10.00/1M | $12.08/1M | +20.8% |
| Claude 3.5 Sonnet (Input) | $3.00/1M | $3.62/1M | +20.7% |

**Why the difference?**
- UniRoute pays providers on your behalf
- Includes operational overhead (infrastructure, support)
- Includes profit margin for sustainable business
- Provides unified billing and simplified management

**Value Proposition:**
- ✅ One API for all providers
- ✅ Intelligent routing and failover
- ✅ Built-in analytics and monitoring
- ✅ Unified billing (one invoice)
- ✅ No key management
- ✅ Free tunneling, webhook testing, and more

---

## 🎯 Future Enhancements

### Planned Pricing Features

1. **Usage-Based Discounts**
   - Automatic discounts at volume thresholds
   - Tiered pricing based on monthly usage

2. **Reserved Capacity**
   - Pre-purchase tokens at discounted rates
   - Guaranteed availability for high-volume users

3. **Custom Pricing**
   - Enterprise custom pricing
   - Volume commitments with better rates

4. **Billing Features**
   - Detailed cost breakdowns
   - Budget alerts and limits
   - Cost optimization recommendations

5. **Subscription Management**
   - Easy plan upgrades/downgrades
   - Prorated billing
   - Usage forecasting

---

## 📝 Implementation Notes

### Current Status

- ✅ Pay-as-you-go pricing implemented
- ✅ Cost calculation based on provider costs
- ✅ Analytics tracking for usage and costs
- ⏳ Subscription plans (to be implemented)
- ⏳ Volume discounts (to be implemented)
- ⏳ Billing system (to be implemented)

### Technical Implementation

**Cost Calculation:**
- Implemented in `internal/gateway/cost_calculator.go`
- Real-time cost tracking per request
- Historical cost data in analytics

**Billing System (Future):**
- Database schema for subscriptions
- Payment processing integration
- Invoice generation
- Usage metering and billing

**Subscription Management (Future):**
- Plan selection and upgrades
- Usage tracking per plan
- Billing cycle management
- Payment method management

---

## 🔐 Pricing Security

### Cost Protection

1. **Budget Limits**
   - Set monthly spending limits
   - Automatic alerts when approaching limits
   - Request blocking at limit (optional)

2. **Cost Transparency**
   - Real-time cost tracking
   - Detailed usage reports
   - Cost breakdown by provider/model

3. **Fraud Prevention**
   - Rate limiting per API key
   - Usage anomaly detection
   - Automatic fraud alerts

---

## 📞 Support & Questions

### Pricing Questions

- **General Pricing**: See pricing page on website
- **Custom Pricing**: Contact sales@uniroute.dev
- **Billing Issues**: Contact support@uniroute.dev
- **Cost Optimization**: Use analytics dashboard

### Resources

- [Pricing Calculator](https://uniroute.dev/pricing/calculator) (Future)
- [Cost Optimization Guide](https://uniroute.dev/docs/cost-optimization) (Future)
- [Billing FAQ](https://uniroute.dev/docs/billing-faq) (Future)

---

## 📋 Summary

**UniRoute Pricing Model:**
- ✅ **Pay-as-you-go**: Current model, no monthly minimums
- ✅ **Transparent**: Based on provider costs + margin
- ✅ **Flexible**: BYOK and self-hosted options are free
- ✅ **Value**: Unified billing, intelligent routing, free features
- ⏳ **Subscriptions**: Future plans with volume discounts
- ⏳ **Enterprise**: Custom pricing for high-volume users

**Key Points:**
- Only AI model usage is charged (when using managed service)
- All other features (tunneling, analytics, etc.) are free
- BYOK and self-hosted options are completely free
- Pricing is transparent and based on actual provider costs
- Future subscription plans will offer better value for high-volume users

---

**Last Updated**: 2024-12-26  
**Status**: Pay-as-you-go implemented, subscriptions planned

