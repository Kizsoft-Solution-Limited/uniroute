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
          <h1 class="text-4xl md:text-5xl font-bold text-white mb-6">Terms of Service</h1>
          <div class="prose prose-invert max-w-none space-y-6 text-slate-300">
            <p class="text-lg text-slate-200">
              By using UniRoute, you agree to the following terms and conditions.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">1. Acceptance of Terms</h2>
            <p>
              By accessing and using UniRoute, you accept and agree to be bound by the terms and provision of this agreement. If you do not agree to these terms, you should not use UniRoute.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">2. Use License</h2>
            <p>
              UniRoute is provided free of charge. You are granted a non-exclusive, non-transferable license to use UniRoute for your personal or commercial projects. This license does not include the right to resell, redistribute, or create derivative works for commercial purposes without explicit permission.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">3. User Responsibilities</h2>
            <p>
              You are responsible for maintaining the confidentiality of your account credentials and API keys. You agree to use UniRoute in compliance with all applicable laws and regulations. You will not use UniRoute for any illegal or unauthorized purpose.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">4. API Keys and Provider Keys</h2>
            <p>
              You are responsible for the security of your API keys and provider API keys. UniRoute stores your provider keys securely, but you are responsible for keeping your UniRoute API keys confidential. You agree not to share your API keys with unauthorized parties.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">5. Service Availability</h2>
            <p>
              While we strive to maintain high availability, UniRoute is provided "as is" without warranties of any kind. We do not guarantee uninterrupted or error-free service. We reserve the right to modify, suspend, or discontinue any part of the service at any time.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">6. Limitation of Liability</h2>
            <p>
              UniRoute shall not be liable for any indirect, incidental, special, consequential, or punitive damages resulting from your use or inability to use the service. Our total liability shall not exceed the amount you paid for the service (which is currently $0 as the service is free).
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">7. Intellectual Property</h2>
            <p>
              UniRoute is open source software licensed under the MIT License. You retain all rights to your data and content. By using UniRoute, you grant us a license to use your data solely for the purpose of providing the service.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">8. Termination</h2>
            <p>
              We reserve the right to terminate or suspend your account at any time, with or without cause, with or without notice. You may terminate your account at any time by contacting us or deleting your account through the dashboard.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">9. Changes to Terms</h2>
            <p>
              We reserve the right to modify these terms at any time. We will notify users of significant changes via email or through the service. Continued use of UniRoute after changes constitutes acceptance of the new terms.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">10. Contact</h2>
            <p>
              If you have any questions about these Terms of Service, please contact us at <a href="mailto:legal@uniroute.dev" class="text-blue-400 hover:text-blue-300">legal@uniroute.dev</a>.
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

