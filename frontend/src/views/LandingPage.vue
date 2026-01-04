<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
    <!-- Navigation -->
    <nav class="container mx-auto px-4 py-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-2">
          <div class="w-10 h-10 bg-gradient-to-br from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
            <span class="text-white font-bold text-xl">U</span>
          </div>
          <span class="text-2xl font-bold text-gray-900 dark:text-white">UniRoute</span>
        </div>
        <div class="flex items-center space-x-4">
          <a href="#features" class="hidden md:block px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors">
            Features
          </a>
          <a href="#pricing" class="hidden md:block px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors">
            Pricing
          </a>
          <a href="#docs" class="hidden md:block px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors">
            Docs
          </a>
          <router-link
            to="/login"
            class="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors"
          >
            Sign in
          </router-link>
          <router-link
            to="/register"
            class="px-6 py-2.5 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg hover:from-blue-700 hover:to-purple-700 transition-all shadow-lg hover:shadow-xl transform hover:scale-105"
          >
            Get started
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <section class="container mx-auto px-4 py-20 md:py-32 text-center">
      <div class="max-w-4xl mx-auto">
        <div class="inline-block px-4 py-2 mb-6 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-full text-sm font-medium animate-slide-down">
          🚀 One unified gateway for every AI model
        </div>
        <h1 class="text-5xl md:text-7xl font-bold text-gray-900 dark:text-white mb-6 animate-fade-in leading-tight">
          Route, secure, and manage
          <span class="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600">
            traffic to any LLM
          </span>
        </h1>
        <p class="text-xl md:text-2xl text-gray-600 dark:text-gray-300 mb-10 max-w-3xl mx-auto animate-slide-up leading-relaxed">
          Cloud or local. Bring your own keys or use ours. 
          <span class="font-semibold text-gray-800 dark:text-gray-200">One platform for all your AI needs.</span>
        </p>
        <div class="flex flex-col sm:flex-row gap-4 justify-center items-center animate-slide-up">
          <router-link
            to="/register"
            class="px-8 py-4 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg text-lg font-semibold hover:from-blue-700 hover:to-purple-700 transition-all shadow-lg hover:shadow-2xl transform hover:scale-105"
          >
            Start free →
          </router-link>
          <a
            href="#demo"
            class="px-8 py-4 bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded-lg text-lg font-semibold border-2 border-gray-300 dark:border-gray-700 hover:border-blue-500 dark:hover:border-blue-500 transition-all shadow-md hover:shadow-lg"
          >
            Watch demo
          </a>
        </div>
        <p class="mt-6 text-sm text-gray-500 dark:text-gray-400">
          No credit card required • Free forever plan available
        </p>
      </div>

      <!-- Hero Visual -->
      <div class="mt-16 max-w-5xl mx-auto animate-fade-in">
        <div class="relative">
          <div class="absolute inset-0 bg-gradient-to-r from-blue-400 to-purple-400 rounded-2xl blur-3xl opacity-20"></div>
          <div class="relative glass rounded-2xl p-8 md:p-12 border border-white/20 dark:border-gray-700/20">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div class="bg-white/50 dark:bg-gray-800/50 rounded-lg p-6 backdrop-blur-sm">
                <div class="text-3xl mb-2">🤖</div>
                <h3 class="font-semibold mb-1">OpenAI</h3>
                <p class="text-sm text-gray-600 dark:text-gray-400">GPT-4, GPT-3.5</p>
              </div>
              <div class="bg-white/50 dark:bg-gray-800/50 rounded-lg p-6 backdrop-blur-sm">
                <div class="text-3xl mb-2">🧠</div>
                <h3 class="font-semibold mb-1">Anthropic</h3>
                <p class="text-sm text-gray-600 dark:text-gray-400">Claude 3, Claude 2</p>
              </div>
              <div class="bg-white/50 dark:bg-gray-800/50 rounded-lg p-6 backdrop-blur-sm">
                <div class="text-3xl mb-2">⚡</div>
                <h3 class="font-semibold mb-1">Local</h3>
                <p class="text-sm text-gray-600 dark:text-gray-400">Ollama, vLLM</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section id="features" class="container mx-auto px-4 py-20 md:py-32">
      <div class="text-center mb-16">
        <h2 class="text-4xl md:text-5xl font-bold text-gray-900 dark:text-white mb-4">
          Everything you need
        </h2>
        <p class="text-xl text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
          Built for developers, designed for scale
        </p>
      </div>
      <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
        <Card class="text-center hover:shadow-xl transition-all transform hover:-translate-y-2">
          <div class="text-5xl mb-4">🔐</div>
          <h3 class="text-2xl font-semibold mb-3 text-gray-900 dark:text-white">Enterprise Security</h3>
          <p class="text-gray-600 dark:text-gray-400">
            AES-256 encryption, rate limiting, API key management, and comprehensive access control. 
            Your data is protected at every layer.
          </p>
        </Card>
        <Card class="text-center hover:shadow-xl transition-all transform hover:-translate-y-2">
          <div class="text-5xl mb-4">⚡</div>
          <h3 class="text-2xl font-semibold mb-3 text-gray-900 dark:text-white">Lightning Fast</h3>
          <p class="text-gray-600 dark:text-gray-400">
            Low-latency routing with intelligent load balancing, failover, and latency-based selection. 
            Get the best performance automatically.
          </p>
        </Card>
        <Card class="text-center hover:shadow-xl transition-all transform hover:-translate-y-2">
          <div class="text-5xl mb-4">🔌</div>
          <h3 class="text-2xl font-semibold mb-3 text-gray-900 dark:text-white">Universal Support</h3>
          <p class="text-gray-600 dark:text-gray-400">
            OpenAI, Anthropic, Google, and local models like Ollama. 
            One API for all your AI needs.
          </p>
        </Card>
        <Card class="text-center hover:shadow-xl transition-all transform hover:-translate-y-2">
          <div class="text-5xl mb-4">💰</div>
          <h3 class="text-2xl font-semibold mb-3 text-gray-900 dark:text-white">BYOK Model</h3>
          <p class="text-gray-600 dark:text-gray-400">
            Bring your own API keys. You pay providers directly. 
            We handle the routing, security, and management.
          </p>
        </Card>
        <Card class="text-center hover:shadow-xl transition-all transform hover:-translate-y-2">
          <div class="text-5xl mb-4">📊</div>
          <h3 class="text-2xl font-semibold mb-3 text-gray-900 dark:text-white">Analytics & Insights</h3>
          <p class="text-gray-600 dark:text-gray-400">
            Track usage, costs, latency, and performance. 
            Make data-driven decisions with comprehensive analytics.
          </p>
        </Card>
        <Card class="text-center hover:shadow-xl transition-all transform hover:-translate-y-2">
          <div class="text-5xl mb-4">🌐</div>
          <h3 class="text-2xl font-semibold mb-3 text-gray-900 dark:text-white">Built-in Tunneling</h3>
          <p class="text-gray-600 dark:text-gray-400">
            Expose your local servers instantly. 
            No ngrok, no cloudflared. Built-in, secure, and free.
          </p>
        </Card>
      </div>
    </section>

    <!-- Social Proof Section -->
    <section class="container mx-auto px-4 py-20">
      <div class="text-center mb-12">
        <h2 class="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-4">
          Trusted by developers
        </h2>
        <p class="text-lg text-gray-600 dark:text-gray-300">
          Join thousands of developers building the next generation of AI applications
        </p>
      </div>
      <div class="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
        <Card class="glass">
          <div class="flex items-center mb-4">
            <div class="w-12 h-12 bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center mr-4">
              <span class="text-blue-600 dark:text-blue-400 font-bold">JD</span>
            </div>
            <div>
              <h4 class="font-semibold text-gray-900 dark:text-white">John Doe</h4>
              <p class="text-sm text-gray-500 dark:text-gray-400">CTO, TechCorp</p>
            </div>
          </div>
          <p class="text-gray-600 dark:text-gray-400">
            "UniRoute simplified our AI infrastructure. One API for all providers, 
            and we can use our own keys. Perfect!"
          </p>
        </Card>
        <Card class="glass">
          <div class="flex items-center mb-4">
            <div class="w-12 h-12 bg-purple-100 dark:bg-purple-900/30 rounded-full flex items-center justify-center mr-4">
              <span class="text-purple-600 dark:text-purple-400 font-bold">SM</span>
            </div>
            <div>
              <h4 class="font-semibold text-gray-900 dark:text-white">Sarah Miller</h4>
              <p class="text-sm text-gray-500 dark:text-gray-400">Lead Developer, StartupXYZ</p>
            </div>
          </div>
          <p class="text-gray-600 dark:text-gray-400">
            "The built-in tunneling is a game-changer. No more external tools, 
            everything just works out of the box."
          </p>
        </Card>
        <Card class="glass">
          <div class="flex items-center mb-4">
            <div class="w-12 h-12 bg-pink-100 dark:bg-pink-900/30 rounded-full flex items-center justify-center mr-4">
              <span class="text-pink-600 dark:text-pink-400 font-bold">AK</span>
            </div>
            <div>
              <h4 class="font-semibold text-gray-900 dark:text-white">Alex Kim</h4>
              <p class="text-sm text-gray-500 dark:text-gray-400">Founder, AI Labs</p>
            </div>
          </div>
          <p class="text-gray-600 dark:text-gray-400">
            "Best AI gateway we've used. The analytics and cost tracking help us 
            optimize our AI spending significantly."
          </p>
        </Card>
      </div>
    </section>

    <!-- Pricing Section -->
    <section id="pricing" class="container mx-auto px-4 py-20 md:py-32">
      <div class="text-center mb-16">
        <h2 class="text-4xl md:text-5xl font-bold text-gray-900 dark:text-white mb-4">
          Simple, transparent pricing
        </h2>
        <p class="text-xl text-gray-600 dark:text-gray-300">
          Free forever plan. Upgrade when you need more.
        </p>
      </div>
      <div class="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
        <!-- Free Plan -->
        <Card class="relative">
          <div class="text-center">
            <h3 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">Free</h3>
            <div class="mb-4">
              <span class="text-4xl font-bold text-gray-900 dark:text-white">$0</span>
              <span class="text-gray-600 dark:text-gray-400">/month</span>
            </div>
            <ul class="text-left space-y-3 mb-6">
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">10,000 requests/month</span>
              </li>
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">Basic analytics</span>
              </li>
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">Built-in tunneling</span>
              </li>
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">BYOK support</span>
              </li>
            </ul>
            <router-link
              to="/register"
              class="block w-full px-6 py-3 bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors text-center font-semibold"
            >
              Get started
            </router-link>
          </div>
        </Card>

        <!-- Pro Plan -->
        <Card class="relative border-2 border-blue-500 transform scale-105">
          <div class="absolute top-0 right-0 bg-blue-600 text-white px-4 py-1 rounded-bl-lg rounded-tr-lg text-sm font-semibold">
            Popular
          </div>
          <div class="text-center">
            <h3 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">Pro</h3>
            <div class="mb-4">
              <span class="text-4xl font-bold text-gray-900 dark:text-white">$29</span>
              <span class="text-gray-600 dark:text-gray-400">/month</span>
            </div>
            <ul class="text-left space-y-3 mb-6">
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">100,000 requests/month</span>
              </li>
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">Advanced analytics</span>
              </li>
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">Priority support</span>
              </li>
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">Custom domains</span>
              </li>
            </ul>
            <router-link
              to="/register"
              class="block w-full px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg hover:from-blue-700 hover:to-purple-700 transition-all text-center font-semibold shadow-lg"
            >
              Start Pro trial
            </router-link>
          </div>
        </Card>

        <!-- Enterprise Plan -->
        <Card class="relative">
          <div class="text-center">
            <h3 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">Enterprise</h3>
            <div class="mb-4">
              <span class="text-4xl font-bold text-gray-900 dark:text-white">Custom</span>
            </div>
            <ul class="text-left space-y-3 mb-6">
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">Unlimited requests</span>
              </li>
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">Dedicated support</span>
              </li>
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">SLA guarantee</span>
              </li>
              <li class="flex items-center">
                <svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
                <span class="text-gray-600 dark:text-gray-400">Self-hosted option</span>
              </li>
            </ul>
            <a
              href="mailto:sales@uniroute.dev"
              class="block w-full px-6 py-3 bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors text-center font-semibold"
            >
              Contact sales
            </a>
          </div>
        </Card>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="container mx-auto px-4 py-20 md:py-32">
      <div class="max-w-4xl mx-auto text-center glass rounded-3xl p-12 md:p-16 border border-white/20 dark:border-gray-700/20">
        <h2 class="text-4xl md:text-5xl font-bold text-gray-900 dark:text-white mb-6">
          Ready to get started?
        </h2>
        <p class="text-xl md:text-2xl text-gray-600 dark:text-gray-300 mb-10">
          Join thousands of developers building the next generation of AI applications
        </p>
        <div class="flex flex-col sm:flex-row gap-4 justify-center">
          <router-link
            to="/register"
            class="px-8 py-4 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg text-lg font-semibold hover:from-blue-700 hover:to-purple-700 transition-all shadow-lg hover:shadow-2xl transform hover:scale-105"
          >
            Create free account
          </router-link>
          <a
            href="#docs"
            class="px-8 py-4 bg-white/50 dark:bg-gray-800/50 text-gray-900 dark:text-white rounded-lg text-lg font-semibold border-2 border-gray-300 dark:border-gray-700 hover:border-blue-500 dark:hover:border-blue-500 transition-all"
          >
            View documentation
          </a>
        </div>
        <p class="mt-8 text-sm text-gray-500 dark:text-gray-400">
          No credit card required • Setup in 2 minutes • Free forever plan
        </p>
      </div>
    </section>

    <!-- Footer -->
    <footer class="container mx-auto px-4 py-12 border-t border-gray-200 dark:border-gray-800">
      <div class="grid md:grid-cols-4 gap-8 mb-8">
        <div>
          <div class="flex items-center space-x-2 mb-4">
            <div class="w-8 h-8 bg-gradient-to-br from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
              <span class="text-white font-bold">U</span>
            </div>
            <span class="text-xl font-bold text-gray-900 dark:text-white">UniRoute</span>
          </div>
          <p class="text-gray-600 dark:text-gray-400 text-sm">
            One unified gateway for every AI model
          </p>
        </div>
        <div>
          <h4 class="font-semibold text-gray-900 dark:text-white mb-4">Product</h4>
          <ul class="space-y-2 text-sm">
            <li><a href="#features" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">Features</a></li>
            <li><a href="#pricing" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">Pricing</a></li>
            <li><a href="#docs" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">Documentation</a></li>
          </ul>
        </div>
        <div>
          <h4 class="font-semibold text-gray-900 dark:text-white mb-4">Company</h4>
          <ul class="space-y-2 text-sm">
            <li><a href="/about" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">About</a></li>
            <li><a href="/blog" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">Blog</a></li>
            <li><a href="/careers" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">Careers</a></li>
          </ul>
        </div>
        <div>
          <h4 class="font-semibold text-gray-900 dark:text-white mb-4">Legal</h4>
          <ul class="space-y-2 text-sm">
            <li><a href="/privacy" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">Privacy</a></li>
            <li><a href="/terms" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">Terms</a></li>
            <li><a href="/security" class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">Security</a></li>
          </ul>
        </div>
      </div>
      <div class="pt-8 border-t border-gray-200 dark:border-gray-800 text-center text-sm text-gray-600 dark:text-gray-400">
        <p>&copy; 2024 UniRoute. All rights reserved.</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import Card from '@/components/ui/Card.vue'
</script>
