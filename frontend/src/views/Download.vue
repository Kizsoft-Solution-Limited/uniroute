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
        </div>
      </div>
    </nav>

    <!-- Download Section -->
    <section class="pt-32 pb-24 relative">
      <div class="container mx-auto px-6">
        <div class="max-w-4xl mx-auto">
          <!-- Header -->
          <div class="text-center mb-16">
            <h1 class="text-5xl md:text-6xl font-bold text-white mb-6 leading-tight tracking-tight">
              Download UniRoute CLI
            </h1>
            <p class="text-xl text-slate-300 max-w-2xl mx-auto">
              Get the CLI tool for your platform.
            </p>
          </div>

          <!-- Detected Platform -->
          <div v-if="detectedPlatform" class="bg-slate-800/60 rounded-2xl border border-slate-700/50 p-8 mb-8">
            <div class="flex items-center justify-between mb-6">
              <div>
                <h2 class="text-2xl font-semibold text-white mb-2">Your Platform</h2>
                <p class="text-slate-400">
                  {{ detectedPlatform.os }} {{ detectedPlatform.arch }}
                </p>
              </div>
              <div class="w-12 h-12 text-slate-400">
                <component :is="detectedPlatform.icon" class="w-full h-full" />
              </div>
            </div>
            <a
              :href="downloadUrl"
              class="block w-full px-8 py-4 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-lg font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20 hover:shadow-xl text-center"
              @click="trackDownload"
            >
              Download for {{ detectedPlatform.os }} {{ detectedPlatform.arch }}
            </a>
          </div>

          <!-- Manual Selection -->
          <div class="bg-slate-800/60 rounded-2xl border border-slate-700/50 p-8">
            <h2 class="text-2xl font-semibold text-white mb-6">Or choose your platform</h2>
            <div class="grid md:grid-cols-2 gap-4">
              <!-- macOS -->
              <div class="bg-slate-900/60 rounded-xl p-6 border border-slate-700/50 hover:border-blue-500/50 transition-all">
                <div class="flex items-center justify-between mb-4">
                  <h3 class="text-lg font-semibold text-white">macOS</h3>
                  <Apple class="w-8 h-8 text-slate-400" />
                </div>
                <div class="space-y-3">
                  <a
                    href="https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-arm64"
                    class="block w-full px-4 py-2 bg-slate-800/60 text-white rounded-lg text-sm font-medium hover:bg-slate-700/60 transition-all text-center"
                    @click="trackDownload('darwin-arm64')"
                  >
                    Apple Silicon (ARM64)
                  </a>
                  <a
                    href="https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-amd64"
                    class="block w-full px-4 py-2 bg-slate-800/60 text-white rounded-lg text-sm font-medium hover:bg-slate-700/60 transition-all text-center"
                    @click="trackDownload('darwin-amd64')"
                  >
                    Intel (AMD64)
                  </a>
                </div>
              </div>

              <!-- Linux -->
              <div class="bg-slate-900/60 rounded-xl p-6 border border-slate-700/50 hover:border-green-500/50 transition-all">
                <div class="flex items-center justify-between mb-4">
                  <h3 class="text-lg font-semibold text-white">Linux</h3>
                  <Server class="w-8 h-8 text-slate-400" />
                </div>
                <div class="space-y-3">
                  <a
                    href="https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-linux-amd64"
                    class="block w-full px-4 py-2 bg-slate-800/60 text-white rounded-lg text-sm font-medium hover:bg-slate-700/60 transition-all text-center"
                    @click="trackDownload('linux-amd64')"
                  >
                    AMD64
                  </a>
                  <a
                    href="https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-linux-arm64"
                    class="block w-full px-4 py-2 bg-slate-800/60 text-white rounded-lg text-sm font-medium hover:bg-slate-700/60 transition-all text-center"
                    @click="trackDownload('linux-arm64')"
                  >
                    ARM64
                  </a>
                </div>
              </div>

              <!-- Windows -->
              <div class="bg-slate-900/60 rounded-xl p-6 border border-slate-700/50 hover:border-cyan-500/50 transition-all">
                <div class="flex items-center justify-between mb-4">
                  <h3 class="text-lg font-semibold text-white">Windows</h3>
                  <Laptop class="w-8 h-8 text-slate-400" />
                </div>
                <div class="space-y-3">
                  <a
                    href="https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-windows-amd64.exe"
                    class="block w-full px-4 py-2 bg-slate-800/60 text-white rounded-lg text-sm font-medium hover:bg-slate-700/60 transition-all text-center"
                    @click="trackDownload('windows-amd64')"
                  >
                    AMD64 (.exe)
                  </a>
                </div>
              </div>

              <!-- Quick Install Script -->
              <div class="bg-slate-900/60 rounded-xl p-6 border border-slate-700/50 hover:border-purple-500/50 transition-all">
                <div class="flex items-center justify-between mb-4">
                  <h3 class="text-lg font-semibold text-white">Quick Install</h3>
                  <Zap class="w-8 h-8 text-slate-400" />
                </div>
                <div class="space-y-3">
                  <div class="bg-slate-950/60 rounded-lg p-4 font-mono text-xs text-slate-300 overflow-x-auto">
                    <code>curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/download-cli.sh | bash</code>
                  </div>
                  <p class="text-sm text-slate-400">
                    Automatically detects your platform and installs the CLI
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Installation Instructions -->
          <div class="mt-12 bg-slate-800/60 rounded-2xl border border-slate-700/50 p-8">
            <h2 class="text-2xl font-semibold text-white mb-6">Installation Instructions</h2>
            <div class="space-y-6">
              <!-- macOS/Linux -->
              <div v-if="detectedPlatform && detectedPlatform.os !== 'Windows'">
                <h3 class="text-lg font-semibold text-white mb-3">macOS / Linux</h3>
                <div class="bg-slate-950/60 rounded-lg p-4 font-mono text-sm text-slate-300 space-y-2">
                  <div>
                    <span class="text-slate-500"># Make executable</span><br>
                    <span class="text-green-400">chmod +x uniroute</span>
                  </div>
                  <div>
                    <span class="text-slate-500"># Move to PATH (optional)</span><br>
                    <span class="text-green-400">sudo mv uniroute /usr/local/bin/</span>
                  </div>
                  <div>
                    <span class="text-slate-500"># Verify installation</span><br>
                    <span class="text-green-400">uniroute --version</span>
                  </div>
                </div>
              </div>

              <!-- Windows -->
              <div v-if="detectedPlatform && detectedPlatform.os === 'Windows'">
                <h3 class="text-lg font-semibold text-white mb-3">Windows</h3>
                <div class="bg-slate-950/60 rounded-lg p-4 font-mono text-sm text-slate-300 space-y-2">
                  <div>
                    <span class="text-slate-500"># Add to PATH or use directly</span><br>
                    <span class="text-green-400">.\uniroute.exe --version</span>
                  </div>
                </div>
              </div>

              <!-- Next Steps -->
              <div class="pt-6 border-t border-slate-700/50">
                <h3 class="text-lg font-semibold text-white mb-3">Next Steps</h3>
                <ol class="list-decimal list-inside space-y-2 text-slate-300">
                  <li>Login to your account: <code class="bg-slate-900/60 px-2 py-1 rounded text-sm">uniroute auth login</code></li>
                  <li>Create an API key: <code class="bg-slate-900/60 px-2 py-1 rounded text-sm">uniroute keys create</code></li>
                  <li>Start using UniRoute!</li>
                </ol>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Monitor, Apple, Server, Laptop, Zap } from 'lucide-vue-next'

interface Platform {
  os: string
  arch: string
  icon: any
  binary: string
}

const detectedPlatform = ref<Platform | null>(null)

const downloadUrl = computed(() => {
  if (!detectedPlatform.value) return ''
  return `https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/${detectedPlatform.value.binary}`
})

function detectPlatform() {
  const userAgent = navigator.userAgent.toLowerCase()
  const platform = navigator.platform.toLowerCase()
  
  let os = 'Unknown'
  let arch = 'amd64'
  let icon = Monitor
  let binary = ''

  // Detect OS
  if (platform.includes('win') || userAgent.includes('windows')) {
    os = 'Windows'
    icon = Laptop
    arch = 'amd64'
    binary = 'uniroute-windows-amd64.exe'
  } else if (platform.includes('mac') || userAgent.includes('mac')) {
    os = 'macOS'
    icon = Apple
    // Detect architecture for macOS
    if (navigator.userAgent.includes('Intel') || platform.includes('x86')) {
      arch = 'amd64'
      binary = 'uniroute-darwin-amd64'
    } else {
      // Assume Apple Silicon for modern Macs
      arch = 'arm64'
      binary = 'uniroute-darwin-arm64'
    }
  } else if (platform.includes('linux') || userAgent.includes('linux')) {
    os = 'Linux'
    icon = Server
    // Default to amd64 for Linux (can't reliably detect arch in browser)
    arch = 'amd64'
    binary = 'uniroute-linux-amd64'
  }

  detectedPlatform.value = {
    os,
    arch,
    icon,
    binary
  }
}

function trackDownload(platform?: string) {
  const platformName = platform || detectedPlatform.value?.binary || 'unknown'
  // Track download analytics
  if (typeof window !== 'undefined' && (window as any).gtag) {
    ;(window as any).gtag('event', 'download', {
      platform: platformName
    })
  }
}

onMounted(() => {
  detectPlatform()
})
</script>

