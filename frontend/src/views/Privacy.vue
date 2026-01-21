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
          <h1 class="text-4xl md:text-5xl font-bold text-white mb-6">Privacy Policy</h1>
          <div class="prose prose-invert max-w-none space-y-6 text-slate-300">
            <p class="text-lg text-slate-200">
              Your privacy is important to us. This Privacy Policy explains how UniRoute collects, uses, and protects your information.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Information We Collect</h2>
            <p>
              We collect information that you provide directly to us, including:
            </p>
            <ul class="list-disc list-inside space-y-2 ml-4">
              <li>Account information (email address, password)</li>
              <li>API keys and provider API keys (encrypted)</li>
              <li>Usage data and analytics</li>
              <li>Request logs and performance metrics</li>
            </ul>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">How We Use Your Information</h2>
            <p>
              We use the information we collect to:
            </p>
            <ul class="list-disc list-inside space-y-2 ml-4">
              <li>Provide and maintain the UniRoute service</li>
              <li>Process and route your API requests</li>
              <li>Monitor service performance and usage</li>
              <li>Send you service-related notifications</li>
              <li>Improve our services and develop new features</li>
            </ul>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Data Storage and Security</h2>
            <p>
              We implement industry-standard security measures to protect your data:
            </p>
            <ul class="list-disc list-inside space-y-2 ml-4">
              <li>All API keys are encrypted using AES-256-GCM encryption</li>
              <li>Passwords are hashed using bcrypt before storage</li>
              <li>All communications are encrypted using TLS/SSL</li>
              <li>Regular security audits and updates</li>
            </ul>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Data Sharing</h2>
            <p>
              We do not sell, trade, or rent your personal information to third parties. We may share aggregated, anonymized data for analytics purposes. We may disclose your information if required by law or to protect our rights and safety.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Third-Party Services</h2>
            <p>
              UniRoute integrates with third-party AI providers (OpenAI, Anthropic, Google) when you use their services. We only share the minimum information necessary to process your requests. Please review each provider's privacy policy for how they handle your data.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Your Rights</h2>
            <p>
              You have the right to:
            </p>
            <ul class="list-disc list-inside space-y-2 ml-4">
              <li>Access your personal data</li>
              <li>Correct inaccurate data</li>
              <li>Delete your account and data</li>
              <li>Export your data</li>
              <li>Opt out of certain data collection</li>
            </ul>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Cookies and Tracking</h2>
            <p>
              We use cookies and similar technologies to maintain your session, remember your preferences, and analyze usage. You can control cookies through your browser settings, though this may affect service functionality.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Data Retention</h2>
            <p>
              We retain your data for as long as your account is active or as needed to provide services. When you delete your account, we will delete your personal data within 30 days, except where we are required to retain it for legal purposes.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Children's Privacy</h2>
            <p>
              UniRoute is not intended for users under the age of 13. We do not knowingly collect personal information from children under 13. If you believe we have collected information from a child under 13, please contact us immediately.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Changes to This Policy</h2>
            <p>
              We may update this Privacy Policy from time to time. We will notify you of significant changes by email or through the service. Your continued use of UniRoute after changes constitutes acceptance of the updated policy.
            </p>

            <h2 class="text-2xl font-semibold text-white mt-8 mb-4">Contact Us</h2>
            <p>
              If you have questions about this Privacy Policy or our data practices, please contact us at <a href="mailto:privacy@uniroute.co" class="text-blue-400 hover:text-blue-300">privacy@uniroute.co</a>.
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

