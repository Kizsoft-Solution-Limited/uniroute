<template>
  <div class="input-wrapper">
    <label v-if="label" :for="inputId" class="block text-sm font-semibold text-slate-200 mb-2">
      {{ label }}
      <span v-if="required" class="text-red-400 ml-1">*</span>
    </label>
    <div class="relative">
      <div v-if="iconLeft" class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <component :is="iconLeft" class="h-5 w-5 text-slate-400" />
      </div>
      <input
        :id="inputId"
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :required="required"
        :class="inputClasses"
        @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
        @blur="$emit('blur', $event)"
        @focus="$emit('focus', $event)"
      />
      <div v-if="iconRight" class="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
        <component :is="iconRight" class="h-5 w-5 text-slate-400" />
      </div>
    </div>
    <p v-if="error" class="mt-1 text-sm text-red-400">{{ error }}</p>
    <p v-else-if="hint" class="mt-1 text-sm text-slate-400">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  modelValue: string
  label?: string
  type?: string
  placeholder?: string
  disabled?: boolean
  required?: boolean
  error?: string
  hint?: string
  iconLeft?: any
  iconRight?: any
  inputId?: string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false,
  required: false,
  inputId: () => `input-${Math.random().toString(36).substr(2, 9)}`
})

defineEmits<{
  'update:modelValue': [value: string]
  blur: [event: FocusEvent]
  focus: [event: FocusEvent]
}>()

const inputClasses = computed(() => {
  const base = 'block w-full rounded-lg border transition-smooth focus:outline-none focus:ring-2 focus:ring-offset-0 disabled:opacity-50 disabled:cursor-not-allowed'
  const padding = props.iconLeft ? 'pl-10' : props.iconRight ? 'pr-10' : 'px-3'
  const state = props.error
    ? 'border-red-500 text-red-100 bg-red-900/20 placeholder-red-400 focus:border-red-400 focus:ring-red-500'
    : 'border-slate-700 bg-slate-800/60 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-blue-500'
  
  return `${base} ${padding} py-2 ${state}`
})
</script>

<style scoped>
.input-wrapper {
  @apply w-full;
}
</style>

