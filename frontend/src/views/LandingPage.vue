<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-950 overflow-x-hidden relative overscroll-none">
    <!-- Navigation -->
    <nav class="fixed top-0 left-0 right-0 z-[9999] bg-slate-950/80 backdrop-blur-md border-b border-slate-800/50 transition-all duration-300 will-change-auto">
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
            <router-link
              to="/pricing"
              class="text-sm font-medium text-slate-300 hover:text-white transition-colors"
            >
              Pricing
            </router-link>
            <router-link
              to="/docs"
              class="text-sm font-medium text-slate-300 hover:text-white transition-colors"
            >
              Docs
            </router-link>
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
          <router-link
            to="/pricing"
            @click="mobileMenuOpen = false"
            class="block text-sm font-medium text-slate-300 hover:text-white transition-colors py-2"
          >
            Pricing
          </router-link>
          <router-link
            to="/docs"
            @click="mobileMenuOpen = false"
            class="block text-sm font-medium text-slate-300 hover:text-white transition-colors py-2"
          >
            Docs
          </router-link>
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
    <section class="relative overflow-hidden pt-24 pb-16 md:pb-24 min-h-screen flex items-center overscroll-contain">
      <div class="container mx-auto px-6 relative z-10">
        <!-- Main Hero Content -->
        <div class="max-w-6xl mx-auto mb-12">
          <div class="text-center mb-12">
            <h1 class="text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-4 leading-tight tracking-tight">
            The AI Gateway
            <span class="block mt-2 bg-gradient-to-r from-blue-400 via-indigo-400 to-purple-400 bg-clip-text text-transparent">
                & Secure Tunneling
            </span>
          </h1>
            <p class="text-lg md:text-xl text-slate-400 mb-8 max-w-2xl mx-auto">
              One unified API for all AI models. Plus secure tunneling for any service.
            </p>
            <div class="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <router-link
              to="/register"
                class="px-6 py-3 bg-white text-slate-900 rounded-lg text-sm font-semibold hover:bg-slate-100 transition-all shadow-lg hover:shadow-xl w-full sm:w-auto"
            >
                Get started free
            </router-link>
            <router-link
              to="/download"
                class="px-6 py-3 bg-slate-800/50 backdrop-blur-sm text-white rounded-lg text-sm font-semibold border border-slate-700/50 hover:border-slate-600/50 hover:bg-slate-800/70 transition-all w-full sm:w-auto"
            >
              Download CLI
            </router-link>
          </div>
        </div>

          <!-- Two Panel Layout (OutRay Style) -->
          <div class="grid md:grid-cols-2 gap-4 md:gap-6 max-w-5xl mx-auto">
            <!-- Left Panel: Online in one line -->
            <div class="bg-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-800/50 p-4 md:p-6">
              <div class="flex items-center mb-3 md:mb-4">
                <div class="w-7 h-7 md:w-8 md:h-8 bg-purple-500/20 rounded-lg flex items-center justify-center mr-2 md:mr-3 flex-shrink-0">
                  <svg class="w-3.5 h-3.5 md:w-4 md:h-4 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                </div>
                <h2 class="text-base md:text-lg font-bold text-white">Get online instantly</h2>
              </div>
              <p class="text-xs md:text-sm text-slate-400 mb-3 md:mb-4">
                One command for tunneling or AI Gateway. No configuration needed.
              </p>
              
              <!-- Tunneling Command -->
              <div class="bg-slate-950 rounded-lg p-3 md:p-4 border border-slate-800 mb-2 md:mb-3">
                <div class="flex items-center justify-between mb-2">
                  <span class="text-xs text-slate-500">Secure Tunneling</span>
                </div>
                <div class="flex items-center space-x-2 min-w-0">
                  <span class="text-slate-500 text-xs md:text-sm flex-shrink-0">$</span>
                  <code class="text-xs md:text-sm text-white font-mono break-all">uniroute http 8080</code>
                </div>
              </div>

              <!-- AI Gateway Command -->
              <div class="bg-slate-950 rounded-lg p-3 md:p-4 border border-slate-800">
                <div class="flex items-center justify-between mb-2">
                  <span class="text-xs text-slate-500">AI Gateway</span>
                </div>
                <div class="flex items-start space-x-2 min-w-0">
                  <span class="text-slate-500 text-xs md:text-sm flex-shrink-0 mt-0.5">$</span>
                  <div class="flex-1 min-w-0">
                    <code class="text-xs md:text-sm text-white font-mono break-all block">curl https://api.uniroute.co/v1/chat \</code>
                    <code class="text-xs md:text-sm text-slate-400 font-mono break-all block mt-1">-H "Authorization: Bearer ur_key"</code>
                  </div>
                </div>
              </div>
            </div>

            <!-- Right Panel: Instant Observability -->
            <div class="bg-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-800/50 p-4 md:p-6">
              <div class="flex items-center mb-3 md:mb-4">
                <div class="w-7 h-7 md:w-8 md:h-8 bg-purple-500/20 rounded-lg flex items-center justify-center mr-2 md:mr-3 flex-shrink-0">
                  <svg class="w-3.5 h-3.5 md:w-4 md:h-4 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <h2 class="text-base md:text-lg font-bold text-white">Live Request Monitoring</h2>
              </div>
              <p class="text-xs md:text-sm text-slate-400 mb-3 md:mb-4">
                Watch every request flow through your tunnel in real-time as it happens.
              </p>
              
              <!-- Real-time Request List -->
              <div class="bg-slate-950 rounded-lg border border-slate-800 p-2 md:p-4 space-y-1.5 md:space-y-2 max-h-[300px] md:max-h-[400px] overflow-y-auto">
                <div 
                  v-for="(request, index) in liveRequests" 
                  :key="index"
                  class="request-entry flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 px-2 md:px-3 py-1.5 md:py-2 bg-slate-900/50 rounded border border-slate-800/50 hover:border-slate-700/50 transition-all"
                  :class="{ 'animate-slide-in': request.isNew }"
                >
                  <div class="flex items-center space-x-2 md:space-x-3 flex-1 min-w-0">
                    <span 
                      class="text-[10px] md:text-xs font-semibold px-1.5 md:px-2 py-0.5 rounded flex-shrink-0"
                      :class="{
                        'bg-green-900/30 text-green-400': request.status >= 200 && request.status < 300,
                        'bg-yellow-900/30 text-yellow-400': request.status >= 300 && request.status < 400,
                        'bg-red-900/30 text-red-400': request.status >= 400,
                        'bg-purple-900/30 text-purple-400': request.status === 201
                      }"
                    >
                      {{ request.status }} {{ request.statusText }}
                    </span>
                    <span class="text-[10px] md:text-xs text-white font-mono flex-shrink-0">{{ request.method }}</span>
                    <span class="text-[10px] md:text-xs text-slate-300 truncate min-w-0">{{ request.path }}</span>
                  </div>
                  <span class="text-[10px] md:text-xs text-slate-500 flex-shrink-0 sm:ml-4">{{ request.latency }}ms</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section id="features" class="py-20 md:py-24 relative overscroll-contain">
      <div class="container mx-auto px-6">
        <div class="text-center mb-16">
          <h2 class="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            Everything you need
          </h2>
          <p class="text-lg text-slate-400 max-w-2xl mx-auto">
            Enterprise-grade features, zero complexity
          </p>
        </div>
        <div class="max-w-5xl mx-auto">
          <div class="grid md:grid-cols-2 gap-6 md:gap-8">

          <!-- Unified API -->
            <div class="bg-slate-800/30 backdrop-blur-sm rounded-xl p-6 border border-slate-700/30 hover:border-slate-600/50 hover:bg-slate-800/40 transition-all">
              <div class="flex items-start gap-4">
                <div class="w-12 h-12 bg-gradient-to-br from-blue-500 to-indigo-500 rounded-lg flex items-center justify-center flex-shrink-0">
                  <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
              </svg>
            </div>
                <div class="flex-1 min-w-0">
                  <h3 class="text-lg font-semibold text-white mb-2">Unified API</h3>
                  <p class="text-sm text-slate-400 leading-relaxed">
                    One API format for all AI models. Consistent request/response format across all providers.
                  </p>
                </div>
              </div>
          </div>

          <!-- Intelligent Routing -->
            <div class="bg-slate-800/30 backdrop-blur-sm rounded-xl p-6 border border-slate-700/30 hover:border-slate-600/50 hover:bg-slate-800/40 transition-all group relative">
              <div class="flex items-start gap-4">
                <div class="w-12 h-12 bg-gradient-to-br from-purple-500 to-pink-500 rounded-lg flex items-center justify-center flex-shrink-0">
                  <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
                <div class="flex-1 min-w-0">
                  <h3 class="text-lg font-semibold text-white mb-2">Intelligent Routing</h3>
                  <p class="text-sm text-slate-400 leading-relaxed">
                    Automatic model selection based on cost, latency, and availability. Switch providers seamlessly.
                  </p>
                </div>
              </div>
          </div>

          <!-- Multi-Provider Support -->
            <div class="bg-slate-800/30 backdrop-blur-sm rounded-xl p-6 border border-slate-700/30 hover:border-slate-600/50 hover:bg-slate-800/40 transition-all group relative">
              <div class="flex items-start gap-4">
                <div class="w-12 h-12 bg-gradient-to-br from-green-500 to-emerald-500 rounded-lg flex items-center justify-center flex-shrink-0">
                  <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </div>
                <div class="flex-1 min-w-0">
                  <h3 class="text-lg font-semibold text-white mb-2">Multi-Provider</h3>
                  <p class="text-sm text-slate-400 leading-relaxed">
                    Support for OpenAI, Anthropic, Google, and local models. Add new providers easily.
                  </p>
                </div>
              </div>
          </div>

          <!-- Multimodal Support -->
            <div class="bg-slate-800/30 backdrop-blur-sm rounded-xl p-6 border border-slate-700/30 hover:border-slate-600/50 hover:bg-slate-800/40 transition-all group relative">
              <div class="flex items-start gap-4">
                <div class="w-12 h-12 bg-gradient-to-br from-cyan-500 to-blue-500 rounded-lg flex items-center justify-center flex-shrink-0">
                  <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
                <div class="flex-1 min-w-0">
                  <h3 class="text-lg font-semibold text-white mb-2">Multimodal</h3>
                  <p class="text-sm text-slate-400 leading-relaxed">
                    Send multiple images and audio files. Support for vision models and voice transcription.
                  </p>
                </div>
              </div>
          </div>

          <!-- Enterprise Security -->
          <div class="bg-slate-800/30 backdrop-blur-sm rounded-xl p-6 border border-slate-700/30 hover:border-slate-600/50 hover:bg-slate-800/40 transition-all">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-gradient-to-br from-yellow-500 to-orange-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
              <div class="flex-1 min-w-0">
                <h3 class="text-lg font-semibold text-white mb-2">Security</h3>
                <p class="text-sm text-slate-400 leading-relaxed">
                  AES-256 encryption, rate limiting, and comprehensive access control.
                </p>
              </div>
            </div>
          </div>

          <!-- Analytics & Monitoring -->
          <div class="bg-slate-800/30 backdrop-blur-sm rounded-xl p-6 border border-slate-700/30 hover:border-slate-600/50 hover:bg-slate-800/40 transition-all">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-gradient-to-br from-pink-500 to-rose-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
              <div class="flex-1 min-w-0">
                <h3 class="text-lg font-semibold text-white mb-2">Analytics</h3>
                <p class="text-sm text-slate-400 leading-relaxed">
                  Track usage, costs, latency, and performance metrics.
                </p>
              </div>
            </div>
          </div>

          <!-- Built-in Tunneling -->
          <div class="bg-slate-800/30 backdrop-blur-sm rounded-xl p-6 border border-slate-700/30 hover:border-slate-600/50 hover:bg-slate-800/40 transition-all">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-gradient-to-br from-indigo-500 to-blue-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
              </svg>
            </div>
              <div class="flex-1 min-w-0">
                <h3 class="text-lg font-semibold text-white mb-2">Tunneling</h3>
                <p class="text-sm text-slate-400 leading-relaxed">
                  Expose local services with HTTP, TCP, TLS, and UDP support.
                </p>
              </div>
            </div>
          </div>

          <!-- Webhook Testing -->
          <div class="bg-slate-800/30 backdrop-blur-sm rounded-xl p-6 border border-slate-700/30 hover:border-slate-600/50 hover:bg-slate-800/40 transition-all">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-gradient-to-br from-cyan-500 to-blue-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
              </svg>
            </div>
              <div class="flex-1 min-w-0">
                <h3 class="text-lg font-semibold text-white mb-2">Webhook Testing</h3>
                <p class="text-sm text-slate-400 leading-relaxed">
                  Get a secure public URL for your local web server.
                </p>
              </div>
            </div>
          </div>

          <!-- BYOK Model -->
          <div class="bg-slate-800/30 backdrop-blur-sm rounded-xl p-6 border border-slate-700/30 hover:border-slate-600/50 hover:bg-slate-800/40 transition-all">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-gradient-to-br from-emerald-500 to-teal-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
              </svg>
            </div>
              <div class="flex-1 min-w-0">
                <h3 class="text-lg font-semibold text-white mb-2">Bring Your Own Keys</h3>
                <p class="text-sm text-slate-400 leading-relaxed">
                  Securely store your provider API keys with AES-256 encryption.
                </p>
              </div>
            </div>
          </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Unified API Section -->
    <section id="providers" class="py-20 md:py-24 relative overflow-hidden overscroll-contain">
      <!-- Background gradient -->
      <div class="absolute inset-0 bg-gradient-to-br from-blue-500/5 via-purple-500/5 to-indigo-500/5"></div>
      
      <div class="container mx-auto px-6 relative z-10">
        <div class="text-center mb-16">
          <div class="inline-flex items-center gap-2 px-4 py-2 bg-blue-500/10 border border-blue-500/20 rounded-full mb-6">
            <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            <span class="text-sm font-medium text-blue-400">Unified Gateway</span>
          </div>
          <h2 class="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-4 tracking-tight">
            One API,<br/>
            <span class="bg-gradient-to-r from-blue-400 via-indigo-400 to-purple-400 bg-clip-text text-transparent">
              Every AI Model
            </span>
          </h2>
          <p class="text-lg md:text-xl text-slate-400 max-w-3xl mx-auto">
            Switch between providers without changing your code. Same request format, different models.
          </p>
        </div>

        <!-- Unified API Visualization -->
        <div class="max-w-6xl mx-auto mb-16">
          <div class="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800/50 p-8 md:p-12 relative overflow-hidden">
            <!-- Central API -->
            <div class="text-center mb-12">
              <div class="inline-flex items-center gap-3 px-6 py-4 bg-gradient-to-br from-blue-500/20 to-indigo-500/20 border-2 border-blue-400/30 rounded-xl mb-4">
                <svg class="w-8 h-8 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
                </svg>
                <div class="text-left">
                  <div class="text-sm text-slate-400">UniRoute API</div>
                  <div class="text-xl font-bold text-white">POST /v1/chat</div>
            </div>
          </div>
              <p class="text-sm text-slate-400">One endpoint, infinite possibilities</p>
            </div>

            <!-- Provider Grid -->
            <div class="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
              <!-- OpenAI -->
              <div class="group relative">
                <div class="absolute -inset-0.5 bg-gradient-to-r from-blue-500 to-blue-600 rounded-xl opacity-0 group-hover:opacity-20 blur transition-opacity"></div>
                <div class="relative bg-slate-800/50 backdrop-blur-sm rounded-xl p-6 border border-slate-700/50 hover:border-blue-400/50 transition-all h-full">
                  <div class="flex items-center gap-4 mb-4">
                    <div class="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center flex-shrink-0">
                      <span class="text-xl font-bold text-white">O</span>
          </div>
                    <div>
                      <h3 class="text-base font-bold text-white">OpenAI</h3>
                      <p class="text-xs text-slate-400">GPT-4, GPT-3.5</p>
            </div>
          </div>
                  <div class="space-y-2">
                    <div class="flex items-center gap-2 text-xs text-slate-400">
                      <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                      <span>Vision models</span>
            </div>
                    <div class="flex items-center gap-2 text-xs text-slate-400">
                      <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                      <span>Function calling</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Anthropic -->
              <div class="group relative">
                <div class="absolute -inset-0.5 bg-gradient-to-r from-purple-500 to-purple-600 rounded-xl opacity-0 group-hover:opacity-20 blur transition-opacity"></div>
                <div class="relative bg-slate-800/50 backdrop-blur-sm rounded-xl p-6 border border-slate-700/50 hover:border-purple-400/50 transition-all h-full">
                  <div class="flex items-center gap-4 mb-4">
                    <div class="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg flex items-center justify-center flex-shrink-0">
                      <span class="text-xl font-bold text-white">A</span>
                    </div>
                    <div>
                      <h3 class="text-base font-bold text-white">Anthropic</h3>
                      <p class="text-xs text-slate-400">Claude 3, Claude 2</p>
                    </div>
                  </div>
                  <div class="space-y-2">
                    <div class="flex items-center gap-2 text-xs text-slate-400">
                      <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                      <span>Long context</span>
                    </div>
                    <div class="flex items-center gap-2 text-xs text-slate-400">
                      <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                      <span>Constitutional AI</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Google -->
              <div class="group relative">
                <div class="absolute -inset-0.5 bg-gradient-to-r from-green-500 to-green-600 rounded-xl opacity-0 group-hover:opacity-20 blur transition-opacity"></div>
                <div class="relative bg-slate-800/50 backdrop-blur-sm rounded-xl p-6 border border-slate-700/50 hover:border-green-400/50 transition-all h-full">
                  <div class="flex items-center gap-4 mb-4">
                    <div class="w-12 h-12 bg-gradient-to-br from-green-500 to-green-600 rounded-lg flex items-center justify-center flex-shrink-0">
                      <span class="text-xl font-bold text-white">G</span>
                    </div>
                    <div>
                      <h3 class="text-base font-bold text-white">Google</h3>
                      <p class="text-xs text-slate-400">Gemini Pro, Ultra</p>
                    </div>
                  </div>
                  <div class="space-y-2">
                    <div class="flex items-center gap-2 text-xs text-slate-400">
                      <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                      <span>Multimodal</span>
                    </div>
                    <div class="flex items-center gap-2 text-xs text-slate-400">
                      <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                      <span>Code generation</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Local LLMs -->
              <div class="group relative">
                <div class="absolute -inset-0.5 bg-gradient-to-r from-cyan-500 to-cyan-600 rounded-xl opacity-0 group-hover:opacity-20 blur transition-opacity"></div>
                <div class="relative bg-slate-800/50 backdrop-blur-sm rounded-xl p-6 border border-slate-700/50 hover:border-cyan-400/50 transition-all h-full">
                  <div class="flex items-center gap-4 mb-4">
                    <div class="w-12 h-12 bg-gradient-to-br from-cyan-500 to-cyan-600 rounded-lg flex items-center justify-center flex-shrink-0">
                      <span class="text-xl font-bold text-white">L</span>
                    </div>
                    <div>
                      <h3 class="text-base font-bold text-white">Local LLMs</h3>
                      <p class="text-xs text-slate-400">Ollama, vLLM</p>
                    </div>
                  </div>
                  <div class="space-y-2">
                    <div class="flex items-center gap-2 text-xs text-slate-400">
                      <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                      <span>100% free & private</span>
                    </div>
                    <div class="flex items-center gap-2 text-xs text-slate-400">
                      <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                      <span>Offline capable</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Key Benefit -->
        <div class="max-w-4xl mx-auto text-center">
          <div class="inline-flex items-center gap-3 px-6 py-4 bg-slate-800/50 backdrop-blur-sm rounded-xl border border-slate-700/50">
            <svg class="w-6 h-6 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div class="text-left">
              <div class="text-sm font-semibold text-white">Switch providers instantly</div>
              <div class="text-xs text-slate-400">Change the model name, keep the same code</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Use Cases Section -->
    <section class="py-20 md:py-24 relative overflow-hidden overscroll-contain">
      <!-- Background gradient -->
      <div class="absolute inset-0 bg-gradient-to-br from-indigo-500/5 via-purple-500/5 to-pink-500/5"></div>
      
      <div class="container mx-auto px-6 relative z-10">
        <div class="text-center mb-16">
          <div class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-500/10 border border-indigo-500/20 rounded-full mb-6">
            <svg class="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
            <span class="text-sm font-medium text-indigo-400">Developer Experience</span>
          </div>
          <h2 class="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-4 tracking-tight">
            Built for <span class="bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">developers</span>
          </h2>
          <p class="text-lg md:text-xl text-slate-400 max-w-3xl mx-auto">
            Real-world use cases that accelerate your AI development workflow
          </p>
        </div>
        
        <div class="grid md:grid-cols-2 gap-6 max-w-6xl mx-auto">
          <!-- Use Case 1: Multi-Provider -->
          <div class="group relative">
            <div class="absolute -inset-0.5 bg-gradient-to-r from-blue-500 to-indigo-500 rounded-2xl opacity-0 group-hover:opacity-20 blur-xl transition-opacity"></div>
            <div class="relative bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800/50 hover:border-blue-400/30 transition-all h-full">
              <div class="flex items-start gap-5 mb-6">
                <div class="w-14 h-14 bg-gradient-to-br from-blue-500/20 to-indigo-500/20 border-2 border-blue-400/30 rounded-xl flex items-center justify-center flex-shrink-0">
                  <svg class="w-7 h-7 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                </div>
                <div class="flex-1">
                  <h3 class="text-2xl font-bold text-white mb-3">Multi-Provider Apps</h3>
                  <p class="text-slate-400 leading-relaxed mb-5">
                    Build resilient applications that automatically switch between AI providers. No code changes needed when providers go down.
                  </p>
                </div>
              </div>
              <div class="space-y-3">
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-blue-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                  </div>
                  <span class="text-slate-300">One API for all providers</span>
                </div>
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-blue-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                  </div>
                  <span class="text-slate-300">Automatic failover & load balancing</span>
                </div>
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-blue-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
          </div>
                  <span class="text-slate-300">Cost-optimized routing</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Use Case 2: Local Development -->
          <div class="group relative">
            <div class="absolute -inset-0.5 bg-gradient-to-r from-green-500 to-emerald-500 rounded-2xl opacity-0 group-hover:opacity-20 blur-xl transition-opacity"></div>
            <div class="relative bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800/50 hover:border-green-400/30 transition-all h-full">
              <div class="flex items-start gap-5 mb-6">
                <div class="w-14 h-14 bg-gradient-to-br from-green-500/20 to-emerald-500/20 border-2 border-green-400/30 rounded-xl flex items-center justify-center flex-shrink-0">
                  <svg class="w-7 h-7 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                </div>
                <div class="flex-1">
                  <h3 class="text-2xl font-bold text-white mb-3">Local AI Development</h3>
                  <p class="text-slate-400 leading-relaxed mb-5">
                    Develop and test with free local models, then seamlessly deploy to production. Same API, zero code changes.
                  </p>
                </div>
              </div>
              <div class="space-y-3">
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-green-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                  </div>
                  <span class="text-slate-300">Free local models (Ollama, vLLM)</span>
                </div>
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-green-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                  </div>
                  <span class="text-slate-300">100% private & offline capable</span>
                </div>
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-green-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
          </div>
                  <span class="text-slate-300">Seamless production deployment</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Use Case 3: Secure Tunneling -->
          <div class="group relative">
            <div class="absolute -inset-0.5 bg-gradient-to-r from-purple-500 to-pink-500 rounded-2xl opacity-0 group-hover:opacity-20 blur-xl transition-opacity"></div>
            <div class="relative bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800/50 hover:border-purple-400/30 transition-all h-full">
              <div class="flex items-start gap-5 mb-6">
                <div class="w-14 h-14 bg-gradient-to-br from-purple-500/20 to-pink-500/20 border-2 border-purple-400/30 rounded-xl flex items-center justify-center flex-shrink-0">
                  <svg class="w-7 h-7 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <div class="flex-1">
                  <h3 class="text-2xl font-bold text-white mb-3">Secure Tunneling</h3>
                  <p class="text-slate-400 leading-relaxed mb-5">
                    Expose local services instantly. Perfect for webhook testing, API development, and database access with full protocol support.
                  </p>
                </div>
              </div>
              <div class="space-y-3">
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-purple-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-purple-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                  </div>
                  <span class="text-slate-300">HTTP, TCP, TLS, UDP protocols</span>
                </div>
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-purple-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-purple-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                  </div>
                  <span class="text-slate-300">Custom domains with auto SSL</span>
                </div>
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-purple-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-purple-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
          </div>
                  <span class="text-slate-300">One command, instant public URL</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Use Case 4: Enterprise -->
          <div class="group relative">
            <div class="absolute -inset-0.5 bg-gradient-to-r from-orange-500 to-yellow-500 rounded-2xl opacity-0 group-hover:opacity-20 blur-xl transition-opacity"></div>
            <div class="relative bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800/50 hover:border-orange-400/30 transition-all h-full">
              <div class="flex items-start gap-5 mb-6">
                <div class="w-14 h-14 bg-gradient-to-br from-orange-500/20 to-yellow-500/20 border-2 border-orange-400/30 rounded-xl flex items-center justify-center flex-shrink-0">
                  <svg class="w-7 h-7 text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                  </svg>
                </div>
                <div class="flex-1">
                  <h3 class="text-2xl font-bold text-white mb-3">Enterprise Infrastructure</h3>
                  <p class="text-slate-400 leading-relaxed mb-5">
                    Centralized AI gateway for your organization. Complete control with self-hosting or managed service options.
                  </p>
                </div>
              </div>
              <div class="space-y-3">
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-orange-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-orange-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                  </div>
                  <span class="text-slate-300">Centralized API key management</span>
                </div>
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-orange-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-orange-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                  </div>
                  <span class="text-slate-300">Usage analytics & cost tracking</span>
                </div>
                <div class="flex items-center gap-3 text-sm">
                  <div class="w-6 h-6 bg-orange-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg class="w-4 h-4 text-orange-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                  </div>
                  <span class="text-slate-300">Self-hostable or managed</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Tunnel Flow Visualization -->
    <section class="py-12 md:py-16 lg:py-20 relative overflow-hidden overscroll-contain">
      <div class="container mx-auto px-6">
        <!-- Unique Title -->
        <div class="text-center mb-12">
          <h2 class="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-4 tracking-tight">
            How <span class="bg-gradient-to-r from-blue-400 via-indigo-400 to-purple-400 bg-clip-text text-transparent">Tunneling</span> Works
          </h2>
          <p class="text-lg text-slate-400 max-w-2xl mx-auto">
            From your local machine to the global internet in four simple steps
          </p>
        </div>
        
        <!-- Unique Vertical Flow Design -->
        <div class="max-w-4xl mx-auto">
          <!-- Flow Steps -->
          <div class="space-y-8 relative">
            <!-- Connecting Line (Vertical) -->
            <div class="hidden md:block absolute left-8 top-0 bottom-0 w-0.5 bg-gradient-to-b from-yellow-400 via-green-400 via-purple-400 to-blue-400 opacity-30"></div>
            
            <!-- Step 1: Localhost -->
            <div class="relative flex items-start gap-6">
              <div class="flex-shrink-0 relative z-10">
                <div class="w-16 h-16 rounded-xl bg-gradient-to-br from-yellow-500/20 to-yellow-600/20 border-2 border-yellow-400/50 flex items-center justify-center backdrop-blur-sm">
                  <div class="absolute inset-0 bg-yellow-500/10 rounded-xl blur-lg"></div>
                  <svg class="w-8 h-8 text-yellow-400 relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
            </div>
                <div class="hidden md:block absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-4 h-4 bg-yellow-400 rounded-full border-4 border-slate-900 z-20"></div>
              </div>
              <div class="flex-1 pt-2">
                <div class="inline-flex items-center gap-2 mb-2">
                  <span class="text-xs font-bold text-yellow-400 bg-yellow-400/10 px-2 py-1 rounded">STEP 1</span>
                  <h3 class="text-xl font-bold text-white">Localhost</h3>
                </div>
                <p class="text-slate-400 mb-2">
                  Your application running locally
                </p>
                <code class="text-sm text-yellow-400 font-mono bg-slate-900/50 px-3 py-1.5 rounded border border-slate-800">localhost:8080</code>
          </div>
            </div>

            <!-- Step 2: Secure Tunnel -->
            <div class="relative flex items-start gap-6">
              <div class="flex-shrink-0 relative z-10">
                <div class="w-16 h-16 rounded-xl bg-gradient-to-br from-green-500/20 to-green-600/20 border-2 border-green-400/50 flex items-center justify-center backdrop-blur-sm">
                  <div class="absolute inset-0 bg-green-500/10 rounded-xl blur-lg"></div>
                  <svg class="w-8 h-8 text-green-400 relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                  </svg>
                </div>
                <div class="hidden md:block absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-4 h-4 bg-green-400 rounded-full border-4 border-slate-900 z-20"></div>
              </div>
              <div class="flex-1 pt-2">
                <div class="inline-flex items-center gap-2 mb-2">
                  <span class="text-xs font-bold text-green-400 bg-green-400/10 px-2 py-1 rounded">STEP 2</span>
                  <h3 class="text-xl font-bold text-white">Secure Tunnel</h3>
                </div>
                <p class="text-slate-400 mb-2">
                  Creates an encrypted bridge that survives restarts
                </p>
                <code class="text-sm text-green-400 font-mono bg-slate-900/50 px-3 py-1.5 rounded border border-slate-800">uniroute http 8080</code>
          </div>
            </div>

            <!-- Step 3: UniRoute Edge -->
            <div class="relative flex items-start gap-6">
              <div class="flex-shrink-0 relative z-10">
                <div class="w-16 h-16 rounded-xl bg-gradient-to-br from-purple-500/20 to-purple-600/20 border-2 border-purple-400/50 flex items-center justify-center backdrop-blur-sm">
                  <div class="absolute inset-0 bg-purple-500/10 rounded-xl blur-lg"></div>
                  <svg class="w-8 h-8 text-purple-400 relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                  </svg>
                </div>
                <div class="hidden md:block absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-4 h-4 bg-purple-400 rounded-full border-4 border-slate-900 z-20"></div>
              </div>
              <div class="flex-1 pt-2">
                <div class="inline-flex items-center gap-2 mb-2">
                  <span class="text-xs font-bold text-purple-400 bg-purple-400/10 px-2 py-1 rounded">STEP 3</span>
                  <h3 class="text-xl font-bold text-white">UniRoute Edge</h3>
                </div>
                <p class="text-slate-400 mb-2">
                  Our global network forwards requests to your service
                </p>
                <code class="text-sm text-purple-400 font-mono bg-slate-900/50 px-3 py-1.5 rounded border border-slate-800">tunnel.uniroute.co</code>
          </div>
            </div>

            <!-- Step 4: Public Internet -->
            <div class="relative flex items-start gap-6">
              <div class="flex-shrink-0 relative z-10">
                <div class="w-16 h-16 rounded-xl bg-gradient-to-br from-blue-500/20 to-blue-600/20 border-2 border-blue-400/50 flex items-center justify-center backdrop-blur-sm">
                  <div class="absolute inset-0 bg-blue-500/10 rounded-xl blur-lg"></div>
                  <svg class="w-8 h-8 text-blue-400 relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div class="hidden md:block absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-4 h-4 bg-blue-400 rounded-full border-4 border-slate-900 z-20"></div>
              </div>
              <div class="flex-1 pt-2">
                <div class="inline-flex items-center gap-2 mb-2">
                  <span class="text-xs font-bold text-blue-400 bg-blue-400/10 px-2 py-1 rounded">STEP 4</span>
                  <h3 class="text-xl font-bold text-white">Public Internet</h3>
                </div>
                <p class="text-slate-400 mb-2">
                  Instantly available globally via
                </p>
                <code class="text-sm text-blue-400 font-mono bg-slate-900/50 px-3 py-1.5 rounded border border-slate-800">{subdomain}.uniroute.co</code>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Custom Domain Support -->
    <section class="py-16 md:py-20 lg:py-24 relative overflow-hidden overscroll-contain">
      <!-- Background effects -->
      <div class="absolute inset-0 bg-gradient-to-br from-indigo-500/5 via-purple-500/5 to-pink-500/5"></div>
      <div class="absolute top-0 left-1/4 w-96 h-96 bg-indigo-500/10 rounded-full blur-3xl"></div>
      <div class="absolute bottom-0 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl"></div>
      
      <div class="container mx-auto px-6 relative z-10">
        <div class="max-w-6xl mx-auto">
          <!-- Header -->
          <div class="text-center mb-12">
            <div class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-500/10 border border-indigo-500/20 rounded-full mb-6">
              <svg class="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
              </svg>
              <span class="text-sm font-medium text-indigo-400">Domain Management</span>
            </div>
            <h2 class="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-4 tracking-tight">
              Your domain, <span class="bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">your brand</span>
          </h2>
            <p class="text-lg md:text-xl text-slate-400 max-w-3xl mx-auto">
              Replace random subdomains with your own domain. Set it up in moments, keep it forever.
          </p>
        </div>

          <!-- Main Content: Split Layout -->
          <div class="grid lg:grid-cols-2 gap-8 lg:gap-12 items-center">
            <!-- Left: Features & Benefits -->
            <div class="space-y-6">
              <div class="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800/50 p-6 hover:border-indigo-500/50 transition-all">
                <div class="flex items-start gap-4">
                  <div class="flex-shrink-0 w-12 h-12 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 border border-green-500/30 flex items-center justify-center">
                    <svg class="w-6 h-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                    </svg>
              </div>
                  <div>
                    <h3 class="text-lg font-semibold text-white mb-2">Managed Certificates</h3>
                    <p class="text-slate-400 text-sm leading-relaxed">
                      Certificates issued and renewed automatically. HTTPS enabled by default, no manual setup needed.
                    </p>
                  </div>
                  </div>
                  </div>

              <div class="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800/50 p-6 hover:border-purple-500/50 transition-all">
                <div class="flex items-start gap-4">
                  <div class="flex-shrink-0 w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 border border-purple-500/30 flex items-center justify-center">
                    <svg class="w-6 h-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                    </svg>
                  </div>
                  <div>
                    <h3 class="text-lg font-semibold text-white mb-2">Simple DNS Setup</h3>
                    <p class="text-slate-400 text-sm leading-relaxed">
                      Add one CNAME record in your DNS provider. Works with any registrar or DNS service.
                    </p>
                </div>
              </div>
            </div>

              <div class="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800/50 p-6 hover:border-blue-500/50 transition-all">
                <div class="flex items-start gap-4">
                  <div class="flex-shrink-0 w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 border border-blue-500/30 flex items-center justify-center">
                    <svg class="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
                  <div>
                    <h3 class="text-lg font-semibold text-white mb-2">Zero Downtime</h3>
                    <p class="text-slate-400 text-sm leading-relaxed">
                      Switch domains or update configurations seamlessly. Your tunnel stays online throughout.
                    </p>
              </div>
              </div>
              </div>

              <!-- Command Example -->
              <div class="mt-8 bg-slate-900/70 rounded-xl border border-slate-800 p-5">
                <div class="flex items-center gap-2 mb-3">
                  <div class="w-2 h-2 bg-red-500 rounded-full"></div>
                  <div class="w-2 h-2 bg-yellow-500 rounded-full"></div>
                  <div class="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span class="ml-2 text-xs text-slate-500">Terminal</span>
                </div>
                <div class="space-y-2">
                  <div class="flex items-center gap-2">
                    <span class="text-slate-500 text-sm">$</span>
                    <code class="text-sm text-indigo-400 font-mono">uniroute domain webhook.mysite.io</code>
                  </div>
                  <div class="text-xs text-slate-500 mt-3 pl-4 border-l-2 border-indigo-500/30">
                    Link your domain with a single command
                  </div>
                </div>
              </div>
            </div>

            <!-- Right: Visual Domain Configuration -->
            <div class="relative">
              <div class="bg-gradient-to-br from-slate-900/90 to-slate-800/90 backdrop-blur-sm rounded-2xl border-2 border-slate-700/50 p-6 md:p-8 shadow-2xl relative overflow-hidden">
                <!-- Animated background gradient -->
                <div class="absolute inset-0 bg-gradient-to-br from-indigo-500/10 via-purple-500/10 to-pink-500/10 opacity-50"></div>
                
                <div class="relative z-10">
                  <!-- Domain Header -->
                  <div class="flex items-center justify-between mb-6 pb-4 border-b border-slate-700/50">
                    <div class="flex items-center gap-3">
                      <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center">
                        <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                </svg>
              </div>
                      <div>
                        <h4 class="text-lg font-bold text-white">webhook.mysite.io</h4>
                        <p class="text-xs text-slate-500">Your domain</p>
                      </div>
                    </div>
                    <div class="px-3 py-1.5 bg-green-500/20 border border-green-500/30 rounded-lg">
                      <span class="text-xs font-medium text-green-400 flex items-center gap-1.5">
                        <span class="w-1.5 h-1.5 bg-green-400 rounded-full animate-pulse"></span>
                        Active
                      </span>
            </div>
          </div>

                  <!-- DNS Configuration Card -->
                  <div class="bg-slate-800/50 rounded-xl border border-slate-700/50 p-5 mb-4">
                    <h5 class="text-sm font-semibold text-slate-300 mb-4 flex items-center gap-2">
                      <svg class="w-4 h-4 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                      </svg>
                      DNS Configuration
                    </h5>
                    <div class="space-y-3">
                      <div class="flex items-center justify-between py-2 border-b border-slate-700/30">
                        <span class="text-xs text-slate-500 uppercase tracking-wide">Type</span>
                        <code class="text-sm text-indigo-400 font-mono bg-indigo-500/10 px-2 py-1 rounded">CNAME</code>
              </div>
                      <div class="flex items-center justify-between py-2 border-b border-slate-700/30">
                        <span class="text-xs text-slate-500 uppercase tracking-wide">Name</span>
                        <code class="text-sm text-white font-mono">webhook</code>
              </div>
                      <div class="flex items-center justify-between py-2 border-b border-slate-700/30">
                        <span class="text-xs text-slate-500 uppercase tracking-wide">Value</span>
                        <code class="text-sm text-purple-400 font-mono bg-purple-500/10 px-2 py-1 rounded">tunnel.uniroute.co</code>
            </div>
                      <div class="flex items-center justify-between py-2">
                        <span class="text-xs text-slate-500 uppercase tracking-wide">TTL</span>
                        <span class="text-sm text-slate-400">Auto</span>
                      </div>
                    </div>
                  </div>

                  <!-- Status Info -->
                  <div class="flex items-start gap-3 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
                    <svg class="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                    <div>
                      <p class="text-xs font-medium text-blue-400 mb-1">Certificate Ready</p>
                      <p class="text-xs text-slate-400">HTTPS enabled and renewing automatically</p>
              </div>
                  </div>
                </div>
              </div>

              <!-- Decorative elements -->
              <div class="absolute -top-4 -right-4 w-24 h-24 bg-indigo-500/20 rounded-full blur-2xl"></div>
              <div class="absolute -bottom-4 -left-4 w-32 h-32 bg-purple-500/20 rounded-full blur-2xl"></div>
            </div>
          </div>

          <!-- Bottom CTA -->
          <div class="mt-12 text-center">
            <p class="text-slate-400 mb-4">Want to connect your domain?</p>
            <router-link
              to="/docs/tunnels/custom-domains"
              class="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-indigo-500 to-purple-500 text-white rounded-xl font-semibold hover:from-indigo-600 hover:to-purple-600 transition-all shadow-lg hover:shadow-xl transform hover:scale-105"
            >
              <span>Read the Guide</span>
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                </svg>
            </router-link>
              </div>
        </div>
      </div>
    </section>


    <!-- Get Started Section -->
    <section class="py-20 md:py-24 relative overflow-hidden overscroll-contain">
      <!-- Background gradient -->
      <div class="absolute inset-0 bg-gradient-to-br from-blue-500/5 via-indigo-500/5 to-purple-500/5"></div>
      
      <div class="container mx-auto px-6 relative z-10">
        <div class="text-center mb-16">
          <div class="inline-flex items-center gap-2 px-4 py-2 bg-blue-500/10 border border-blue-500/20 rounded-full mb-6">
            <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
            <span class="text-sm font-medium text-blue-400">Quick Start</span>
              </div>
          <h2 class="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-4 tracking-tight">
            Get started in <span class="bg-gradient-to-r from-blue-400 via-indigo-400 to-purple-400 bg-clip-text text-transparent">minutes</span>
          </h2>
          <p class="text-lg md:text-xl text-slate-400 max-w-3xl mx-auto">
            From signup to production in four simple steps
          </p>
        </div>

        <!-- Steps with connecting line -->
        <div class="max-w-5xl mx-auto relative">
          <!-- Connecting line (hidden on mobile) -->
          <div class="hidden md:block absolute top-24 left-0 right-0 h-0.5 bg-gradient-to-r from-blue-500 via-indigo-500 via-purple-500 to-pink-500 opacity-20"></div>
          
          <div class="grid md:grid-cols-4 gap-8 md:gap-4 relative">
            <!-- Step 1 -->
            <div class="relative group">
              <div class="flex flex-col items-center text-center">
                <div class="relative mb-6">
                  <div class="absolute -inset-2 bg-blue-500/20 rounded-full blur-xl opacity-0 group-hover:opacity-100 transition-opacity"></div>
                  <div class="relative w-20 h-20 bg-gradient-to-br from-blue-500 to-indigo-500 rounded-2xl flex items-center justify-center shadow-lg shadow-blue-500/30 transform group-hover:scale-110 transition-transform">
                    <svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              </div>
                  <div class="absolute -top-2 -right-2 w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-sm font-bold text-white border-4 border-slate-900">
                    1
                  </div>
                </div>
                <h3 class="text-xl font-bold text-white mb-3">Sign up free</h3>
                <p class="text-slate-400 leading-relaxed text-sm">
                  Create your account in seconds. No credit card required.
                </p>
            </div>
          </div>

            <!-- Step 2 -->
            <div class="relative group">
              <div class="flex flex-col items-center text-center">
                <div class="relative mb-6">
                  <div class="absolute -inset-2 bg-indigo-500/20 rounded-full blur-xl opacity-0 group-hover:opacity-100 transition-opacity"></div>
                  <div class="relative w-20 h-20 bg-gradient-to-br from-indigo-500 to-purple-500 rounded-2xl flex items-center justify-center shadow-lg shadow-indigo-500/30 transform group-hover:scale-110 transition-transform">
                    <svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                    </svg>
              </div>
                  <div class="absolute -top-2 -right-2 w-8 h-8 bg-indigo-500 rounded-full flex items-center justify-center text-sm font-bold text-white border-4 border-slate-900">
                    2
              </div>
                </div>
                <h3 class="text-xl font-bold text-white mb-3">Add API keys</h3>
                <p class="text-slate-400 leading-relaxed text-sm">
                  Configure providers or use local models. All encrypted and secure.
              </p>
            </div>
              </div>

            <!-- Step 3 -->
            <div class="relative group">
              <div class="flex flex-col items-center text-center">
                <div class="relative mb-6">
                  <div class="absolute -inset-2 bg-purple-500/20 rounded-full blur-xl opacity-0 group-hover:opacity-100 transition-opacity"></div>
                  <div class="relative w-20 h-20 bg-gradient-to-br from-purple-500 to-pink-500 rounded-2xl flex items-center justify-center shadow-lg shadow-purple-500/30 transform group-hover:scale-110 transition-transform">
                    <svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
                  <div class="absolute -top-2 -right-2 w-8 h-8 bg-purple-500 rounded-full flex items-center justify-center text-sm font-bold text-white border-4 border-slate-900">
                    3
              </div>
              </div>
                <h3 class="text-xl font-bold text-white mb-3">Start routing</h3>
                <p class="text-slate-400 leading-relaxed text-sm">
                  Use unified API. We handle security, failover, and analytics automatically.
                </p>
            </div>
          </div>

            <!-- Step 4 -->
            <div class="relative group">
              <div class="flex flex-col items-center text-center">
                <div class="relative mb-6">
                  <div class="absolute -inset-2 bg-pink-500/20 rounded-full blur-xl opacity-0 group-hover:opacity-100 transition-opacity"></div>
                  <div class="relative w-20 h-20 bg-gradient-to-br from-pink-500 to-rose-500 rounded-2xl flex items-center justify-center shadow-lg shadow-pink-500/30 transform group-hover:scale-110 transition-transform">
                    <svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
        </div>
                  <div class="absolute -top-2 -right-2 w-8 h-8 bg-pink-500 rounded-full flex items-center justify-center text-sm font-bold text-white border-4 border-slate-900">
                    4
                  </div>
                </div>
                <h3 class="text-xl font-bold text-white mb-3">Expose services</h3>
                <p class="text-slate-400 leading-relaxed text-sm">
                  Tunnel local services instantly. HTTP, TCP, TLS, UDP with custom domains.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- CTA Section -->
        <div class="mt-16 text-center">
          <div class="inline-flex flex-col sm:flex-row gap-4 items-center">
              <router-link
                to="/register"
              class="px-8 py-4 bg-white text-slate-900 rounded-xl text-base font-semibold hover:bg-slate-100 transition-all shadow-lg hover:shadow-xl transform hover:scale-105"
              >
                Get started free
              </router-link>
            <router-link
              to="/download"
              class="px-8 py-4 bg-slate-800/50 backdrop-blur-sm text-white rounded-xl text-base font-semibold border-2 border-slate-700/50 hover:border-slate-600/50 hover:bg-slate-800/70 transition-all"
            >
              Download CLI
            </router-link>
          </div>
          <p class="mt-6 text-sm text-slate-500">
            No credit card required • Free with BYOK or self-hosted
          </p>
        </div>
      </div>
    </section>

    <!-- Open Source Section -->
    <section class="py-20 md:py-24 lg:py-28 relative overflow-hidden overscroll-contain">
      <!-- Background effects -->
      <div class="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-teal-500/5 to-cyan-500/5"></div>
      <div class="absolute top-1/4 right-0 w-96 h-96 bg-emerald-500/10 rounded-full blur-3xl"></div>
      <div class="absolute bottom-1/4 left-0 w-96 h-96 bg-cyan-500/10 rounded-full blur-3xl"></div>
      
      <div class="container mx-auto px-6 relative z-10">
        <div class="max-w-6xl mx-auto">
          <!-- Header -->
          <div class="text-center mb-16">
            <div class="inline-flex items-center gap-2 px-4 py-2 bg-emerald-500/10 border border-emerald-500/20 rounded-full mb-6">
              <svg class="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
              </svg>
              <span class="text-sm font-medium text-emerald-400">Open Source</span>
            </div>
            <h2 class="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-4 tracking-tight">
              Built in the <span class="bg-gradient-to-r from-emerald-400 via-teal-400 to-cyan-400 bg-clip-text text-transparent">open</span>
          </h2>
            <p class="text-lg md:text-xl text-slate-400 max-w-3xl mx-auto">
              Every line of code is visible, every feature is auditable. Run it yourself or contribute to make it better.
          </p>
        </div>

          <!-- Main Content: Unique Layout -->
          <div class="grid lg:grid-cols-3 gap-8 mb-12">
            <!-- Left: Code Preview -->
            <div class="lg:col-span-2">
              <div class="bg-slate-900/80 backdrop-blur-sm rounded-2xl border-2 border-slate-800/50 p-6 md:p-8 shadow-2xl relative overflow-hidden">
                <!-- Terminal header -->
                <div class="flex items-center gap-2 mb-6 pb-4 border-b border-slate-700/50">
                  <div class="flex gap-2">
                    <div class="w-3 h-3 rounded-full bg-red-500"></div>
                    <div class="w-3 h-3 rounded-full bg-yellow-500"></div>
                    <div class="w-3 h-3 rounded-full bg-green-500"></div>
              </div>
                  <span class="ml-4 text-sm text-slate-500 font-mono">uniroute/main.go</span>
                  <div class="ml-auto flex items-center gap-2">
                    <svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                    <span class="text-xs text-slate-500">MIT License</span>
            </div>
                </div>
                
                <!-- Code snippet -->
                <div class="space-y-4">
                  <div class="flex items-start gap-4">
                    <span class="text-slate-600 text-sm font-mono mt-1">42</span>
                    <code class="text-sm text-slate-300 flex-1">
                      <span class="text-purple-400">func</span> <span class="text-blue-400">StartTunnel</span>(<span class="text-cyan-400">port</span> <span class="text-yellow-400">int</span>) <span class="text-purple-400">error</span> {
                    </code>
                  </div>
                  <div class="flex items-start gap-4">
                    <span class="text-slate-600 text-sm font-mono mt-1">43</span>
                    <code class="text-sm text-slate-300 flex-1">
                      &nbsp;&nbsp;<span class="text-slate-500">// Secure connection with auto-reconnect</span>
                    </code>
                  </div>
                  <div class="flex items-start gap-4">
                    <span class="text-slate-600 text-sm font-mono mt-1">44</span>
                    <code class="text-sm text-slate-300 flex-1">
                      &nbsp;&nbsp;<span class="text-purple-400">return</span> <span class="text-blue-400">client</span>.<span class="text-blue-400">Connect</span>(<span class="text-cyan-400">port</span>)
                    </code>
                  </div>
                  <div class="flex items-start gap-4">
                    <span class="text-slate-600 text-sm font-mono mt-1">45</span>
                    <code class="text-sm text-slate-300 flex-1">
                      }
                    </code>
                  </div>
                </div>

                <!-- Stats bar -->
                <div class="mt-6 pt-6 border-t border-slate-700/50 flex items-center justify-between text-xs">
                  <div class="flex items-center gap-4">
                    <div class="flex items-center gap-2">
                      <svg class="w-4 h-4 text-emerald-400" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                    </svg>
                      <span class="text-slate-400">Public Repository</span>
                    </div>
                    <div class="flex items-center gap-2">
                      <svg class="w-4 h-4 text-cyan-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                    </svg>
                      <span class="text-slate-400">MIT License</span>
                    </div>
                  </div>
                  <div class="text-slate-500">View on GitHub →</div>
                </div>
              </div>
            </div>

            <!-- Right: Benefits List -->
            <div class="space-y-4">
              <div class="bg-gradient-to-br from-slate-900/60 to-slate-800/60 backdrop-blur-sm rounded-xl border border-emerald-500/20 p-6 hover:border-emerald-500/40 transition-all group">
                <div class="flex items-start gap-4">
                  <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-emerald-500/20 border border-emerald-500/30 flex items-center justify-center group-hover:bg-emerald-500/30 transition-colors">
                    <svg class="w-6 h-6 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                    </svg>
                  </div>
                  <div>
                    <h3 class="text-base font-semibold text-white mb-1.5">Code Transparency</h3>
                    <p class="text-sm text-slate-400 leading-relaxed">
                      Review every implementation detail. No hidden logic, no black boxes.
                    </p>
                  </div>
                </div>
              </div>

              <div class="bg-gradient-to-br from-slate-900/60 to-slate-800/60 backdrop-blur-sm rounded-xl border border-teal-500/20 p-6 hover:border-teal-500/40 transition-all group">
                <div class="flex items-start gap-4">
                  <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-teal-500/20 border border-teal-500/30 flex items-center justify-center group-hover:bg-teal-500/30 transition-colors">
                    <svg class="w-6 h-6 text-teal-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
              </div>
                  <div>
                    <h3 class="text-base font-semibold text-white mb-1.5">Deploy Anywhere</h3>
                    <p class="text-sm text-slate-400 leading-relaxed">
                      Host on your infrastructure. Full control, no dependencies on our servers.
                    </p>
                  </div>
                </div>
              </div>

              <div class="bg-gradient-to-br from-slate-900/60 to-slate-800/60 backdrop-blur-sm rounded-xl border border-cyan-500/20 p-6 hover:border-cyan-500/40 transition-all group">
                <div class="flex items-start gap-4">
                  <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-cyan-500/20 border border-cyan-500/30 flex items-center justify-center group-hover:bg-cyan-500/30 transition-colors">
                    <svg class="w-6 h-6 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                    </svg>
                  </div>
                  <div>
                    <h3 class="text-base font-semibold text-white mb-1.5">Community Built</h3>
                    <p class="text-sm text-slate-400 leading-relaxed">
                      Shaped by developers worldwide. Your ideas and contributions drive innovation.
                    </p>
                  </div>
                </div>
              </div>

              <div class="bg-gradient-to-br from-slate-900/60 to-slate-800/60 backdrop-blur-sm rounded-xl border border-emerald-500/20 p-6 hover:border-emerald-500/40 transition-all group">
                <div class="flex items-start gap-4">
                  <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-emerald-500/20 border border-emerald-500/30 flex items-center justify-center group-hover:bg-emerald-500/30 transition-colors">
                    <svg class="w-6 h-6 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
              </div>
                  <div>
                    <h3 class="text-base font-semibold text-white mb-1.5">Always Free</h3>
                    <p class="text-sm text-slate-400 leading-relaxed">
                      Core functionality remains free forever. No paywalls, no feature restrictions.
                    </p>
            </div>
          </div>
        </div>
      </div>
          </div>

          <!-- CTA Buttons -->
          <div class="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <a
              href="https://github.com/uniroute/uniroute"
              target="_blank"
              rel="noopener noreferrer"
              class="group px-8 py-4 bg-slate-800/50 backdrop-blur-sm text-white rounded-xl font-semibold border-2 border-slate-700/50 hover:border-emerald-500/50 hover:bg-slate-800/70 transition-all flex items-center gap-3"
            >
              <svg class="w-5 h-5 text-emerald-400 group-hover:scale-110 transition-transform" fill="currentColor" viewBox="0 0 24 24">
                <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-1.04-.013-1.845-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482C19.138 20.197 22 16.442 22 12.017 22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
              </svg>
              <span>View Source Code</span>
            </a>
            <a
              href="https://discord.gg/uniroute"
              target="_blank"
              rel="noopener noreferrer"
              class="px-8 py-4 bg-gradient-to-r from-emerald-500 to-teal-500 text-white rounded-xl font-semibold hover:from-emerald-600 hover:to-teal-600 transition-all shadow-lg hover:shadow-xl transform hover:scale-105 flex items-center gap-3"
            >
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M20.317 4.37a19.791 19.791 0 00-4.885-1.515.074.074 0 00-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 00-5.487 0 12.64 12.64 0 00-.617-1.25.077.077 0 00-.079-.037A19.736 19.736 0 003.677 4.37a.07.07 0 00-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 00.031.057 19.9 19.9 0 005.993 3.03.078.078 0 00.084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 00-.041-.106 13.107 13.107 0 01-1.872-.892.077.077 0 01-.008-.128 10.2 10.2 0 00.372-.292.074.074 0 01.077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 01.078.01c.12.098.246.198.373.292a.077.077 0 01-.006.127 12.299 12.299 0 01-1.873.892.077.077 0 00-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 00.084.028 19.839 19.839 0 006.002-3.03.077.077 0 00.032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 00-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"/>
              </svg>
              <span>Join Our Community</span>
            </a>
          </div>
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
              <li><router-link to="/pricing" class="text-slate-400 hover:text-white transition-colors">Pricing</router-link></li>
            </ul>
          </div>
          <div>
            <h4 class="font-semibold text-white mb-4">Company</h4>
            <ul class="space-y-3 text-sm">
              <li><a href="/about" class="text-slate-400 hover:text-white transition-colors">About</a></li>
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
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'

const mobileMenuOpen = ref(false)
const route = useRoute()

// Real-time request simulation
interface LiveRequest {
  method: string
  path: string
  status: number
  statusText: string
  latency: number
  isNew: boolean
}

const liveRequests = ref<LiveRequest[]>([])
const maxRequests = 8

// Sample request templates - UniRoute endpoints
const requestTemplates = [
  { method: 'POST', path: '/v1/chat', status: 200, statusText: 'OK', latency: 245 },
  { method: 'POST', path: '/v1/chat/stream', status: 200, statusText: 'OK', latency: 189 },
  { method: 'GET', path: '/auth/tunnels', status: 200, statusText: 'OK', latency: 22 },
  { method: 'GET', path: '/auth/analytics/requests', status: 200, statusText: 'OK', latency: 45 },
  { method: 'POST', path: '/auth/conversations', status: 201, statusText: 'Created', latency: 32 },
  { method: 'GET', path: '/auth/conversations', status: 200, statusText: 'OK', latency: 18 },
  { method: 'PUT', path: '/auth/routing/strategy', status: 200, statusText: 'OK', latency: 24 },
  { method: 'GET', path: '/auth/routing/custom-rules', status: 200, statusText: 'OK', latency: 15 },
  { method: 'POST', path: '/v1/chat', status: 200, statusText: 'OK', latency: 312 },
  { method: 'GET', path: '/auth/tunnels/stats', status: 200, statusText: 'OK', latency: 12 },
  { method: 'POST', path: '/v1/chat/stream', status: 200, statusText: 'OK', latency: 156 },
  { method: 'GET', path: '/auth/analytics/usage', status: 200, statusText: 'OK', latency: 28 },
]

let requestInterval: ReturnType<typeof setInterval> | null = null

const addNewRequest = () => {
  const template = requestTemplates[Math.floor(Math.random() * requestTemplates.length)]
  const newRequest: LiveRequest = {
    ...template,
    latency: template.latency + Math.floor(Math.random() * 20) - 10, // Add some variance
    isNew: true
  }
  
  liveRequests.value.unshift(newRequest)
  
  // Remove isNew flag after animation
  setTimeout(() => {
    if (liveRequests.value[0]) {
      liveRequests.value[0].isNew = false
    }
  }, 500)
  
  // Keep only last N requests
  if (liveRequests.value.length > maxRequests) {
    liveRequests.value = liveRequests.value.slice(0, maxRequests)
  }
}

// Navbar scroll effect
const handleScroll = () => {
  const nav = document.querySelector('nav')
  if (nav) {
    if (window.scrollY > 50) {
      nav.classList.add('shadow-2xl')
      nav.style.borderBottomColor = 'rgba(59, 130, 246, 0.2)'
    } else {
      nav.classList.remove('shadow-2xl')
      nav.style.borderBottomColor = 'rgba(51, 65, 85, 0.5)'
    }
  }
}

onMounted(() => {
  // Ensure all content is visible immediately (no scroll reveal animations)
  // This prevents the "opening" effect when scrolling
  const allSections = document.querySelectorAll('section')
  allSections.forEach(section => {
    section.style.opacity = '1'
    section.style.transform = 'none'
    // Remove any scroll-reveal classes
    section.classList.remove('scroll-reveal')
    section.classList.add('revealed')
  })
  
  // Also ensure all direct children are visible
  const allContent = document.querySelectorAll('section > *')
  allContent.forEach(el => {
    (el as HTMLElement).style.opacity = '1'
    ;(el as HTMLElement).style.transform = 'none'
  })
  
  // If there's no hash in the URL, scroll to top
  if (!route.hash) {
    window.scrollTo({ top: 0, behavior: 'instant' })
  }

  // Initialize with a few requests
  for (let i = 0; i < 3; i++) {
    setTimeout(() => addNewRequest(), i * 300)
  }

  // Start real-time simulation - add new request every 2-4 seconds
  requestInterval = setInterval(() => {
    addNewRequest()
  }, 2000 + Math.random() * 2000)

  // Scroll reveal disabled for better scroll performance
  // Content is visible immediately without animations

  // Setup scroll handler for navbar with throttling for smooth performance
  let ticking = false
  const optimizedScroll = () => {
    if (!ticking) {
      window.requestAnimationFrame(() => {
        handleScroll()
        ticking = false
      })
      ticking = true
    }
  }
  
  window.addEventListener('scroll', optimizedScroll, { passive: true })
  handleScroll() // Initial call
  
  // Prevent white flash on initial load
  document.body.style.backgroundColor = '#020617'
  document.documentElement.style.backgroundColor = '#020617'
})

onUnmounted(() => {
  if (requestInterval) {
    clearInterval(requestInterval)
  }
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped>
/* Background colors - scroll-smooth is handled by Tailwind in main.css */
:deep(html) {
  background-color: #020617; /* slate-950 */
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  /* Prevent scroll chaining between sections */
  overscroll-behavior-y: none;
}

:deep(body) {
  background-color: #020617; /* slate-950 */
  overflow-x: hidden;
  margin: 0;
  padding: 0;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  /* Prevent scroll chaining - prevents sections from interfering */
  overscroll-behavior-y: none;
  /* Don't use contain - it causes layout issues with fixed nav */
}

/* Ensure no white background shows through */
:deep(*) {
  box-sizing: border-box;
}

/* Optimize scroll performance - prevent content "opening" animations */
/* Ensure all main content containers are visible immediately */
:deep(section) {
  opacity: 1 !important;
  visibility: visible !important;
  transform: none !important;
}

:deep(.container),
:deep([class*="max-w"]) {
  opacity: 1 !important;
  transform: none !important;
}

/* Prevent white flashes during scroll and section interference */
:deep(section) {
  background-color: transparent;
  position: relative;
  z-index: 1;
  /* Prevent scroll chaining between sections - prevents interference */
  overscroll-behavior: contain;
  /* Sections already have overflow-hidden class - don't override */
  /* Ensure absolute backgrounds stay within section bounds */
  /* Don't use isolation or contain - they cause stacking context issues */
}

/* Smooth transitions for all sections - optimized for scroll performance */
/* All content is visible immediately - no scroll-triggered animations */
section {
  will-change: auto; /* Disabled to prevent GPU layer creation during scroll */
  backface-visibility: hidden;
  transform: none !important; /* No transforms to prevent animation effects */
  background-color: transparent;
  position: relative;
  opacity: 1 !important; /* Always visible */
  /* Prevent scroll chaining - sections won't interfere with each other */
  overscroll-behavior: contain;
  /* Sections already have overflow-hidden class - ensures absolute backgrounds are clipped */
  /* Ensure sections stay in document flow - no stacking context issues with nav */
  isolation: auto;
  /* Only allow safe transitions that don't affect scroll */
  transition-property: background-color, border-color, color, fill, stroke !important;
  transition-duration: 150ms !important;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1) !important;
}

/* Ensure all section content is visible immediately */
section > * {
  opacity: 1 !important;
  transform: none !important;
  /* Only allow safe transitions */
  transition-property: background-color, border-color, color, fill, stroke !important;
  transition-duration: 150ms !important;
}

/* Ensure absolute positioned backgrounds stay within section bounds */
/* These backgrounds should be clipped by section's overflow-hidden */
section > div[class*="absolute"] {
  position: absolute !important;
  /* Ensure they don't move during scroll - stay relative to section */
  transform: none !important;
  will-change: auto;
  /* Clip to section bounds - sections have overflow-hidden */
  pointer-events: none;
  /* Ensure they're positioned relative to section, not viewport */
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

/* Prevent any transform or opacity animations on headings and text during scroll */
/* Content should be visible immediately - no scroll-triggered animations */
section h1,
section h2,
section h3,
section h4,
section p {
  opacity: 1 !important;
  transform: none !important;
  /* Only allow safe transitions that don't affect scroll */
  transition-property: background-color, border-color, color, fill, stroke !important;
  transition-duration: 150ms !important;
  animation: none !important;
}

/* Ensure divs and containers are visible immediately */
section > div,
section > .container {
  opacity: 1 !important;
  transform: none !important;
}

/* Ensure continuous background coverage */
.min-h-screen {
  background: linear-gradient(to bottom, #020617 0%, #0c1222 20%, #1e1b4b 50%, #0c1222 80%, #020617 100%);
  min-height: 100vh;
  position: relative;
  background-size: 100% 100%;
  background-repeat: no-repeat;
  /* Don't use background-attachment: fixed - it causes "following" effect */
  /* Prevent scroll chaining and improve performance */
  overscroll-behavior-y: none;
}

/* Fixed navigation - ensure it stays above all content */
:deep(nav) {
  position: fixed !important;
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  z-index: 9999 !important;
  /* Ensure nav doesn't create stacking context issues */
  isolation: auto;
  /* Prevent nav from being affected by section transforms */
  transform: none !important;
  will-change: auto;
}

/* Mobile responsive improvements for code blocks */
code {
  word-break: break-all;
  overflow-wrap: break-word;
  /* Ensure code blocks are scrollable on mobile */
  display: block;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

/* Request entry mobile improvements */
.request-entry {
  /* Ensure proper spacing on mobile */
  min-height: 2.5rem;
}

@media (max-width: 640px) {
  .request-entry {
    min-height: auto;
  }
  
  /* Better code block scrolling on mobile */
  code {
    font-size: 0.7rem;
    line-height: 1.4;
  }
}

/* Smooth scroll is handled by Tailwind's scroll-smooth class in main.css */
/* No need to override here - it causes conflicts */

@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

@keyframes slideInRight {
  from {
    opacity: 0;
    transform: translateX(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.request-item {
  opacity: 0;
  animation: slideInRight 0.5s ease-out forwards;
}

.request-item-0 {
  animation-delay: 0s;
}

.request-item-1 {
  animation-delay: 0.1s;
}

.request-item-2 {
  animation-delay: 0.2s;
}

.request-item-3 {
  animation-delay: 0.3s;
}

/* Real-time request streaming animation */
.request-stream-container {
  position: relative;
}

.request-item-new {
  position: relative;
  overflow: hidden;
}

.request-item-new::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(34, 197, 94, 0.2), transparent);
  animation: requestFlow 2s infinite;
}

.request-time,
.request-method,
.request-path,
.request-status,
.request-latency {
  opacity: 0;
  animation: typeIn 0.3s ease-out forwards;
}

.request-time {
  animation-delay: 0.1s;
}

.request-method {
  animation-delay: 0.2s;
}

.request-path {
  animation-delay: 0.3s;
}

.request-status {
  animation-delay: 0.4s;
}

.request-latency {
  animation-delay: 0.5s;
}

.request-item-1 .request-time { animation-delay: 0.6s; }
.request-item-1 .request-method { animation-delay: 0.7s; }
.request-item-1 .request-path { animation-delay: 0.8s; }
.request-item-1 .request-status { animation-delay: 0.9s; }
.request-item-1 .request-latency { animation-delay: 1.0s; }

.request-item-2 .request-time { animation-delay: 1.1s; }
.request-item-2 .request-method { animation-delay: 1.2s; }
.request-item-2 .request-path { animation-delay: 1.3s; }
.request-item-2 .request-status { animation-delay: 1.4s; }
.request-item-2 .request-latency { animation-delay: 1.5s; }

.request-item-3 .request-time { animation-delay: 1.6s; }
.request-item-3 .request-method { animation-delay: 1.7s; }
.request-item-3 .request-path { animation-delay: 1.8s; }
.request-item-3 .request-status { animation-delay: 1.9s; }
.request-item-3 .request-latency { animation-delay: 2.0s; }

@keyframes requestFlow {
  0% {
    left: -100%;
  }
  100% {
    left: 100%;
  }
}

@keyframes typeIn {
  from {
    opacity: 0;
    transform: translateY(-5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Continuous new request animation */
.request-item-new {
  animation: newRequestPulse 3s infinite;
}

@keyframes newRequestPulse {
  0%, 100% {
    border-color: rgba(51, 65, 85, 0.3);
    box-shadow: 0 0 0 0 rgba(34, 197, 94, 0);
  }
  50% {
    border-color: rgba(34, 197, 94, 0.5);
    box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.1);
  }
}

/* Feature card flow design */
.feature-card {
  position: relative;
  transition: all 0.3s ease;
}

.feature-card::before {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: 0.75rem;
  padding: 2px;
  background: linear-gradient(135deg, transparent, transparent);
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  mask-composite: exclude;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.feature-card:hover::before {
  opacity: 1;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.5), rgba(139, 92, 246, 0.5), rgba(34, 197, 94, 0.5));
}

.feature-card:hover {
  transform: translateY(-4px);
}

/* Scroll reveal animation - COMPLETELY DISABLED */
/* Content is now visible immediately without any scroll-triggered animations */
.scroll-reveal,
.scroll-reveal.revealed {
  opacity: 1 !important;
  transform: none !important;
  transition: none !important;
  animation: none !important;
}

/* Real-time request animation */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.animate-slide-in {
  animation: slideIn 0.4s ease-out;
}

.request-entry {
  transition: all 0.2s ease;
}
</style>