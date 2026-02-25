<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-950 overflow-x-hidden">
    <!-- Navigation -->
    <nav class="fixed top-0 left-0 right-0 z-50 bg-slate-950/95 backdrop-blur-xl border-b border-slate-800/50">
      <div class="container mx-auto px-4 sm:px-6 py-3 sm:py-4">
        <div class="flex items-center justify-between">
          <router-link to="/" class="flex items-center space-x-2 sm:space-x-3 min-w-0">
            <div class="w-9 h-9 sm:w-10 sm:h-10 flex-shrink-0 bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-500 rounded-lg flex items-center justify-center shadow-lg shadow-blue-500/20">
              <span class="text-white font-bold text-base sm:text-lg">U</span>
            </div>
            <span class="text-lg sm:text-xl font-bold text-white tracking-tight truncate">UniRoute</span>
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
          <!-- Mobile nav links -->
          <div class="flex md:hidden items-center gap-2">
            <router-link
              to="/login"
              class="px-3 py-1.5 text-sm font-medium text-slate-300 hover:text-white transition-colors"
            >
              Sign in
            </router-link>
            <router-link
              to="/register"
              class="px-3 py-1.5 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-sm font-semibold"
            >
              Get started
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <!-- Download Section -->
    <section class="pt-24 sm:pt-28 md:pt-32 pb-16 sm:pb-24 relative">
      <div class="container mx-auto px-4 sm:px-6">
        <div class="max-w-4xl mx-auto min-w-0">
          <!-- Header -->
          <div class="text-center mb-8 sm:mb-12">
            <h1 class="text-3xl sm:text-4xl md:text-5xl font-bold text-white mb-3 sm:mb-4 leading-tight tracking-tight px-1">
              Download UniRoute
            </h1>
            <p class="text-base sm:text-lg text-slate-300 max-w-2xl mx-auto mb-6 sm:mb-8">
              CLI for the terminal, or IDE extensions for VS Code and JetBrains.
            </p>
            <!-- Tabs -->
            <div class="inline-flex rounded-xl bg-slate-800/80 p-1 border border-slate-700/50">
              <button
                type="button"
                :class="[
                  'flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-semibold transition-all',
                  activeTab === 'cli'
                    ? 'bg-gradient-to-r from-blue-500 to-indigo-500 text-white shadow-lg shadow-blue-500/20'
                    : 'text-slate-400 hover:text-white'
                ]"
                @click="activeTab = 'cli'"
              >
                <Terminal class="w-4 h-4" />
                CLI
              </button>
              <button
                type="button"
                :class="[
                  'flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-semibold transition-all',
                  activeTab === 'extension'
                    ? 'bg-gradient-to-r from-blue-500 to-indigo-500 text-white shadow-lg shadow-blue-500/20'
                    : 'text-slate-400 hover:text-white'
                ]"
                @click="activeTab = 'extension'"
              >
                <Puzzle class="w-4 h-4" />
                Extension
              </button>
            </div>
          </div>

          <!-- CLI Tab -->
          <div v-show="activeTab === 'cli'" class="min-w-0">
          <!-- Detected Platform -->
          <div v-if="detectedPlatform" class="bg-slate-800/60 rounded-xl sm:rounded-2xl border border-slate-700/50 p-4 sm:p-6 md:p-8 mb-6 sm:mb-8">
            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4 sm:mb-6">
              <div class="min-w-0">
                <h2 class="text-xl sm:text-2xl font-semibold text-white mb-1 sm:mb-2">Your Platform</h2>
                <p class="text-sm sm:text-base text-slate-400">
                  {{ detectedPlatform.os }} {{ detectedPlatform.arch }}
                </p>
              </div>
              <div class="w-10 h-10 sm:w-12 sm:h-12 text-slate-400 flex-shrink-0">
                <component :is="detectedPlatform.icon" class="w-full h-full" />
              </div>
            </div>
            <a
              :href="downloadUrl"
              class="block w-full px-4 sm:px-8 py-3 sm:py-4 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-base sm:text-lg font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20 hover:shadow-xl text-center"
              @click="trackDownload"
            >
              Download for {{ detectedPlatform.os }} {{ detectedPlatform.arch }}
            </a>
          </div>

          <!-- Manual Selection -->
          <div class="bg-slate-800/60 rounded-xl sm:rounded-2xl border border-slate-700/50 p-4 sm:p-6 md:p-8 min-w-0">
            <h2 class="text-xl sm:text-2xl font-semibold text-white mb-4 sm:mb-6">Or choose your platform</h2>
            <div class="grid sm:grid-cols-2 gap-3 sm:gap-4 min-w-0">
              <!-- macOS -->
              <div class="bg-slate-900/60 rounded-xl p-4 sm:p-6 border border-slate-700/50 hover:border-blue-500/50 transition-all min-w-0">
                <div class="flex items-center justify-between mb-3 sm:mb-4">
                  <h3 class="text-base sm:text-lg font-semibold text-white">macOS</h3>
                  <Apple class="w-7 h-7 sm:w-8 sm:h-8 text-slate-400 flex-shrink-0" />
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
              <div class="bg-slate-900/60 rounded-xl p-4 sm:p-6 border border-slate-700/50 hover:border-green-500/50 transition-all min-w-0">
                <div class="flex items-center justify-between mb-3 sm:mb-4">
                  <h3 class="text-base sm:text-lg font-semibold text-white">Linux</h3>
                  <Server class="w-7 h-7 sm:w-8 sm:h-8 text-slate-400 flex-shrink-0" />
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
              <div class="bg-slate-900/60 rounded-xl p-4 sm:p-6 border border-slate-700/50 hover:border-cyan-500/50 transition-all min-w-0">
                <div class="flex items-center justify-between mb-3 sm:mb-4">
                  <h3 class="text-base sm:text-lg font-semibold text-white">Windows</h3>
                  <Laptop class="w-7 h-7 sm:w-8 sm:h-8 text-slate-400 flex-shrink-0" />
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
              <div class="bg-slate-900/60 rounded-xl p-4 sm:p-6 border border-slate-700/50 hover:border-purple-500/50 transition-all min-w-0">
                <div class="flex items-center justify-between mb-3 sm:mb-4 min-w-0">
                  <h3 class="text-base sm:text-lg font-semibold text-white truncate">Quick Install</h3>
                  <Zap class="w-7 h-7 sm:w-8 sm:h-8 text-slate-400 flex-shrink-0" />
                </div>
                <div class="space-y-3 min-w-0">
                  <div class="bg-slate-950/60 rounded-lg p-3 sm:p-4 font-mono text-xs sm:text-sm text-slate-300 min-w-0 overflow-x-auto overflow-y-hidden max-w-full scrollbar-thin" style="-webkit-overflow-scrolling: touch;">
                    <code class="whitespace-nowrap">curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/download-cli.sh | bash</code>
                  </div>
                  <p class="text-xs sm:text-sm text-slate-400">
                    Automatically detects your platform and installs the CLI
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Installation Instructions -->
          <div class="mt-8 sm:mt-12 bg-slate-800/60 rounded-xl sm:rounded-2xl border border-slate-700/50 p-4 sm:p-6 md:p-8">
            <h2 class="text-xl sm:text-2xl font-semibold text-white mb-4 sm:mb-6">Installation Instructions</h2>
            <div class="space-y-4 sm:space-y-6">
              <!-- macOS/Linux -->
              <div v-if="detectedPlatform && detectedPlatform.os !== 'Windows'" class="min-w-0">
                <h3 class="text-base sm:text-lg font-semibold text-white mb-2 sm:mb-3">macOS / Linux</h3>
                <p class="text-sm text-slate-400 mb-3">
                  On Mac and Linux, the downloaded file is not runnable until you make it executable with <code class="bg-slate-800/60 px-1 rounded">chmod +x</code>. Use the actual filename (e.g. <code class="bg-slate-800/60 px-1 rounded">uniroute-darwin-arm64</code>, <code class="bg-slate-800/60 px-1 rounded">uniroute-darwin-amd64</code>) or rename it to <code class="bg-slate-800/60 px-1 rounded">uniroute</code> first.
                </p>
                <div class="bg-slate-950/60 rounded-lg p-3 sm:p-4 font-mono text-xs sm:text-sm text-slate-300 space-y-2 overflow-x-auto max-w-full" style="-webkit-overflow-scrolling: touch;">
                  <div>
                    <span class="text-slate-500"># If the file is in your Downloads folder:</span><br>
                    <span class="text-green-400">cd ~/Downloads</span>
                  </div>
                  <div>
                    <span class="text-slate-500"># Make executable (required after download)</span><br>
                    <span class="text-green-400">chmod +x uniroute-darwin-amd64</span>
                    <span class="text-slate-500 text-xs block mt-0.5">(or uniroute-darwin-arm64 / uniroute-linux-amd64)</span>
                  </div>
                  <div>
                    <span class="text-slate-500"># Move to PATH (optional)</span><br>
                    <span class="text-green-400">sudo mv uniroute-darwin-amd64 /usr/local/bin/uniroute</span>
                  </div>
                  <div>
                    <span class="text-slate-500"># Verify installation</span><br>
                    <span class="text-green-400">uniroute --version</span>
                  </div>
                </div>
              </div>

              <!-- Windows -->
              <div v-if="detectedPlatform && detectedPlatform.os === 'Windows'" class="min-w-0">
                <h3 class="text-base sm:text-lg font-semibold text-white mb-2 sm:mb-3">Windows</h3>
                <div class="bg-slate-950/60 rounded-lg p-3 sm:p-4 font-mono text-xs sm:text-sm text-slate-300 space-y-2 overflow-x-auto max-w-full" style="-webkit-overflow-scrolling: touch;">
                  <div>
                    <span class="text-slate-500"># Add to PATH or use directly</span><br>
                    <span class="text-green-400">.\uniroute.exe --version</span>
                  </div>
                </div>
              </div>

              <!-- Next Steps -->
              <div class="pt-6 border-t border-slate-700/50 min-w-0">
                <h3 class="text-lg font-semibold text-white mb-3">Next Steps</h3>
                <ol class="list-decimal list-inside space-y-2 text-slate-300">
                  <li class="break-words">Login to your account: <code class="bg-slate-900/60 px-2 py-1 rounded text-sm break-all">uniroute auth login</code></li>
                  <li class="break-words">Create an API key: <code class="bg-slate-900/60 px-2 py-1 rounded text-sm break-all">uniroute keys create</code></li>
                  <li>Start using UniRoute!</li>
                </ol>
              </div>
            </div>
          </div>
          </div>

          <!-- Extension Tab -->
          <div v-show="activeTab === 'extension'" class="min-w-0 space-y-6">
            <p class="text-slate-300 text-center">
              Get UniRoute in your IDE: chat, accept/reject AI edits, and tunnels. Download the latest release and install from disk.
            </p>
            <div class="grid sm:grid-cols-2 gap-4 sm:gap-6">
              <a
                :href="releasesUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="block bg-slate-800/60 rounded-xl sm:rounded-2xl border border-slate-700/50 p-6 sm:p-8 hover:border-blue-500/50 transition-all group"
              >
                <div class="flex items-center gap-3 mb-4">
                  <div class="w-12 h-12 rounded-xl bg-slate-700/60 flex items-center justify-center group-hover:bg-blue-500/20 transition-colors">
                    <Code2 class="w-6 h-6 text-slate-300 group-hover:text-blue-400" />
                  </div>
                  <h2 class="text-xl font-semibold text-white">VS Code</h2>
                </div>
                <p class="text-sm text-slate-400 mb-4">
                  Chat in the sidebar, accept or reject AI code edits, start tunnels. Install the <code class="bg-slate-900/60 px-1.5 py-0.5 rounded text-slate-300">.vsix</code> from the release.
                </p>
                <span class="inline-flex items-center text-blue-400 text-sm font-medium">
                  Download from GitHub →
                </span>
              </a>
              <a
                :href="releasesUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="block bg-slate-800/60 rounded-xl sm:rounded-2xl border border-slate-700/50 p-6 sm:p-8 hover:border-indigo-500/50 transition-all group"
              >
                <div class="flex items-center gap-3 mb-4">
                  <div class="w-12 h-12 rounded-xl bg-slate-700/60 flex items-center justify-center group-hover:bg-indigo-500/20 transition-colors">
                    <Puzzle class="w-6 h-6 text-slate-300 group-hover:text-indigo-400" />
                  </div>
                  <h2 class="text-xl font-semibold text-white">IntelliJ / Android Studio</h2>
                </div>
                <p class="text-sm text-slate-400 mb-4">
                  Same plugin for IntelliJ IDEA and Android Studio. Download the <code class="bg-slate-900/60 px-1.5 py-0.5 rounded text-slate-300">.zip</code> and use Plugins → Install from Disk.
                </p>
                <span class="inline-flex items-center text-indigo-400 text-sm font-medium">
                  Download from GitHub →
                </span>
              </a>
            </div>
            <div class="bg-slate-800/60 rounded-xl border border-slate-700/50 p-4 sm:p-6">
              <h3 class="text-base font-semibold text-white mb-2">How to install</h3>
              <ul class="text-sm text-slate-400 space-y-1">
                <li><strong class="text-slate-300">VS Code:</strong> Extensions → ⋯ → Install from VSIX… → select the <code class="bg-slate-900/60 px-1 rounded">.vsix</code> file.</li>
                <li><strong class="text-slate-300">IntelliJ / Android Studio:</strong> Settings → Plugins → ⚙ → Install Plugin from Disk… → select the <code class="bg-slate-900/60 px-1 rounded">.zip</code> file.</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Monitor, Apple, Server, Laptop, Zap, Code2, Puzzle, Terminal } from 'lucide-vue-next'

const activeTab = ref<'cli' | 'extension'>('cli')
const releasesUrl = 'https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest'

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

  if (platform.includes('win') || userAgent.includes('windows')) {
    os = 'Windows'
    icon = Laptop
    arch = 'amd64'
    binary = 'uniroute-windows-amd64.exe'
  } else if (platform.includes('mac') || userAgent.includes('mac')) {
    os = 'macOS'
    icon = Apple
    if (navigator.userAgent.includes('Intel') || platform.includes('x86')) {
      arch = 'amd64'
      binary = 'uniroute-darwin-amd64'
    } else {
      arch = 'arm64'
      binary = 'uniroute-darwin-arm64'
    }
  } else if (platform.includes('linux') || userAgent.includes('linux')) {
    os = 'Linux'
    icon = Server
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

