<template>
  <div :class="cardClasses">
    <div v-if="$slots.header || title" class="card-header">
      <slot name="header">
        <h3 v-if="title" class="text-lg font-semibold text-white">
          {{ title }}
        </h3>
      </slot>
    </div>
    <div class="card-body">
      <slot />
    </div>
    <div v-if="$slots.footer" class="card-footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  title?: string
  variant?: 'default' | 'elevated' | 'outlined'
  padding?: 'none' | 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  padding: 'md'
})

const cardClasses = computed(() => {
  const base = 'rounded-lg transition-smooth'
  
  const variants = {
    default: 'bg-slate-800/60 border border-slate-700/50',
    elevated: 'bg-slate-800/60 shadow-lg border border-slate-700/50',
    outlined: 'bg-transparent border-2 border-slate-700/50'
  }
  
  const paddings = {
    none: '',
    sm: 'p-3',
    md: 'p-6',
    lg: 'p-8'
  }
  
  return `${base} ${variants[props.variant]} ${paddings[props.padding]}`
})
</script>

<style scoped>
.card-header {
  @apply mb-4 pb-4 border-b border-slate-700/50;
}

.card-body {
  @apply space-y-4;
}

.card-footer {
  @apply mt-4 pt-4 border-t border-slate-700/50;
}
</style>

