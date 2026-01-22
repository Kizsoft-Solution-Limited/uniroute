<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-950">
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
          <div class="hidden md:flex items-center space-x-8">
            <router-link to="/" class="text-sm font-medium text-slate-300 hover:text-white transition-colors">
              Home
            </router-link>
            <router-link to="/docs" class="text-sm font-medium text-slate-300 hover:text-white transition-colors">
              Docs
            </router-link>
            <router-link to="/register" class="px-4 py-2 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-sm font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20">
              Get started
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <!-- Pricing Section -->
    <section class="pt-24 pb-12 md:pb-16 relative">
      <div class="container mx-auto px-6">
        <div class="text-center mb-12 md:mb-16">
          <h1 class="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-4 tracking-tight">
            Simple, transparent pricing
          </h1>
          <p class="text-xl text-slate-400 max-w-2xl mx-auto">
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
            <router-link
              to="/register"
              class="block w-full text-center px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-sm font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20"
            >
              Get started
            </router-link>
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
            <router-link
              to="/register"
              class="block w-full text-center px-6 py-3 bg-gradient-to-r from-green-500 to-emerald-500 text-white rounded-lg text-sm font-semibold hover:from-green-600 hover:to-emerald-600 transition-all shadow-lg shadow-green-500/20"
            >
              Get started
            </router-link>
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
            <router-link
              to="/docs"
              class="block w-full text-center px-6 py-3 bg-gradient-to-r from-purple-500 to-pink-500 text-white rounded-lg text-sm font-semibold hover:from-purple-600 hover:to-pink-600 transition-all shadow-lg shadow-purple-500/20"
            >
              View Docs
            </router-link>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
// Pricing page component
</script>
