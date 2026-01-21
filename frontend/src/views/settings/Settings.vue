<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-3xl font-bold text-white mb-2">Settings</h1>
      <p class="text-slate-400">Manage your account settings and preferences</p>
    </div>

    <!-- Navigation Tabs -->
    <div class="border-b border-slate-800">
      <nav class="flex space-x-8" aria-label="Settings navigation">
        <router-link
          to="/dashboard/settings/profile"
          class="py-4 px-1 border-b-2 font-medium text-sm transition-colors"
          :class="
            $route.name === 'settings-profile'
              ? 'border-blue-500 text-blue-400'
              : 'border-transparent text-slate-400 hover:text-slate-300 hover:border-slate-300'
          "
        >
          Profile
        </router-link>
        <router-link
          to="/dashboard/settings/provider-keys"
          class="py-4 px-1 border-b-2 font-medium text-sm transition-colors"
          :class="
            $route.name === 'settings-provider-keys'
              ? 'border-blue-500 text-blue-400'
              : 'border-transparent text-slate-400 hover:text-slate-300 hover:border-slate-300'
          "
        >
          Provider Keys
        </router-link>
        <router-link
          to="/dashboard/settings/routing-strategy"
          class="py-4 px-1 border-b-2 font-medium text-sm transition-colors"
          :class="
            $route.name === 'settings-routing-strategy'
              ? 'border-blue-500 text-blue-400'
              : 'border-transparent text-slate-400 hover:text-slate-300 hover:border-slate-300'
          "
        >
          Routing Strategy
        </router-link>
        <router-link
          to="/dashboard/settings/custom-rules"
          class="py-4 px-1 border-b-2 font-medium text-sm transition-colors"
          :class="
            $route.name === 'settings-custom-rules'
              ? 'border-blue-500 text-blue-400'
              : 'border-transparent text-slate-400 hover:text-slate-300 hover:border-slate-300'
          "
        >
          Custom Rules
        </router-link>
      </nav>
    </div>

    <!-- Settings Content -->
    <router-view />
  </div>
</template>

