# UniRoute: Hosted vs Self-Hosted

## 🎯 Two Ways to Use UniRoute

UniRoute can be used in **two ways**:

1. **Hosted Service** (UniRoute Cloud) - Use UniRoute's public service
2. **Self-Hosted** - Run UniRoute on your own infrastructure

---

## ☁️ Option 1: Hosted Service (UniRoute Cloud)

### What It Is
- UniRoute runs the gateway for you
- You access it at `https://api.uniroute.dev`
- No setup, no infrastructure management
- Just sign up and start using

### How It Works

#### Model A: UniRoute Managed Service ⭐ **PRIMARY MODEL**

**UniRoute provides and manages the provider API keys:**

1. **Sign up** for UniRoute account
2. **Get your UniRoute API key** from dashboard
3. **Start using** - UniRoute handles provider keys automatically
4. **Pay UniRoute** - UniRoute bills you (includes provider costs)

**Flow:**
```
Your App
  ↓ (uses UniRoute API key: ur_abc123...)
UniRoute Cloud (api.uniroute.dev)
  ↓ (uses UniRoute's provider keys)
AI Providers (OpenAI, Anthropic, Google)
  ↓ (bills UniRoute)
UniRoute bills you
Your App
```

**Benefits:**
- ✅ No need to get provider API keys
- ✅ No infrastructure to manage
- ✅ Simple billing (one invoice)
- ✅ Easier for teams (no key management)
- ✅ Unified cost tracking

**Cost:**
- UniRoute: **Pay-as-you-go** (includes provider costs)
- Providers: Billed through UniRoute
- **Note**: UniRoute charges you because UniRoute pays the providers on your behalf

---

#### Model B: Bring Your Own Keys (BYOK) - Optional

**You provide your own provider API keys:**

1. **Sign up** for UniRoute account
2. **Get your UniRoute API key** from dashboard
3. **Configure your provider keys** in UniRoute dashboard:
   - Add your OpenAI API key
   - Add your Anthropic API key
   - Add your Google API key
4. **Start using** UniRoute API

**Flow:**
```
Your App
  ↓ (uses UniRoute API key: ur_abc123...)
UniRoute Cloud (api.uniroute.dev)
  ↓ (uses YOUR provider keys stored securely)
AI Providers (OpenAI, Anthropic, Google)
  ↓ (bills YOU directly)
Your App
```

**Benefits:**
- ✅ You control your provider accounts
- ✅ You pay providers directly (no markup)
- ✅ Secure key storage (encrypted)
- ✅ Useful for existing provider accounts

**Cost:**
- UniRoute: **FREE** - No charges when you provide your own keys
- Providers: You pay them directly
- **Note**: UniRoute does NOT charge you when using BYOK - you pay providers directly

**Note:** BYOK is optional. Most users prefer the managed service for simplicity.

---

## 🏠 Option 2: Self-Hosted

### What It Is
- You run UniRoute on your own infrastructure
- Full control over everything
- Deploy anywhere (Docker, Coolify, Kubernetes, etc.)
- 100% free (no UniRoute charges)

### How It Works

1. **Deploy UniRoute** (Docker, Coolify, etc.)
2. **Get provider API keys** from OpenAI, Anthropic, Google
3. **Configure provider keys** in environment variables
4. **Create UniRoute API keys** in your instance
5. **Start using** your UniRoute instance

**Flow:**
```
Your App
  ↓ (uses UniRoute API key: ur_abc123...)
Your UniRoute Instance (your-server.com)
  ↓ (uses YOUR provider keys from .env)
AI Providers (OpenAI, Anthropic, Google)
  ↓ (bills YOU directly)
Your App
```

**Benefits:**
- ✅ 100% free (no UniRoute charges)
- ✅ Full control and privacy
- ✅ Your keys stay on your infrastructure
- ✅ Customizable and extensible
- ✅ No vendor lock-in

**Cost:**
- UniRoute: **FREE** - No charges for self-hosting (open source)
- Infrastructure: You pay for hosting (VPS, cloud, etc.)
- Providers: You pay them directly
- **Note**: UniRoute does NOT charge you when self-hosting - completely free

---

## 📊 Comparison Table

| Feature | Hosted (BYOK) | Hosted (Managed) | Self-Hosted |
|---------|---------------|------------------|-------------|
| **Setup Time** | Minutes | Minutes | 1-2 hours |
| **Infrastructure** | Managed by UniRoute | Managed by UniRoute | You manage |
| **Provider Keys** | You provide | UniRoute provides | You provide |
| **UniRoute Cost** | Pay-as-you-go | Pay-as-you-go | FREE |
| **Provider Billing** | You pay directly | Through UniRoute | You pay directly |
| **Privacy** | High (keys encrypted) | Medium | Highest (your infra) |
| **Control** | Medium | Low | Full |
| **Maintenance** | None | None | You maintain |
| **Scalability** | Auto-scales | Auto-scales | You scale |
| **Best For** | Most users | Teams, enterprises | Privacy-focused, cost-conscious |

---

## 🎯 Which Should You Choose?

### Choose **Hosted (Managed Service)** if:
- ✅ You want the simplest setup (no keys to manage)
- ✅ You want unified billing (one invoice)
- ✅ You're a team/enterprise
- ✅ You want UniRoute to handle everything
- ✅ You prefer pay-as-you-go pricing

### Choose **Hosted (BYOK)** if:
- ✅ You want to control your provider accounts
- ✅ You want to pay providers directly (no markup)
- ✅ You have existing provider accounts
- ✅ You want secure key storage

### Choose **Self-Hosted** if:
- ✅ You want 100% free (no UniRoute charges)
- ✅ You want maximum privacy and control
- ✅ You have infrastructure already
- ✅ You want to customize/extend UniRoute
- ✅ You're privacy-conscious or cost-conscious

---

## 🔄 Switching Between Options

### You Can Switch Anytime!

**From Hosted to Self-Hosted:**
- Export your UniRoute API keys
- Deploy UniRoute on your infrastructure
- Configure your provider keys
- Update your app to point to your instance

**From Self-Hosted to Hosted:**
- Sign up for UniRoute account
- Configure your provider keys in dashboard
- Get your UniRoute API key
- Update your app to use `api.uniroute.dev`

**Your provider keys stay the same** - just where they're stored changes!

---

## 🔐 Security Comparison

### Hosted Service
- ✅ Keys encrypted at rest
- ✅ Keys encrypted in transit
- ✅ Secure key storage (industry standards)
- ✅ Regular security audits
- ✅ DDoS protection
- ✅ Auto-scaling and redundancy

### Self-Hosted
- ✅ Keys on your infrastructure
- ✅ You control security
- ✅ No third-party key storage
- ✅ Full audit trail
- ✅ You manage security updates

---

## 💰 Cost Breakdown

### Hosted (Managed Service - Primary)
```
UniRoute:        Pay-as-you-go (includes provider costs)
Infrastructure:  Included
OpenAI:          Included in UniRoute bill
Anthropic:       Included in UniRoute bill
Google:          Included in UniRoute bill
─────────────────────────
Total:           Provider costs (UniRoute pays providers, bills you)
```

### Hosted (BYOK - Optional)
```
UniRoute:        FREE
Infrastructure:  FREE (UniRoute covers it)
OpenAI:          You pay directly
Anthropic:       You pay directly
Google:          You pay directly
─────────────────────────
Total:           Just provider costs
```

### Self-Hosted
```
UniRoute:        FREE
Infrastructure:  $5-50/month (VPS/cloud)
OpenAI:          You pay directly
Anthropic:       You pay directly
Google:          Included in UniRoute bill
─────────────────────────
Total:           Infrastructure + provider costs
```

---

## 🚀 Getting Started

### Hosted Service (Managed) - Recommended

1. **Sign up** at https://uniroute.dev (or api.uniroute.dev)
2. **Get your UniRoute API key** from dashboard
3. **Start using** the API (UniRoute handles provider keys):

```javascript
const response = await fetch('https://api.uniroute.dev/v1/chat', {
  headers: {
    'Authorization': 'Bearer ur_your-key'
  },
  body: JSON.stringify({
    model: 'gpt-4',
    messages: [...]
  })
});
```

### Self-Hosted

1. **Deploy UniRoute** (see [COOLIFY_DEPLOYMENT.md](../COOLIFY_DEPLOYMENT.md))
2. **Configure provider keys** in `.env`
3. **Create UniRoute API key** in dashboard
4. **Start using** your instance

---

## 📝 Summary

**Hosted Service:**
- ✅ Easy setup, no infrastructure
- ✅ Managed: UniRoute provides keys, you pay UniRoute (pay-as-you-go)
- ✅ BYOK: You provide keys, pay providers directly (FREE, optional)

**Self-Hosted:**
- ✅ 100% free, full control
- ✅ You provide keys, pay providers directly
- ✅ You manage infrastructure

**Both are great options!** Choose based on your needs:
- **Most users**: Hosted (Managed) - Simplest, unified billing
- **Existing provider accounts**: Hosted (BYOK) - Use your own keys
- **Privacy/Cost-conscious**: Self-Hosted - Maximum control, free

---

**Ready to get started?** Check out the [Quick Start Guide](../QUICKSTART.md)!

