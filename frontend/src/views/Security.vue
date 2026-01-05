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
          <router-link
            to="/"
            class="px-4 py-2 bg-slate-800/60 text-white rounded-lg text-sm font-semibold hover:bg-slate-700/60 transition-all"
          >
            Back to Home
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Content -->
    <section class="pt-32 pb-24">
      <div class="container mx-auto px-6 max-w-4xl">
        <div class="bg-slate-800/60 rounded-2xl border border-slate-700/50 p-8 md:p-12">
          <h1 class="text-4xl md:text-5xl font-bold text-white mb-6">Security</h1>
          <div class="prose prose-invert max-w-none space-y-6 text-slate-300">
            <p class="text-lg text-slate-200">
              At UniRoute, security is our top priority. We implement industry-standard security measures to protect your data and API keys.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Data Encryption</h2>
            <p>
              All API keys and sensitive data are encrypted at rest using AES-256-GCM encryption. Your provider API keys (OpenAI, Anthropic, Google) are encrypted before being stored in our database.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Authentication</h2>
            <p>
              We use JWT (JSON Web Tokens) for secure authentication. All API requests require authentication via API keys or JWT tokens. Passwords are hashed using bcrypt before storage.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Rate Limiting</h2>
            <p>
              To prevent abuse and ensure fair usage, we implement rate limiting on all API endpoints. Rate limits are configurable per API key and can be set per minute and per day.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Network Security</h2>
            <p>
              All communications are encrypted using TLS/SSL. We enforce HTTPS for all API endpoints and web interfaces. Our infrastructure is regularly updated with security patches.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Access Control</h2>
            <p>
              We implement role-based access control (RBAC) to ensure users only have access to resources they're authorized to use. API keys are scoped to specific users and projects.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Security Headers</h2>
            <p>
              We implement security headers including Content Security Policy (CSP), X-Frame-Options, X-Content-Type-Options, and Strict-Transport-Security to protect against common web vulnerabilities.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Monitoring & Incident Response</h2>
            <p>
              We continuously monitor our systems for security threats and anomalies. In the event of a security incident, we will notify affected users promptly and take immediate action to mitigate any risks.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Reporting Security Issues</h2>
            <p>
              If you discover a security vulnerability, please report it to us immediately at <a href="mailto:security@uniroute.dev" class="text-blue-400 hover:text-blue-300">security@uniroute.dev</a>. We take security issues seriously and will respond promptly.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Compliance</h2>
            <p>
              We are committed to maintaining high security standards and are continuously working to improve our security practices. We follow industry best practices and security frameworks.
            </p>

            <p class="text-sm text-slate-400 mt-12">
              Last updated: {{ new Date().toLocaleDateString() }}
            </p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
</script>

