<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-950">
    <!-- Sidebar Navigation - Hidden on mobile -->
    <div class="hidden md:flex fixed inset-y-0 left-0 w-64 bg-slate-950/95 backdrop-blur-xl border-r border-slate-800/50 overflow-y-auto z-40">
      <div class="p-6">
        <router-link to="/" class="flex items-center space-x-3 mb-8">
          <div class="w-8 h-8 bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-500 rounded-lg flex items-center justify-center">
            <span class="text-white font-bold text-sm">U</span>
          </div>
          <span class="text-lg font-bold text-white">UniRoute</span>
        </router-link>
        
        <nav class="space-y-1">
          <router-link
            v-for="section in navigation"
            :key="section.path"
            :to="`/docs${section.path}`"
            class="block px-3 py-2 rounded-lg text-sm font-medium transition-colors"
            :class="
              $route.path === `/docs${section.path}`
                ? 'bg-blue-500/20 text-blue-400'
                : 'text-slate-300 hover:bg-slate-800/50 hover:text-white'
            "
          >
            {{ section.title }}
          </router-link>
          
          <!-- Sub-sections -->
          <div v-if="currentSection?.children" class="ml-4 mt-2 space-y-1">
            <router-link
              v-for="child in currentSection.children"
              :key="child.path"
              :to="`/docs${child.path}`"
              class="block px-3 py-2 rounded-lg text-sm transition-colors"
              :class="
                $route.path === `/docs${child.path}`
                  ? 'bg-blue-500/20 text-blue-400 font-medium'
                  : 'text-slate-400 hover:bg-slate-800/50 hover:text-white'
              "
            >
              {{ child.title }}
            </router-link>
          </div>
        </nav>
      </div>
    </div>

    <!-- Main Content -->
    <div class="md:pl-64">
      <!-- Mobile Menu Button -->
      <button
        @click="mobileMenuOpen = !mobileMenuOpen"
        class="fixed top-4 left-4 z-50 md:hidden p-2 bg-slate-950/95 backdrop-blur-xl rounded-lg border border-slate-800/50 shadow-lg"
        aria-label="Toggle menu"
      >
        <svg class="w-6 h-6 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>

      <!-- Mobile Sidebar -->
      <div
        v-if="mobileMenuOpen"
        class="fixed inset-0 z-40 md:hidden"
      >
        <div class="fixed inset-0 bg-black/50" @click="mobileMenuOpen = false"></div>
        <div class="fixed inset-y-0 left-0 w-72 sm:w-80 bg-slate-950/95 backdrop-blur-xl border-r border-slate-800/50 overflow-y-auto">
          <div class="p-6">
            <div class="flex items-center justify-between mb-8">
              <router-link to="/" class="flex items-center space-x-3" @click="mobileMenuOpen = false">
                <div class="w-8 h-8 bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-500 rounded-lg flex items-center justify-center">
                  <span class="text-white font-bold text-sm">U</span>
                </div>
                <span class="text-lg font-bold text-white">UniRoute</span>
              </router-link>
              <button @click="mobileMenuOpen = false" class="p-2 text-slate-300">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <nav class="space-y-1">
              <router-link
                v-for="section in navigation"
                :key="section.path"
                :to="`/docs${section.path}`"
                @click="mobileMenuOpen = false"
                class="block px-3 py-2 rounded-lg text-sm font-medium transition-colors"
                :class="
                  $route.path === `/docs${section.path}`
                    ? 'bg-blue-500/20 text-blue-400'
                    : 'text-slate-300 hover:bg-slate-800/50 hover:text-white'
                "
              >
                {{ section.title }}
              </router-link>
              
              <!-- Sub-sections in mobile menu -->
              <div v-if="currentSection?.children" class="ml-4 mt-2 space-y-1">
                <router-link
                  v-for="child in currentSection.children"
                  :key="child.path"
                  :to="`/docs${child.path}`"
                  @click="mobileMenuOpen = false"
                  class="block px-3 py-2 rounded-lg text-sm transition-colors"
                  :class="
                    $route.path === `/docs${child.path}`
                      ? 'bg-blue-500/20 text-blue-400 font-medium'
                      : 'text-slate-400 hover:bg-slate-800/50 hover:text-white'
                  "
                >
                  {{ child.title }}
                </router-link>
              </div>
            </nav>
          </div>
        </div>
      </div>

      <!-- Content Area -->
      <div class="max-w-4xl mx-auto px-4 sm:px-6 py-6 sm:py-12 relative z-0">
        <!-- Decorative background elements -->
        <div class="absolute top-0 right-0 w-96 h-96 bg-blue-500/5 rounded-full blur-3xl pointer-events-none"></div>
        <div class="absolute bottom-0 left-0 w-96 h-96 bg-purple-500/5 rounded-full blur-3xl pointer-events-none"></div>
        <div v-if="loading" class="text-center py-8 sm:py-12">
          <div class="animate-spin rounded-full h-10 w-10 sm:h-12 sm:w-12 border-b-2 border-blue-500 mx-auto"></div>
          <p class="mt-4 text-slate-300 text-sm sm:text-base">Loading documentation...</p>
        </div>
        
        <div v-else-if="error" class="text-center py-8 sm:py-12">
          <p class="text-red-400 text-sm sm:text-base">{{ error }}</p>
        </div>
        
        <div v-else class="prose prose-invert max-w-none relative z-10" @click="onDocContentClick">
          <div v-html="renderedContent" ref="contentRef"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import type { MarkedOptions } from 'marked'
import { updateDocumentHead } from '@/utils/head'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const error = ref<string | null>(null)
const renderedContent = ref('')
const mobileMenuOpen = ref(false)
const contentRef = ref<HTMLElement | null>(null)

const navigation = [
  { path: '/introduction', title: 'Introduction' },
  { path: '/getting-started', title: 'Getting Started' },
  { path: '/installation', title: 'Installation' },
  { path: '/authentication', title: 'Authentication' },
  { path: '/tunnels', title: 'Tunnels', children: [
    { path: '/tunnels/opening', title: 'Opening a Tunnel' },
    { path: '/tunnels/dev-run', title: 'Dev & Run' },
    { path: '/tunnels/protocols', title: 'Protocols' },
    { path: '/tunnels/tcp-use-cases', title: 'TCP use cases' },
    { path: '/tunnels/tls-use-cases', title: 'TLS use cases' },
    { path: '/tunnels/udp-use-cases', title: 'UDP use cases' },
    { path: '/tunnels/custom-domains', title: 'Custom Domains' },
    { path: '/tunnels/reserved-subdomains', title: 'Reserved Subdomains' }
  ]},
  { path: '/api', title: 'API Reference' },
  { path: '/routing', title: 'Routing & Strategy' },
  { path: '/security', title: 'Security' },
  { path: '/deployment', title: 'Deployment' },
  { path: '/cli', title: 'CLI Reference' }
]

const currentSection = computed(() => {
  const currentPath = route.path.replace('/docs', '')
  return navigation.find(section => currentPath === section.path || currentPath.startsWith(section.path + '/'))
})

function getCurrentDocTitle(): string {
  const currentPath = route.path.replace('/docs', '') || '/introduction'
  const path = currentPath.startsWith('/') ? currentPath : '/' + currentPath
  for (const section of navigation) {
    if (section.path === path) return section.title
    const children = (section as { children?: { path: string; title: string }[] }).children
    if (children) {
      const child = children.find(c => c.path === path)
      if (child) return child.title
    }
  }
  return 'Documentation'
}

function setDocHead() {
  const title = getCurrentDocTitle()
  const description = `UniRoute documentation: ${title}. Installation, authentication, tunnels, API reference, and deployment.`
  updateDocumentHead(title, description, route.fullPath)
}

function onDocContentClick(e: MouseEvent) {
  const link = (e.target as HTMLElement).closest('a')
  if (!link || !link.href) return
  try {
    const url = new URL(link.href, window.location.origin)
    if (url.origin === window.location.origin && url.pathname.startsWith('/docs')) {
      e.preventDefault()
      router.push(url.pathname + url.search + url.hash)
    }
  } catch {
  }
}

marked.setOptions({
  breaks: true,
  gfm: true,
  headerIds: true,
  mangle: false
} as MarkedOptions)

const loadDocumentation = async () => {
  loading.value = true
  error.value = null
  
  try {
    let path = route.path.replace('/docs', '') || '/introduction'
    if (!path.startsWith('/')) {
      path = '/' + path
    }

    const docPath = `/docs${path}.md`
    const response = await fetch(docPath)

    if (!response.ok) {
      if (response.status === 404) {
        if (path !== '/introduction') {
          window.location.href = '/docs'
          return
        }
      }
      throw new Error(`Documentation not found: ${path}`)
    }
    
    const markdown = await response.text()
    const html = marked.parse(markdown)
    const sanitized = DOMPurify.sanitize(html)
    
    renderedContent.value = sanitized
  } catch (err: any) {
    error.value = err.message || 'Failed to load documentation'
    console.error('Error loading documentation:', err)
  } finally {
    loading.value = false
  }
}

const enhanceNextSteps = () => {
  setTimeout(() => {
    const prose = contentRef.value?.closest('.prose') || document.querySelector('.prose')
    if (!prose) return

    const headings = prose.querySelectorAll('h2, h3')
    headings.forEach((heading) => {
      if (heading.textContent?.includes('Next Steps')) {
        heading.classList.add('next-steps-heading')

        let nextSibling = heading.nextElementSibling
        while (nextSibling && nextSibling.tagName !== 'UL' && nextSibling.tagName !== 'OL') {
          nextSibling = nextSibling.nextElementSibling
        }

        if (nextSibling && (nextSibling.tagName === 'UL' || nextSibling.tagName === 'OL')) {
          nextSibling.classList.add('next-steps-list')

          const listItems = nextSibling.querySelectorAll('li')
          listItems.forEach((li) => {
            li.classList.add('next-steps-item')
          })
        }
      }
    })
  }, 100)
}

watch(() => route.path, () => {
  mobileMenuOpen.value = false
  loadDocumentation()
  setDocHead()
})

watch(() => renderedContent.value, () => {
  enhanceNextSteps()
})

onMounted(() => {
  loadDocumentation().then(() => {
    enhanceNextSteps()
    setDocHead()
  })
})
</script>

<style scoped>
/* Prose styles for documentation - matching landing page theme */
:deep(.prose) {
  @apply text-slate-100;
}

:deep(.prose h1) {
  @apply text-3xl sm:text-4xl font-bold mb-4 sm:mb-6 text-white;
}

:deep(.prose h2) {
  @apply text-2xl sm:text-3xl font-bold mt-8 sm:mt-12 mb-3 sm:mb-4 text-white border-b border-slate-800 pb-2;
}

:deep(.prose h3) {
  @apply text-xl sm:text-2xl font-semibold mt-6 sm:mt-8 mb-2 sm:mb-3 text-white;
}

:deep(.prose h4) {
  @apply text-lg sm:text-xl font-semibold mt-4 sm:mt-6 mb-2 text-white;
}

:deep(.prose p) {
  @apply mb-4 leading-6 sm:leading-7 text-slate-300 text-sm sm:text-base;
}

:deep(.prose code) {
  @apply bg-slate-800/50 text-blue-400 px-1.5 py-0.5 rounded text-xs sm:text-sm font-mono border border-slate-700;
}

:deep(.prose pre) {
  @apply bg-slate-900/80 backdrop-blur-sm rounded-lg p-3 sm:p-4 overflow-x-auto mb-4 border border-slate-800 text-xs sm:text-sm;
}

:deep(.prose pre code) {
  @apply bg-transparent text-slate-100 p-0 border-0;
}

:deep(.prose ul, .prose ol) {
  @apply mb-4 ml-4 sm:ml-6;
}

:deep(.prose li) {
  @apply mb-2 text-slate-300 text-sm sm:text-base;
}

:deep(.prose a) {
  @apply text-blue-400 hover:text-blue-300 underline;
}

:deep(.prose blockquote) {
  @apply border-l-4 border-blue-500 pl-3 sm:pl-4 italic text-slate-400 my-4 bg-slate-900/30 rounded-r text-sm sm:text-base;
}

:deep(.prose table) {
  @apply w-full border-collapse mb-4 overflow-x-auto block;
  display: block;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

:deep(.prose table thead,
.prose table tbody,
.prose table tr) {
  display: table;
  width: 100%;
  table-layout: fixed;
}

:deep(.prose th) {
  @apply bg-slate-800/50 px-2 sm:px-4 py-2 text-left font-semibold border border-slate-700 text-slate-200 text-xs sm:text-sm;
}

:deep(.prose td) {
  @apply px-2 sm:px-4 py-2 border border-slate-700 text-slate-300 text-xs sm:text-sm;
}

:deep(.prose strong) {
  @apply text-white font-semibold;
}

:deep(.prose hr) {
  @apply border-slate-800 my-8;
}

/* Enhanced Next Steps section styling */
:deep(.prose .next-steps-heading) {
  @apply relative mt-12 mb-6 pb-4;
  background: linear-gradient(to right, rgba(59, 130, 246, 0.1), rgba(139, 92, 246, 0.1));
  padding: 1rem 1.5rem;
  border-radius: 0.5rem;
  border: none;
  border-left: 4px solid;
  border-image: linear-gradient(to bottom, #3b82f6, #8b5cf6) 1;
  background-clip: padding-box;
}

:deep(.prose .next-steps-heading::after) {
  content: "";
  @apply absolute bottom-0 left-0 right-0 h-0.5;
  background: linear-gradient(to right, #3b82f6, #8b5cf6);
  border-radius: 0 0 0.5rem 0.5rem;
}

/* Next Steps list styling */
:deep(.prose .next-steps-list) {
  @apply bg-gradient-to-br from-slate-900/60 to-slate-800/40 backdrop-blur-sm rounded-xl p-6 border border-slate-700/50 mt-6;
  list-style: none;
  margin-left: 0;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
}

:deep(.prose .next-steps-item) {
  @apply flex items-start gap-3 mb-4 last:mb-0 p-4 rounded-lg bg-slate-800/30 hover:bg-slate-800/50 transition-all border border-slate-700/30 hover:border-blue-500/50;
  position: relative;
  overflow: hidden;
}

:deep(.prose .next-steps-item::before) {
  content: "→";
  @apply text-blue-400 font-bold text-xl flex-shrink-0 mt-0.5 mr-2;
  background: linear-gradient(to bottom, #3b82f6, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

:deep(.prose .next-steps-item::after) {
  content: "";
  @apply absolute inset-0 opacity-0 transition-opacity;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(139, 92, 246, 0.1));
  pointer-events: none;
}

:deep(.prose .next-steps-item:hover::after) {
  @apply opacity-100;
}

:deep(.prose .next-steps-item a) {
  @apply text-blue-400 hover:text-blue-300 font-medium no-underline hover:underline transition-colors relative z-10;
}

/* Enhanced code blocks */
:deep(.prose pre) {
  @apply bg-gradient-to-br from-slate-900/90 to-slate-800/90 backdrop-blur-sm;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -1px rgba(0, 0, 0, 0.2);
  position: relative;
  border: 1px solid rgba(59, 130, 246, 0.2);
}

:deep(.prose pre::before) {
  content: "";
  @apply absolute top-0 left-0 right-0 h-1;
  background: linear-gradient(to right, #3b82f6, #6366f1, #8b5cf6);
  border-radius: 0.5rem 0.5rem 0 0;
}

/* Enhanced blockquotes */
:deep(.prose blockquote) {
  @apply border-l-4;
  background: linear-gradient(to right, rgba(59, 130, 246, 0.1), rgba(139, 92, 246, 0.05));
  border-image: linear-gradient(to bottom, #3b82f6, #8b5cf6) 1;
}

/* Better link styling */
:deep(.prose a) {
  @apply text-blue-400 hover:text-blue-300 font-medium transition-colors;
  text-decoration: none;
  border-bottom: 1px solid rgba(59, 130, 246, 0.3);
  transition: all 0.2s ease;
}

:deep(.prose a:hover) {
  @apply text-blue-300;
  border-bottom-color: rgba(59, 130, 246, 0.6);
}

/* Enhanced list items */
:deep(.prose ul:not(.next-steps-list) li,
.prose ol:not(.next-steps-list) li) {
  @apply relative pl-6;
}

:deep(.prose ul:not(.next-steps-list) li::before) {
  content: "▸";
  @apply absolute left-0 text-blue-400 font-bold;
}

/* Regular lists with better spacing */
:deep(.prose ul:not(.next-steps-list),
.prose ol:not(.next-steps-list)) {
  @apply space-y-2;
}

/* Better table styling */
:deep(.prose table) {
  @apply rounded-lg overflow-hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(59, 130, 246, 0.2);
}

:deep(.prose th) {
  background: linear-gradient(to right, rgba(30, 41, 59, 0.9), rgba(51, 65, 85, 0.9));
  @apply text-slate-100;
}

:deep(.prose tr:hover td) {
  @apply bg-slate-800/30;
  transition: background-color 0.2s ease;
}

/* Enhanced headings with gradient accents */
:deep(.prose h1) {
  background: linear-gradient(to right, #ffffff, #e2e8f0);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  position: relative;
  padding-bottom: 0.5rem;
}

:deep(.prose h1::after) {
  content: "";
  @apply absolute bottom-0 left-0 w-24 h-1;
  background: linear-gradient(to right, #3b82f6, #8b5cf6);
  border-radius: 2px;
}

:deep(.prose h2) {
  @apply relative;
}

:deep(.prose h2:not(.next-steps-heading)::after) {
  content: "";
  @apply absolute bottom-0 left-0 w-16 h-0.5;
  background: linear-gradient(to right, #3b82f6, #8b5cf6);
  border-radius: 2px;
}

/* Improved code inline styling */
:deep(.prose code:not(pre code)) {
  @apply bg-slate-800/70 text-blue-300 px-2 py-1 rounded;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  font-weight: 500;
  border: 1px solid rgba(59, 130, 246, 0.2);
}

/* Add subtle animations */
:deep(.prose .next-steps-item) {
  animation: fadeInUp 0.3s ease-out;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Improve overall prose spacing */
:deep(.prose > * + *) {
  @apply mt-4;
}

:deep(.prose > h1 + *,
.prose > h2 + *,
.prose > h3 + *) {
  @apply mt-6;
}

/* Better paragraph spacing */
:deep(.prose p + p) {
  @apply mt-4;
}
</style>
