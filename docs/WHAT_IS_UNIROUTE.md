# What is UniRoute? A Complete Guide

## 🎯 What is UniRoute?

**UniRoute is a unified gateway platform that routes, secures, and manages traffic to any LLM (Large Language Model) or AI model—cloud or local—with one unified API.**

Think of it as a **"smart router"** for AI models. Instead of connecting directly to OpenAI, Anthropic, Google, or your local Ollama instance, you connect to UniRoute, and it handles everything for you.

---

## 🤔 When Would You Use UniRoute?

### Scenario 1: You're Building an AI Application

**Problem**: Your app needs to use AI models, but:
- Different models have different APIs (OpenAI, Anthropic, Google all have different formats)
- You want to switch models easily (maybe OpenAI is down, or you want to use a cheaper model)
- You need to manage API keys for multiple providers
- You want to track usage, costs, and performance

**Solution**: Use UniRoute as a single API endpoint. Your app talks to UniRoute, and UniRoute talks to the AI models.

```javascript
// Instead of this (direct to OpenAI):
const response = await fetch('https://api.openai.com/v1/chat/completions', {
  headers: { 'Authorization': 'Bearer sk-...' },
  body: JSON.stringify({ model: 'gpt-4', messages: [...] })
});

// You do this (through UniRoute):
const response = await fetch('https://api.uniroute.dev/v1/chat', {
  headers: { 'Authorization': 'Bearer ur_your-key' },
  body: JSON.stringify({ model: 'gpt-4', messages: [...] })
});
```

**Benefits**:
- ✅ One API format for all models
- ✅ Automatic failover (if OpenAI is down, use Anthropic)
- ✅ One API key to manage
- ✅ Built-in usage tracking and analytics

---

### Scenario 2: You Want to Use Local AI Models (Privacy & Cost)

**Problem**: 
- You want to use local models (Ollama, vLLM) for privacy
- But you also want cloud models for better performance
- You need to switch between them easily

**Solution**: UniRoute supports both local and cloud models. You can:
- Use local models for sensitive data (privacy)
- Use cloud models for better performance
- Route automatically based on your needs

```bash
# Start your local Ollama
ollama serve

# UniRoute automatically detects and routes to it
# You can also use cloud models when needed
```

**Benefits**:
- ✅ Privacy: Keep sensitive data local
- ✅ Cost: Free local models (no API costs)
- ✅ Flexibility: Use cloud when needed
- ✅ Seamless switching between local and cloud

---

### Scenario 3: You Need Intelligent Routing

**Problem**: 
- You want to use the cheapest model for simple tasks
- You want the fastest model for real-time responses
- You want the best model for complex tasks
- You need automatic load balancing

**Solution**: UniRoute's intelligent routing:
- **Cost-based routing**: Automatically uses cheaper models when appropriate
- **Latency-based routing**: Uses the fastest available model
- **Model-based routing**: Routes to the right model for each request
- **Load balancing**: Distributes requests across multiple instances

**Example**:
```javascript
// UniRoute automatically chooses:
// - GPT-3.5 for simple questions (cheaper)
// - GPT-4 for complex reasoning (better quality)
// - Local Ollama for sensitive data (privacy)
// - Fastest available model for real-time needs
```

---

### Scenario 4: You Need Analytics & Monitoring

**Problem**: 
- You want to track how much you're spending on AI
- You need to monitor performance and latency
- You want to see usage patterns
- You need to optimize costs

**Solution**: UniRoute provides:
- **Real-time analytics**: See requests, costs, latency
- **Usage tracking**: Track per-user, per-model usage
- **Cost analysis**: See exactly what you're spending
- **Performance metrics**: Monitor response times

**Dashboard shows**:
- Total requests today: 10,000
- Total cost: $45.23
- Average latency: 234ms
- Most used model: GPT-4
- Cost per request: $0.0045

---

### Scenario 5: You Need Security & Rate Limiting

**Problem**:
- You need to secure your AI API access
- You want to prevent abuse
- You need rate limiting per user
- You want to track who's using what

**Solution**: UniRoute provides:
- **API key management**: Secure, scoped API keys
- **Rate limiting**: Per-key, per-IP limits
- **Authentication**: JWT-based auth
- **Access control**: Role-based permissions

**Example**:
```javascript
// Create API key with limits
POST /admin/api-keys
{
  "name": "Production API Key",
  "rate_limit_per_minute": 100,
  "rate_limit_per_day": 10000
}

// Use it in your app
const response = await fetch('https://api.uniroute.dev/v1/chat', {
  headers: { 'Authorization': 'Bearer ur_abc123...' }
});
```

---

## 🎯 Real-World Use Cases

### 1. **SaaS Application with AI Features**
- Your SaaS app needs AI for customer support, content generation, etc.
- Use UniRoute to manage all AI model access
- Track costs and usage per customer
- Switch models easily as needs change

### 2. **Development Team**
- Your team is building multiple AI-powered features
- Instead of each developer managing their own API keys
- Use UniRoute as a central gateway
- Track usage across the team
- Control costs and access

### 3. **Privacy-Conscious Application**
- You're building a healthcare or financial app
- Need to use local models for sensitive data
- But also want cloud models for better performance
- UniRoute lets you route based on data sensitivity

### 4. **Cost-Optimized Application**
- You want to minimize AI costs
- Use cheaper models when possible
- Use expensive models only when needed
- UniRoute's cost-based routing does this automatically

### 5. **High-Traffic Application**
- Your app gets thousands of AI requests per day
- Need load balancing across multiple model instances
- Need failover if one provider is down
- UniRoute handles this automatically

---

## 🔄 How It Works

### Simple Flow:

```
Your Application
      ↓
   UniRoute Gateway
      ↓
   [Intelligent Router]
      ↓
┌─────┴─────┬─────────┬──────────┐
│           │         │          │
OpenAI   Anthropic  Google   Local Ollama
```

### Detailed Flow:

1. **Your app sends a request** to UniRoute
   ```javascript
   POST /v1/chat
   {
     "model": "gpt-4",
     "messages": [...]
   }
   ```

2. **UniRoute authenticates** your request
   - Validates API key
   - Checks rate limits
   - Logs the request

3. **UniRoute routes** to the best provider
   - Checks model availability
   - Considers cost, latency, load
   - Selects optimal provider

4. **UniRoute forwards** to the AI provider
   - Formats request correctly
   - Handles provider-specific quirks
   - Manages timeouts and retries

5. **UniRoute processes** the response
   - Normalizes response format
   - Calculates cost and latency
   - Logs analytics

6. **UniRoute returns** unified response
   ```javascript
   {
     "id": "chat-123",
     "model": "gpt-4",
     "provider": "openai",  // Which provider was used
     "choices": [...],
     "usage": { "tokens": 150 },
     "cost": 0.002,         // Actual cost
     "latency_ms": 234      // Response time
   }
   ```

---

## 💡 Key Benefits

### 1. **Unified API**
- One API format for all models
- No need to learn different APIs
- Consistent response format

### 2. **Intelligent Routing**
- Automatic failover
- Cost optimization
- Latency optimization
- Load balancing

### 3. **Security & Control**
- API key management
- Rate limiting
- Access control
- Usage tracking

### 4. **Analytics & Monitoring**
- Real-time metrics
- Cost tracking
- Performance monitoring
- Usage analytics

### 5. **Flexibility**
- Support for cloud and local models
- Easy model switching
- Custom routing rules
- Extensible architecture

### 6. **Cost Optimization**
- Automatic cost-based routing
- Usage tracking
- Cost analysis
- Budget alerts

---

## 🔑 Who Provides the AI Model API Keys?

### Important: Two Types of API Keys

UniRoute uses **two different types of API keys**:

#### 1. **UniRoute API Keys** (Provided by UniRoute)
- **What**: Your authentication key to use UniRoute
- **Format**: `ur_abc123...`
- **Who provides**: UniRoute (you get this from UniRoute dashboard)
- **Purpose**: Authenticate your requests to UniRoute
- **Example**: `Authorization: Bearer ur_your-uniroute-key`

#### 2. **AI Provider API Keys** (Provided by YOU)
- **What**: API keys for OpenAI, Anthropic, Google, etc.
- **Format**: `sk-...` (OpenAI), `sk-ant-...` (Anthropic), `AIza...` (Google)
- **Who provides**: **YOU** (you get these from OpenAI, Anthropic, Google directly)
- **Purpose**: UniRoute uses these to make requests to AI providers on your behalf
- **Where**: Configured on your UniRoute server (self-hosted) or in your UniRoute account (hosted)

### Self-Hosted Setup (You Provide Provider Keys)

If you **self-host UniRoute**, you need to:

1. **Get API keys from AI providers**:
   - OpenAI: Get from https://platform.openai.com/api-keys
   - Anthropic: Get from https://console.anthropic.com/
   - Google: Get from https://makersuite.google.com/app/apikey

2. **Configure them in UniRoute**:
   ```bash
   # In your .env file or environment variables
   OPENAI_API_KEY=sk-your-openai-key
   ANTHROPIC_API_KEY=sk-ant-your-anthropic-key
   GOOGLE_API_KEY=AIza-your-google-key
   ```

3. **UniRoute uses your keys** to make requests to providers
4. **You pay the providers directly** (OpenAI, Anthropic, Google bill you)

### Hosted Setup (UniRoute Managed Service)

If you use **UniRoute's hosted service**:

- **Primary Model**: UniRoute provides provider keys (managed service)
  - UniRoute has pre-configured provider accounts
  - You pay UniRoute (who pays the providers)
  - Easier setup, unified billing
  - Pay-as-you-go pricing
  - **Note**: UniRoute charges you because UniRoute pays the providers

- **Optional**: You provide your own provider keys (bring your own keys)
  - You configure your OpenAI/Anthropic/Google keys in your UniRoute account
  - You pay providers directly
  - More control, you manage your own billing
  - **UniRoute does NOT charge you** - completely free when using BYOK

### Local Models (No API Keys Needed)

For **local models** (Ollama, vLLM):
- ✅ **No API keys needed** - They run on your machine
- ✅ **100% free** - No costs
- ✅ **Private** - Data never leaves your machine

---

## 🚀 Getting Started

### For Developers (Self-Hosted):

1. **Get AI provider API keys** from OpenAI, Anthropic, Google (if you want to use them)
2. **Deploy UniRoute** (Docker, Coolify, etc.)
3. **Configure provider keys** in environment variables
4. **Create a UniRoute API key** in the dashboard
5. **Start using** the unified API:

```javascript
// Install UniRoute SDK (or use REST API)
import { UniRoute } from '@uniroute/sdk';

const client = new UniRoute({
  apiKey: 'ur_your-key',
  baseURL: 'https://api.uniroute.dev'
});

// Use any model with one API
const response = await client.chat({
  model: 'gpt-4',  // or 'claude-3', 'gemini', 'llama2', etc.
  messages: [
    { role: 'user', content: 'Hello!' }
  ]
});
```

### For Self-Hosters:

1. **Get AI provider API keys**:
   - OpenAI: https://platform.openai.com/api-keys
   - Anthropic: https://console.anthropic.com/
   - Google: https://makersuite.google.com/app/apikey

2. **Deploy** UniRoute (Docker, Coolify, etc.)

3. **Configure provider keys** in environment variables:
   ```bash
   OPENAI_API_KEY=sk-your-key
   ANTHROPIC_API_KEY=sk-ant-your-key
   GOOGLE_API_KEY=AIza-your-key
   ```

4. **Start routing** requests - UniRoute will use your keys to access providers

5. **You pay providers directly** - OpenAI, Anthropic, Google bill you, not UniRoute

---

## 📊 Example: Before vs After

### Before UniRoute:

```javascript
// Multiple API clients
import OpenAI from 'openai';
import Anthropic from '@anthropic-ai/sdk';
import { GoogleGenerativeAI } from '@google/generative-ai';

// Different APIs for each
const openai = new OpenAI({ apiKey: 'sk-...' });
const anthropic = new Anthropic({ apiKey: 'sk-ant-...' });
const google = new GoogleGenerativeAI('AIza...');

// Different response formats
const openaiResponse = await openai.chat.completions.create({...});
const anthropicResponse = await anthropic.messages.create({...});
const googleResponse = await google.getGenerativeModel({...}).generateContent({...});

// Manual failover
try {
  return await openai.chat.completions.create({...});
} catch (error) {
  return await anthropic.messages.create({...});
}

// Manual cost tracking
// Manual rate limiting
// Manual analytics
```

### After UniRoute:

```javascript
// One API client
import { UniRoute } from '@uniroute/sdk';

const client = new UniRoute({ apiKey: 'ur_your-key' });

// One API for all models
const response = await client.chat({
  model: 'gpt-4',  // or any model
  messages: [...]
});

// Automatic failover
// Automatic cost tracking
// Built-in rate limiting
// Built-in analytics
```

---

## 🎯 Summary

**UniRoute is for anyone who:**
- ✅ Uses multiple AI models (OpenAI, Anthropic, Google, local)
- ✅ Wants a unified API for all models
- ✅ Needs intelligent routing and failover
- ✅ Wants to track costs and usage
- ✅ Needs security and rate limiting
- ✅ Wants to use local models for privacy
- ✅ Wants to optimize costs automatically

**Think of UniRoute as:**
- 🚪 **A gateway** - Single entry point for all AI models
- 🧠 **A router** - Intelligently routes to the best model
- 🔒 **A security layer** - Manages access and rate limits
- 📊 **An analytics platform** - Tracks usage and costs
- 🎛️ **A control panel** - Manage all your AI models in one place

---

**Ready to get started?** Check out the [Quick Start Guide](../QUICKSTART.md) or [Installation Guide](../CLI_INSTALLATION.md).

