<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-950 overflow-x-hidden">
    <!-- Navigation -->
    <nav class="fixed top-0 left-0 right-0 z-50 bg-slate-950/95 backdrop-blur-xl border-b border-slate-800/50">
      <div class="container mx-auto px-6 py-4">
        <div class="flex items-center justify-between">
          <router-link to="/" class="flex items-center space-x-3">
            <div class="w-10 h-10 bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-500 rounded-lg flex items-center justify-center shadow-lg shadow-blue-500/20">
              <span class="text-white font-bold text-lg">U</span>
            </div>
            <span class="text-xl font-bold text-white tracking-tight">UniRoute</span>
          </router-link>
          <!-- Desktop Navigation -->
          <div class="hidden md:flex items-center space-x-8">
            <a href="#features" class="text-sm font-medium text-slate-300 hover:text-white transition-colors">
              Features
            </a>
            <a href="#providers" class="text-sm font-medium text-slate-300 hover:text-white transition-colors">
              Providers
            </a>
            <a href="#pricing" class="text-sm font-medium text-slate-300 hover:text-white transition-colors">
              Pricing
            </a>
            <a href="#docs" class="text-sm font-medium text-slate-300 hover:text-white transition-colors">
              Docs
            </a>
            <router-link
              to="/download"
              class="text-sm font-medium text-slate-300 hover:text-white transition-colors"
            >
              Download
            </router-link>
            <router-link
              to="/login"
              class="text-sm font-medium text-slate-300 hover:text-white transition-colors"
            >
              Sign in
            </router-link>
            <router-link
              to="/register"
              class="px-4 py-2 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-sm font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20"
            >
              Get started
            </router-link>
          </div>
          <!-- Mobile Menu Button -->
          <button
            @click="mobileMenuOpen = !mobileMenuOpen"
            class="md:hidden p-2 text-slate-300 hover:text-white transition-colors"
            aria-label="Toggle menu"
          >
            <svg v-if="!mobileMenuOpen" class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
            <svg v-else class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <!-- Mobile Navigation Menu -->
        <div
          v-if="mobileMenuOpen"
          class="md:hidden mt-4 pb-4 border-t border-slate-800/50 pt-4 space-y-3"
        >
          <a
            href="#features"
            @click="mobileMenuOpen = false"
            class="block text-sm font-medium text-slate-300 hover:text-white transition-colors py-2"
          >
            Features
          </a>
          <a
            href="#providers"
            @click="mobileMenuOpen = false"
            class="block text-sm font-medium text-slate-300 hover:text-white transition-colors py-2"
          >
            Providers
          </a>
          <a
            href="#pricing"
            @click="mobileMenuOpen = false"
            class="block text-sm font-medium text-slate-300 hover:text-white transition-colors py-2"
          >
            Pricing
          </a>
          <a
            href="#docs"
            @click="mobileMenuOpen = false"
            class="block text-sm font-medium text-slate-300 hover:text-white transition-colors py-2"
          >
            Docs
          </a>
          <router-link
            to="/download"
            @click="mobileMenuOpen = false"
            class="block text-sm font-medium text-slate-300 hover:text-white transition-colors py-2"
          >
            Download
          </router-link>
          <router-link
            to="/login"
            @click="mobileMenuOpen = false"
            class="block text-sm font-medium text-slate-300 hover:text-white transition-colors py-2"
          >
            Sign in
          </router-link>
          <router-link
            to="/register"
            @click="mobileMenuOpen = false"
            class="block px-4 py-2 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-sm font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all text-center"
          >
            Get started
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <section class="relative overflow-hidden pt-20 pb-16 md:pb-20">
      <!-- Animated background elements -->
      <div class="absolute inset-0 overflow-hidden">
        <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl animate-pulse"></div>
        <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-pulse" style="animation-delay: 1s"></div>
      </div>
      
      <div class="container mx-auto px-6 py-4 md:py-4 relative z-10">
        <div class="max-w-5xl mx-auto text-center">
          <div class="inline-flex items-center px-4 py-2 mb-8 bg-blue-500/10 backdrop-blur-sm border border-blue-500/20 text-blue-300 rounded-full text-sm font-medium">
            <span class="w-2 h-2 bg-blue-400 rounded-full mr-2 animate-pulse"></span>
            AI Gateway + Secure Tunneling in one platform
          </div>
          <h1 class="text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-6 leading-tight tracking-tight">
            The AI Gateway
            <span class="block mt-2 bg-gradient-to-r from-blue-400 via-indigo-400 to-purple-400 bg-clip-text text-transparent">
              & Secure Tunneling Platform
            </span>
          </h1>
          <p class="text-xl md:text-2xl text-slate-300 mb-12 max-w-3xl mx-auto leading-relaxed">
            Route, secure, and manage traffic to any LLM—cloud or local. 
            <span class="text-white font-semibold">One unified API for OpenAI, Anthropic, Google, and local models.</span>
            <span class="block mt-3 text-lg">Plus, expose local services instantly with built-in HTTP, TCP, and TLS tunneling.</span>
          </p>
          <div class="flex flex-col sm:flex-row gap-4 justify-center items-center mb-8">
            <router-link
              to="/register"
              class="px-8 py-4 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-base font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20 hover:shadow-xl hover:shadow-blue-500/30 w-full sm:w-auto"
            >
              Start free →
            </router-link>
            <router-link
              to="/download"
              class="px-8 py-4 bg-slate-800/60 text-white rounded-lg text-base font-semibold border border-slate-700 hover:border-slate-600 transition-all w-full sm:w-auto"
            >
              Download CLI
            </router-link>
          </div>
          <p class="text-sm text-slate-400">
            No credit card required • Free with BYOK or self-hosted • Open Source
          </p>
        </div>

        <!-- Code Examples -->
        <div class="mt-6 md:mt-8 lg:mt-12 max-w-6xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6 px-4">
          <!-- API Code Example -->
          <div class="bg-slate-900/90 rounded-xl md:rounded-2xl border border-slate-800 p-3 md:p-6 lg:p-8 shadow-2xl">
            <div class="flex items-center space-x-2 mb-3 md:mb-4">
              <div class="w-2 h-2 md:w-3 md:h-3 bg-red-500 rounded-full flex-shrink-0"></div>
              <div class="w-2 h-2 md:w-3 md:h-3 bg-yellow-500 rounded-full flex-shrink-0"></div>
              <div class="w-2 h-2 md:w-3 md:h-3 bg-green-500 rounded-full flex-shrink-0"></div>
              <span class="ml-2 md:ml-4 text-[10px] md:text-xs lg:text-sm text-slate-400 whitespace-nowrap">One API for all AI models</span>
            </div>
            <div class="overflow-x-auto -mx-3 md:mx-0 px-3 md:px-0">
              <pre class="text-[10px] sm:text-xs md:text-sm text-slate-300 whitespace-pre overflow-x-auto scrollbar-thin"><code class="block min-w-max"><span class="text-purple-400">const</span> <span class="text-blue-400">response</span> = <span class="text-purple-400">await</span> <span class="text-yellow-400">fetch</span>(<span class="text-green-400">'https://api.uniroute.co/v1/chat'</span>, {
  <span class="text-blue-400">method</span>: <span class="text-green-400">'POST'</span>,
  <span class="text-blue-400">headers</span>: {
    <span class="text-green-400">'Authorization'</span>: <span class="text-green-400">'Bearer ur_your-key'</span>,
    <span class="text-green-400">'Content-Type'</span>: <span class="text-green-400">'application/json'</span>
  },
  <span class="text-blue-400">body</span>: <span class="text-yellow-400">JSON</span>.<span class="text-yellow-400">stringify</span>({
    <span class="text-blue-400">model</span>: <span class="text-green-400">'gpt-4o'</span>, <span class="text-slate-500">// or 'claude-3', 'gemini', 'llama2'</span>
    <span class="text-blue-400">messages</span>: [{
      <span class="text-blue-400">role</span>: <span class="text-green-400">'user'</span>,
      <span class="text-blue-400">content</span>: [
        { <span class="text-blue-400">type</span>: <span class="text-green-400">'text'</span>, <span class="text-blue-400">text</span>: <span class="text-green-400">'Analyze these images and audio'</span> },
        { <span class="text-blue-400">type</span>: <span class="text-green-400">'image_url'</span>, <span class="text-blue-400">image_url</span>: { <span class="text-blue-400">url</span>: <span class="text-green-400">'data:image/...'</span> } },
        { <span class="text-blue-400">type</span>: <span class="text-green-400">'image_url'</span>, <span class="text-blue-400">image_url</span>: { <span class="text-blue-400">url</span>: <span class="text-green-400">'data:image/...'</span> } },
        { <span class="text-blue-400">type</span>: <span class="text-green-400">'audio_url'</span>, <span class="text-blue-400">audio_url</span>: { <span class="text-blue-400">url</span>: <span class="text-green-400">'data:audio/...'</span> } }
      ]
    }]
  })
});</code></pre>
            </div>
          </div>

          <!-- Tunneling Code Example -->
          <div class="bg-slate-900/90 rounded-xl md:rounded-2xl border border-slate-800 p-3 md:p-6 lg:p-8 shadow-2xl">
            <div class="flex items-center space-x-2 mb-3 md:mb-4">
              <div class="w-2 h-2 md:w-3 md:h-3 bg-red-500 rounded-full flex-shrink-0"></div>
              <div class="w-2 h-2 md:w-3 md:h-3 bg-yellow-500 rounded-full flex-shrink-0"></div>
              <div class="w-2 h-2 md:w-3 md:h-3 bg-green-500 rounded-full flex-shrink-0"></div>
              <span class="ml-2 md:ml-4 text-[10px] md:text-xs lg:text-sm text-slate-400 whitespace-nowrap">Secure tunneling for any service</span>
            </div>
            <div class="overflow-x-auto -mx-3 md:mx-0 px-3 md:px-0">
              <pre class="text-[10px] sm:text-xs md:text-sm text-slate-300 whitespace-pre overflow-x-auto scrollbar-thin"><code class="block min-w-max"><span class="text-slate-500"># Create tunnels (auto-resumes if available)</span>
<span class="text-purple-400">uniroute</span> http 8080
<span class="text-purple-400">uniroute</span> tcp 3306
<span class="text-purple-400">uniroute</span> tls 5432
<span class="text-purple-400">uniroute</span> --all

<span class="text-slate-500"># Force new tunnel (don't resume)</span>
<span class="text-purple-400">uniroute</span> http 8080 --new

<span class="text-slate-500"># Manage tunnels</span>
<span class="text-purple-400">uniroute</span> list
<span class="text-purple-400">uniroute</span> resume abc123

<span class="text-slate-500"># Public URLs automatically assigned</span>
<span class="text-green-400">✓ https://abc123.uniroute.co</span>
<span class="text-green-400">✓ mysql-tunnel.uniroute.co:20001</span></code></pre>
            </div>
          </div>
        </div>

        <!-- Tunnel Commands Reference -->
        <div class="mt-8 md:mt-12 max-w-4xl mx-auto">
          <div class="bg-slate-800/60 rounded-xl border border-slate-700/50 p-6 md:p-8">
            <h3 class="text-2xl font-bold text-white mb-6 text-center">All Tunnel Commands</h3>
            <div class="grid md:grid-cols-2 gap-6">
              <div>
                <h4 class="text-lg font-semibold text-white mb-4 flex items-center">
                  <svg class="w-5 h-5 text-indigo-400 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  Create Tunnels
                </h4>
                <div class="space-y-2 text-sm">
                  <div class="flex items-start">
                    <code class="text-indigo-300 bg-slate-900/50 px-2 py-1 rounded mr-2">uniroute http 8080</code>
                    <span class="text-slate-400 text-xs mt-1">HTTP tunnel</span>
                  </div>
                  <div class="flex items-start">
                    <code class="text-indigo-300 bg-slate-900/50 px-2 py-1 rounded mr-2">uniroute tcp 3306</code>
                    <span class="text-slate-400 text-xs mt-1">TCP tunnel</span>
                  </div>
                  <div class="flex items-start">
                    <code class="text-indigo-300 bg-slate-900/50 px-2 py-1 rounded mr-2">uniroute tls 5432</code>
                    <span class="text-slate-400 text-xs mt-1">TLS tunnel</span>
                  </div>
                  <div class="flex items-start">
                    <code class="text-indigo-300 bg-slate-900/50 px-2 py-1 rounded mr-2">uniroute --all</code>
                    <span class="text-slate-400 text-xs mt-1">Start all tunnels</span>
                  </div>
                  <div class="flex items-start">
                    <code class="text-indigo-300 bg-slate-900/50 px-2 py-1 rounded mr-2">uniroute http 8080 --new</code>
                    <span class="text-slate-400 text-xs mt-1">Force new tunnel</span>
                  </div>
                </div>
              </div>
              <div>
                <h4 class="text-lg font-semibold text-white mb-4 flex items-center">
                  <svg class="w-5 h-5 text-green-400 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Manage Tunnels
                </h4>
                <div class="space-y-2 text-sm">
                  <div class="flex items-start">
                    <code class="text-green-300 bg-slate-900/50 px-2 py-1 rounded mr-2">uniroute list</code>
                    <span class="text-slate-400 text-xs mt-1">List all tunnels</span>
                  </div>
                  <div class="flex items-start">
                    <code class="text-green-300 bg-slate-900/50 px-2 py-1 rounded mr-2">uniroute resume abc123</code>
                    <span class="text-slate-400 text-xs mt-1">Resume tunnel</span>
                  </div>
                  <div class="flex items-start">
                    <code class="text-green-300 bg-slate-900/50 px-2 py-1 rounded mr-2">uniroute tunnel --init</code>
                    <span class="text-slate-400 text-xs mt-1">Create config</span>
                  </div>
                  <div class="flex items-start">
                    <code class="text-green-300 bg-slate-900/50 px-2 py-1 rounded mr-2">uniroute tunnel --clear</code>
                    <span class="text-slate-400 text-xs mt-1">Clear saved state</span>
                  </div>
                </div>
              </div>
            </div>
            <div class="mt-6 pt-6 border-t border-slate-700/50">
              <p class="text-sm text-slate-400 text-center">
                💡 <strong class="text-white">Tip:</strong> Tunnels automatically resume their previous subdomain when you run the same command again
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section id="features" class="py-4 md:py-4 lg:py-4 relative">
      <div class="container mx-auto px-6">
        <div class="text-center mb-12 md:mb-16">
          <h2 class="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            Everything you need to build AI applications
          </h2>
          <p class="text-xl text-slate-400 max-w-2xl mx-auto">
            Enterprise-grade features, zero complexity
          </p>
        </div>
        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8 max-w-7xl mx-auto">
          <!-- Unified API -->
          <div class="bg-slate-800/60 rounded-xl p-8 border border-slate-700/50 hover:border-blue-500/50 transition-all hover:shadow-xl hover:shadow-blue-500/10">
            <div class="w-14 h-14 bg-gradient-to-br from-blue-500 to-indigo-500 rounded-lg flex items-center justify-center mb-6 shadow-lg shadow-blue-500/20">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-3">Unified API</h3>
            <p class="text-slate-400 leading-relaxed">
              One API format for all AI models. No need to learn different APIs for OpenAI, Anthropic, Google, or local models. Consistent request/response format across all providers. Supports multimodal inputs including multiple images and audio files.
            </p>
          </div>

          <!-- Intelligent Routing -->
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50 hover:border-purple-500/50 transition-all hover:shadow-xl hover:shadow-purple-500/10">
            <div class="w-14 h-14 bg-gradient-to-br from-purple-500 to-pink-500 rounded-lg flex items-center justify-center mb-6 shadow-lg shadow-purple-500/20">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-3">Intelligent Routing</h3>
            <p class="text-slate-400 leading-relaxed">
              Automatic model selection based on cost, latency, and availability. Load balancing, failover, and cost-based routing strategies. Switch providers seamlessly without code changes.
            </p>
          </div>

          <!-- Multi-Provider Support -->
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50 hover:border-green-500/50 transition-all hover:shadow-xl hover:shadow-green-500/10">
            <div class="w-14 h-14 bg-gradient-to-br from-green-500 to-emerald-500 rounded-lg flex items-center justify-center mb-6 shadow-lg shadow-green-500/20">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-3">Multi-Provider Support</h3>
            <p class="text-slate-400 leading-relaxed">
              Support for OpenAI (GPT-4, GPT-3.5), Anthropic (Claude 3, Claude 2), Google (Gemini), and local models (Ollama, vLLM). Add new providers easily with our extensible architecture.
            </p>
          </div>

          <!-- Multimodal Support -->
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50 hover:border-cyan-500/50 transition-all hover:shadow-xl hover:shadow-cyan-500/10">
            <div class="w-14 h-14 bg-gradient-to-br from-cyan-500 to-blue-500 rounded-lg flex items-center justify-center mb-6 shadow-lg shadow-cyan-500/20">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-3">Multimodal Support</h3>
            <p class="text-slate-400 leading-relaxed">
              Send multiple images and audio files in a single request. Support for vision models, voice transcription, and multimodal AI interactions. Perfect for image analysis, document processing, and voice-enabled applications.
            </p>
          </div>

          <!-- Enterprise Security -->
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50 hover:border-yellow-500/50 transition-all hover:shadow-xl hover:shadow-yellow-500/10">
            <div class="w-14 h-14 bg-gradient-to-br from-yellow-500 to-orange-500 rounded-lg flex items-center justify-center mb-6 shadow-lg shadow-yellow-500/20">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-3">Enterprise Security</h3>
            <p class="text-slate-400 leading-relaxed">
              AES-256 encryption for API keys, rate limiting (per-key and per-IP), JWT authentication, API key management, and comprehensive access control. Your data is protected at every layer.
            </p>
          </div>

          <!-- Analytics & Monitoring -->
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50 hover:border-pink-500/50 transition-all hover:shadow-xl hover:shadow-pink-500/10">
            <div class="w-14 h-14 bg-gradient-to-br from-pink-500 to-rose-500 rounded-lg flex items-center justify-center mb-6 shadow-lg shadow-pink-500/20">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-3">Analytics & Monitoring</h3>
            <p class="text-slate-400 leading-relaxed">
              Track usage, costs, latency, and performance metrics. Prometheus integration, real-time dashboards, cost calculation per request, and comprehensive usage analytics.
            </p>
          </div>

          <!-- Built-in Tunneling -->
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50 hover:border-indigo-500/50 transition-all hover:shadow-xl hover:shadow-indigo-500/10">
            <div class="w-14 h-14 bg-gradient-to-br from-indigo-500 to-blue-500 rounded-lg flex items-center justify-center mb-6 shadow-lg shadow-indigo-500/20">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-3">Built-in Tunneling</h3>
            <p class="text-slate-400 leading-relaxed mb-3">
              Expose local services instantly with secure tunneling. Support for HTTP, TCP, and TLS protocols. Multiple tunnels, port mapping, and automatic resumption—all built-in and free.
            </p>
            <ul class="space-y-1.5 text-sm text-slate-300">
              <li class="flex items-start">
                <svg class="w-4 h-4 text-indigo-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>HTTP, TCP, and TLS support</span>
              </li>
              <li class="flex items-start">
                <svg class="w-4 h-4 text-indigo-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Multiple tunnels with config management</span>
              </li>
              <li class="flex items-start">
                <svg class="w-4 h-4 text-indigo-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Automatic tunnel resumption</span>
              </li>
            </ul>
          </div>

          <!-- Webhook Testing -->
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50 hover:border-cyan-500/50 transition-all hover:shadow-xl hover:shadow-cyan-500/10">
            <div class="w-14 h-14 bg-gradient-to-br from-cyan-500 to-blue-500 rounded-lg flex items-center justify-center mb-6 shadow-lg shadow-cyan-500/20">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-3">Webhook Testing</h3>
            <p class="text-slate-400 leading-relaxed">
              Get a secure public URL for your local web server. Inspect traffic, replay webhook requests, and iterate quickly while staying in context. Full request/response inspection.
            </p>
          </div>

          <!-- BYOK Model -->
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50 hover:border-emerald-500/50 transition-all hover:shadow-xl hover:shadow-emerald-500/10">
            <div class="w-14 h-14 bg-gradient-to-br from-emerald-500 to-teal-500 rounded-lg flex items-center justify-center mb-6 shadow-lg shadow-emerald-500/20">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-white mb-3">Bring Your Own Keys</h3>
            <p class="text-slate-400 leading-relaxed">
              Securely store your provider API keys (OpenAI, Anthropic, Google) with AES-256 encryption. You pay providers directly, we handle routing and management. No markup, no hidden fees.
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Providers Section -->
    <section id="providers" class="py-6 md:py-6 lg:py-6 relative">
      <div class="container mx-auto px-6">
        <div class="text-center mb-12 md:mb-16">
          <h2 class="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            All AI models, one API
          </h2>
          <p class="text-xl text-slate-400 max-w-2xl mx-auto">
            Support for all major AI providers and local models
          </p>
        </div>
        <div class="grid md:grid-cols-2 lg:grid-cols-4 gap-6 max-w-6xl mx-auto">
          <div class="bg-slate-800/60 rounded-xl p-6 border border-slate-700/50 text-center hover:border-blue-500/50 transition-all">
            <div class="w-16 h-16 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center mb-4 mx-auto shadow-lg shadow-blue-500/20">
              <span class="text-2xl font-bold text-white">O</span>
            </div>
            <h3 class="text-lg font-semibold text-white mb-2">OpenAI</h3>
            <p class="text-sm text-slate-400">GPT-4, GPT-3.5, GPT-4 Turbo</p>
          </div>
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-6 border border-slate-700/50 text-center hover:border-purple-500/50 transition-all">
            <div class="w-16 h-16 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg flex items-center justify-center mb-4 mx-auto shadow-lg shadow-purple-500/20">
              <span class="text-2xl font-bold text-white">A</span>
            </div>
            <h3 class="text-lg font-semibold text-white mb-2">Anthropic</h3>
            <p class="text-sm text-slate-400">Claude 3, Claude 2, Claude Instant</p>
          </div>
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-6 border border-slate-700/50 text-center hover:border-green-500/50 transition-all">
            <div class="w-16 h-16 bg-gradient-to-br from-green-500 to-green-600 rounded-lg flex items-center justify-center mb-4 mx-auto shadow-lg shadow-green-500/20">
              <span class="text-2xl font-bold text-white">G</span>
            </div>
            <h3 class="text-lg font-semibold text-white mb-2">Google</h3>
            <p class="text-sm text-slate-400">Gemini Pro, Gemini Ultra</p>
          </div>
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-6 border border-slate-700/50 text-center hover:border-cyan-500/50 transition-all">
            <div class="w-16 h-16 bg-gradient-to-br from-cyan-500 to-cyan-600 rounded-lg flex items-center justify-center mb-4 mx-auto shadow-lg shadow-cyan-500/20">
              <span class="text-2xl font-bold text-white">L</span>
            </div>
            <h3 class="text-lg font-semibold text-white mb-2">Local LLMs</h3>
            <p class="text-sm text-slate-400">Ollama, vLLM, Custom models</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Use Cases Section -->
    <section class="py-5 md:py-5 lg:py-5 relative">
      <div class="container mx-auto px-6">
        <div class="text-center mb-12 md:mb-16">
          <h2 class="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            Built for developers
          </h2>
          <p class="text-xl text-slate-400 max-w-2xl mx-auto">
            Use cases that make your AI development easier
          </p>
        </div>
        <div class="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50">
            <h3 class="text-xl font-semibold text-white mb-4">Multi-Provider Applications</h3>
            <p class="text-slate-400 leading-relaxed mb-4">
              Build apps that work with multiple AI providers. Switch between OpenAI, Anthropic, and Google without changing your code. Automatic failover ensures your app always works.
            </p>
            <ul class="space-y-2 text-sm text-slate-300">
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>One API for all providers</span>
              </li>
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Automatic failover</span>
              </li>
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Cost-based routing</span>
              </li>
            </ul>
          </div>
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50">
            <h3 class="text-xl font-semibold text-white mb-4">Local AI Development</h3>
            <p class="text-slate-400 leading-relaxed mb-4">
              Develop and test with local models (Ollama) for free, then deploy to cloud models in production. Same API, seamless transition. Perfect for privacy-sensitive applications.
            </p>
            <ul class="space-y-2 text-sm text-slate-300">
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Free local model support</span>
              </li>
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Privacy-first development</span>
              </li>
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Offline capabilities</span>
              </li>
            </ul>
          </div>
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50">
            <h3 class="text-xl font-semibold text-white mb-4">Local Development & Testing</h3>
            <p class="text-slate-400 leading-relaxed mb-4">
              Expose local services instantly for webhook testing, API development, and database access. Secure tunneling with HTTP, TCP, and TLS support. Perfect for development, testing, and demos.
            </p>
            <ul class="space-y-2 text-sm text-slate-300">
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Instant public URLs for local services</span>
              </li>
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>TCP/TLS for databases and services</span>
              </li>
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Multiple tunnels with one command</span>
              </li>
            </ul>
          </div>
          <div class="bg-slate-800/50 backdrop-blur-sm rounded-xl p-8 border border-slate-700/50">
            <h3 class="text-xl font-semibold text-white mb-4">Enterprise AI Infrastructure</h3>
            <p class="text-slate-400 leading-relaxed mb-4">
              Centralized AI gateway for your organization. Manage API keys, track usage, control costs, and ensure security across all AI applications. Self-hostable for complete control.
            </p>
            <ul class="space-y-2 text-sm text-slate-300">
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Centralized management</span>
              </li>
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Usage analytics</span>
              </li>
              <li class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span>Self-hostable</span>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <!-- How It Works -->
    <section class="py-6 md:py-6 lg:py-6 relative">
      <div class="container mx-auto px-6">
        <div class="text-center mb-12 md:mb-16">
          <h2 class="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            How it works
          </h2>
          <p class="text-xl text-slate-400 max-w-2xl mx-auto">
            Get started in minutes, not hours
          </p>
        </div>
        <div class="grid md:grid-cols-4 gap-8 max-w-6xl mx-auto">
          <div class="text-center">
            <div class="w-20 h-20 bg-gradient-to-br from-blue-500 to-indigo-500 rounded-full flex items-center justify-center mb-6 mx-auto text-3xl font-bold text-white shadow-lg shadow-blue-500/20">
              1
            </div>
            <h3 class="text-xl font-semibold text-white mb-4">Sign up free</h3>
            <p class="text-slate-400 leading-relaxed">
              Create your account in seconds. No credit card required. Get your UniRoute API key and start using immediately.
            </p>
          </div>
          <div class="text-center">
            <div class="w-20 h-20 bg-gradient-to-br from-indigo-500 to-purple-500 rounded-full flex items-center justify-center mb-6 mx-auto text-3xl font-bold text-white shadow-lg shadow-indigo-500/20">
              2
            </div>
            <h3 class="text-xl font-semibold text-white mb-4">Configure providers</h3>
            <p class="text-slate-400 leading-relaxed">
              Add your OpenAI, Anthropic, or Google API keys (BYOK). Or use local models with Ollama. All encrypted and secure.
            </p>
          </div>
          <div class="text-center">
            <div class="w-20 h-20 bg-gradient-to-br from-purple-500 to-pink-500 rounded-full flex items-center justify-center mb-6 mx-auto text-3xl font-bold text-white shadow-lg shadow-purple-500/20">
              3
            </div>
            <h3 class="text-xl font-semibold text-white mb-4">Start routing</h3>
            <p class="text-slate-400 leading-relaxed">
              Use our unified API to route requests to any provider. We handle security, failover, analytics, and cost tracking automatically.
            </p>
          </div>
          <div class="text-center">
            <div class="w-20 h-20 bg-gradient-to-br from-pink-500 to-rose-500 rounded-full flex items-center justify-center mb-6 mx-auto text-3xl font-bold text-white shadow-lg shadow-pink-500/20">
              4
            </div>
            <h3 class="text-xl font-semibold text-white mb-4">Expose local services</h3>
            <p class="text-slate-400 leading-relaxed">
              Use built-in tunneling to expose local servers instantly. HTTP, TCP, and TLS support. Perfect for webhooks, databases, and development.
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Pricing Section -->
    <section id="pricing" class="py-6 md:py-6 lg:py-6 relative">
      <div class="container mx-auto px-6">
        <div class="text-center mb-12 md:mb-16">
          <h2 class="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            Simple, transparent pricing
          </h2>
          <p class="text-xl text-slate-400">
            Flexible pricing options. Free when you bring your own keys or self-host.
          </p>
        </div>
        <div class="max-w-6xl mx-auto grid md:grid-cols-3 gap-6">
          <!-- Managed Service Plan -->
          <div class="bg-gradient-to-br from-slate-800/90 to-slate-900/90 rounded-xl md:rounded-2xl border-2 border-blue-500/30 shadow-2xl p-4 md:p-8">
            <div class="text-center mb-8">
              <div class="inline-flex items-center px-4 py-2 mb-4 bg-blue-500/20 backdrop-blur-sm text-blue-300 rounded-full text-sm font-semibold border border-blue-500/30">
                Managed Service
              </div>
              <h3 class="text-2xl font-bold text-white mb-4">Pay-as-you-go</h3>
              <div class="mb-4">
                <span class="text-4xl font-bold text-white">Pay</span>
                <span class="text-slate-400 text-lg ml-2">per usage</span>
              </div>
              <p class="text-sm text-slate-300 mb-4">
                UniRoute provides and manages provider keys. You pay UniRoute for <strong>AI model usage only</strong> (UniRoute pays providers on your behalf). Tunneling, webhook testing, analytics, routing, and all other features are completely free.
              </p>
              <div class="bg-slate-900/60 rounded-lg p-3 md:p-4 mb-4 md:mb-6">
                <p class="text-[10px] md:text-xs text-slate-400 mb-2 font-semibold">Example Pricing (per 1M tokens):</p>
                <div class="space-y-1 text-[10px] md:text-xs text-slate-300">
                  <div class="flex justify-between items-center gap-2">
                    <span class="truncate">GPT-4 Input:</span>
                    <span class="font-semibold text-white whitespace-nowrap">$36.23</span>
                  </div>
                  <div class="flex justify-between items-center gap-2">
                    <span class="truncate">GPT-3.5 Turbo Input:</span>
                    <span class="font-semibold text-white whitespace-nowrap">$0.60</span>
                  </div>
                  <div class="flex justify-between items-center gap-2">
                    <span class="truncate">Claude 3.5 Sonnet Input:</span>
                    <span class="font-semibold text-white whitespace-nowrap">$3.62</span>
                  </div>
                  <div class="flex justify-between items-center gap-2 border-t border-slate-700 pt-1 mt-1">
                    <span class="text-slate-400 truncate">Local Models:</span>
                    <span class="font-semibold text-green-400 whitespace-nowrap">FREE</span>
                  </div>
                </div>
              </div>
            </div>
            <div class="space-y-4 mb-8">
              <div class="flex items-start">
                <svg class="w-5 h-5 text-blue-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">Unified billing</span>
              </div>
              <div class="flex items-start">
                <svg class="w-5 h-5 text-blue-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">No key management</span>
              </div>
              <div class="flex items-start">
                <svg class="w-5 h-5 text-blue-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">Simplest setup</span>
              </div>
              <div class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">Free: Tunneling, webhook testing, analytics</span>
              </div>
            </div>
          </div>

          <!-- BYOK Plan -->
          <div class="bg-gradient-to-br from-slate-800/90 to-slate-900/90 rounded-xl md:rounded-2xl border-2 border-green-500/30 shadow-2xl p-4 md:p-8">
            <div class="text-center mb-8">
              <div class="inline-flex items-center px-4 py-2 mb-4 bg-green-500/20 backdrop-blur-sm text-green-300 rounded-full text-sm font-semibold border border-green-500/30">
                BYOK
              </div>
              <h3 class="text-2xl font-bold text-white mb-4">Free Forever</h3>
              <div class="mb-4">
                <span class="text-4xl font-bold text-white">$0</span>
                <span class="text-slate-400 text-lg ml-2">/month</span>
              </div>
              <p class="text-sm text-slate-300 mb-6">
                Bring your own provider keys. You pay providers directly. UniRoute service, tunneling, webhook testing, analytics, routing, and all features are completely free.
              </p>
            </div>
            <div class="space-y-4 mb-8">
              <div class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">No UniRoute charges</span>
              </div>
              <div class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">Control your keys</span>
              </div>
              <div class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">Pay providers directly</span>
              </div>
              <div class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">Free: Tunneling, webhook testing, analytics</span>
              </div>
            </div>
          </div>

          <!-- Self-Hosted Plan -->
          <div class="bg-gradient-to-br from-slate-800/90 to-slate-900/90 rounded-xl md:rounded-2xl border-2 border-purple-500/30 shadow-2xl p-4 md:p-8">
            <div class="text-center mb-8">
              <div class="inline-flex items-center px-4 py-2 mb-4 bg-purple-500/20 backdrop-blur-sm text-purple-300 rounded-full text-sm font-semibold border border-purple-500/30">
                Self-Hosted
              </div>
              <h3 class="text-2xl font-bold text-white mb-4">Free Forever</h3>
              <div class="mb-4">
                <span class="text-4xl font-bold text-white">$0</span>
                <span class="text-slate-400 text-lg ml-2">/month</span>
              </div>
              <p class="text-sm text-slate-300 mb-6">
                Deploy UniRoute on your infrastructure. Open source, completely free. All features including tunneling, webhook testing, analytics, routing, and more are free.
              </p>
            </div>
            <div class="space-y-4 mb-8">
              <div class="flex items-start">
                <svg class="w-5 h-5 text-purple-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">100% free & open source</span>
              </div>
              <div class="flex items-start">
                <svg class="w-5 h-5 text-purple-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">Full control & privacy</span>
              </div>
              <div class="flex items-start">
                <svg class="w-5 h-5 text-purple-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">Deploy anywhere</span>
              </div>
              <div class="flex items-start">
                <svg class="w-5 h-5 text-green-400 mr-3 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-sm text-slate-300">Free: Tunneling, webhook testing, analytics</span>
              </div>
            </div>
          </div>
        </div>
        <div class="max-w-4xl mx-auto mt-12">
          <div class="bg-slate-800/60 rounded-xl border border-slate-700/50 p-6">
            <div class="text-center">
              <p class="text-slate-300 mb-4">
                <strong class="text-white">Note:</strong> UniRoute charges for <strong>AI model usage only</strong> when using the managed service (because UniRoute pays providers on your behalf). 
                <strong>Tunneling, webhook testing, analytics, routing, and all other features are completely free</strong> in all plans. BYOK and self-hosted options are completely free (you pay providers directly).
              </p>
              <router-link
                to="/register"
                class="inline-block px-8 py-4 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-base font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20 hover:shadow-xl"
              >
                Get started free
              </router-link>
              <p class="mt-4 text-sm text-slate-400">
                No credit card required • Setup in 2 minutes
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Documentation Section -->
    <section id="docs" class="py-8 md:py-8 lg:py-8 relative">
      <div class="container mx-auto px-6">
        <div class="text-center mb-12 md:mb-16">
          <h2 class="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            API Documentation
          </h2>
          <p class="text-xl text-slate-400 max-w-2xl mx-auto">
            Interactive API documentation with Swagger UI. Explore all endpoints, test requests, and integrate with ease.
          </p>
        </div>
        <div class="max-w-4xl mx-auto">
          <div class="bg-slate-800/60 rounded-2xl border border-slate-700/50 p-8 md:p-12 shadow-2xl">
            <div class="text-center mb-8">
              <div class="w-20 h-20 bg-gradient-to-br from-blue-500 to-indigo-500 rounded-xl flex items-center justify-center mb-6 mx-auto shadow-lg shadow-blue-500/20">
                <svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              </div>
              <h3 class="text-2xl font-semibold text-white mb-4">Interactive API Explorer</h3>
              <p class="text-slate-300 mb-8 leading-relaxed">
                Browse all API endpoints, see request/response examples, and test the API directly from your browser. 
                Complete documentation for authentication, chat completion, providers, analytics, and admin endpoints.
              </p>
              <a
                href="http://localhost:8084/swagger"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center px-8 py-4 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-base font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20 hover:shadow-xl"
              >
                <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                </svg>
                Open Swagger UI
              </a>
            </div>
            <div class="grid md:grid-cols-2 gap-6 mt-8">
              <div class="bg-slate-700/50 rounded-lg p-6 border border-slate-600/50">
                <h4 class="text-lg font-semibold text-white mb-3">What's Included</h4>
                <ul class="space-y-2 text-sm text-slate-300">
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                    </svg>
                    <span>All API endpoints documented</span>
                  </li>
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                    </svg>
                    <span>Request/response examples</span>
                  </li>
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                    </svg>
                    <span>Try it out functionality</span>
                  </li>
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-green-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                    </svg>
                    <span>Authentication examples</span>
                  </li>
                </ul>
              </div>
              <div class="bg-slate-700/50 rounded-lg p-6 border border-slate-600/50">
                <h4 class="text-lg font-semibold text-white mb-3">Quick Links</h4>
                <ul class="space-y-2 text-sm text-slate-300">
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-blue-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 0l-3 3a1 1 0 001.414 1.414L9 9.414V13a1 1 0 102 0V9.414l1.293 1.293a1 1 0 001.414-1.414z" clip-rule="evenodd" />
                    </svg>
                    <span>Chat Completion API</span>
                  </li>
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-blue-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 0l-3 3a1 1 0 001.414 1.414L9 9.414V13a1 1 0 102 0V9.414l1.293 1.293a1 1 0 001.414-1.414z" clip-rule="evenodd" />
                    </svg>
                    <span>Authentication Endpoints</span>
                  </li>
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-blue-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 0l-3 3a1 1 0 001.414 1.414L9 9.414V13a1 1 0 102 0V9.414l1.293 1.293a1 1 0 001.414-1.414z" clip-rule="evenodd" />
                    </svg>
                    <span>Provider Management</span>
                  </li>
                  <li class="flex items-start">
                    <svg class="w-5 h-5 text-blue-400 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 0l-3 3a1 1 0 001.414 1.414L9 9.414V13a1 1 0 102 0V9.414l1.293 1.293a1 1 0 001.414-1.414z" clip-rule="evenodd" />
                    </svg>
                    <span>Analytics & Metrics</span>
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="py-6 md:py-6 lg:py-6 relative">
      <div class="container mx-auto px-6">
        <div class="max-w-4xl mx-auto text-center bg-slate-800/60 rounded-2xl border border-slate-700/50 p-12 md:p-16 shadow-2xl">
          <h2 class="text-4xl md:text-5xl font-bold text-white mb-6 tracking-tight">
            Ready to build with AI?
          </h2>
          <p class="text-xl md:text-2xl text-slate-300 mb-10">
            Join thousands of developers building the next generation of AI applications
          </p>
          <div class="flex flex-col sm:flex-row gap-4 justify-center">
            <router-link
              to="/register"
              class="px-8 py-4 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-base font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20 hover:shadow-xl"
            >
              Create free account
            </router-link>
            <a
              href="http://localhost:8084/swagger"
              target="_blank"
              rel="noopener noreferrer"
              class="px-8 py-4 bg-slate-700/50 backdrop-blur-sm text-white rounded-lg text-base font-semibold border border-slate-600 hover:border-slate-500 transition-all inline-flex items-center justify-center"
            >
              <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
              View API Docs
            </a>
          </div>
          <p class="mt-8 text-sm text-slate-400">
            No credit card required • Setup in 2 minutes • Free with BYOK or self-hosted
          </p>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer class="bg-slate-950 border-t border-slate-800">
      <div class="container mx-auto px-6 py-12">
        <div class="grid md:grid-cols-4 gap-12 mb-12">
          <div>
            <div class="flex items-center space-x-3 mb-4">
              <div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-indigo-500 rounded-lg flex items-center justify-center">
                <span class="text-white font-bold">U</span>
              </div>
              <span class="text-lg font-bold text-white">UniRoute</span>
            </div>
            <p class="text-slate-400 text-sm leading-relaxed">
              One unified gateway for every AI model. Plus secure tunneling for local services.
            </p>
          </div>
          <div>
            <h4 class="font-semibold text-white mb-4">Product</h4>
            <ul class="space-y-3 text-sm">
              <li><a href="#features" class="text-slate-400 hover:text-white transition-colors">Features</a></li>
              <li><a href="#providers" class="text-slate-400 hover:text-white transition-colors">Providers</a></li>
              <li><a href="#features" class="text-slate-400 hover:text-white transition-colors">Tunneling</a></li>
              <li><a href="#pricing" class="text-slate-400 hover:text-white transition-colors">Pricing</a></li>
              <li><a href="#docs" class="text-slate-400 hover:text-white transition-colors">Documentation</a></li>
            </ul>
          </div>
          <div>
            <h4 class="font-semibold text-white mb-4">Company</h4>
            <ul class="space-y-3 text-sm">
              <li><a href="/about" class="text-slate-400 hover:text-white transition-colors">About</a></li>
              <li><a href="/blog" class="text-slate-400 hover:text-white transition-colors">Blog</a></li>
              <li><a href="/careers" class="text-slate-400 hover:text-white transition-colors">Careers</a></li>
              <li>
                <a 
                  href="https://polar.sh/uniroute/donate" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  class="text-slate-400 hover:text-white transition-colors inline-flex items-center gap-1"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                  </svg>
                  Donate
                </a>
              </li>
            </ul>
          </div>
          <div>
            <h4 class="font-semibold text-white mb-4">Legal</h4>
            <ul class="space-y-3 text-sm">
              <li><router-link to="/privacy" class="text-slate-400 hover:text-white transition-colors">Privacy</router-link></li>
              <li><router-link to="/terms" class="text-slate-400 hover:text-white transition-colors">Terms</router-link></li>
              <li><router-link to="/security" class="text-slate-400 hover:text-white transition-colors">Security</router-link></li>
            </ul>
          </div>
        </div>
        <div class="pt-8 border-t border-slate-800 text-center text-sm text-slate-400">
          <p>&copy; {{ new Date().getFullYear() }} UniRoute. All rights reserved.</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const mobileMenuOpen = ref(false)
const route = useRoute()

onMounted(() => {
  // If there's no hash in the URL, scroll to top
  if (!route.hash) {
    window.scrollTo({ top: 0, behavior: 'instant' })
  }
})
</script>