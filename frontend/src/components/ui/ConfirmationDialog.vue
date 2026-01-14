<template>
  <Transition
    enter-active-class="transition ease-out duration-200"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition ease-in duration-150"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="show"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      @click.self="$emit('cancel')"
    >
      <Transition
        enter-active-class="transition ease-out duration-200"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition ease-in duration-150"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <Card v-if="show" class="w-full max-w-md">
          <div class="space-y-4">
            <!-- Icon and Title -->
            <div class="flex items-start space-x-3">
              <div
                class="flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center"
                :class="
                  variant === 'danger'
                    ? 'bg-red-100 dark:bg-red-900/30'
                    : variant === 'warning'
                    ? 'bg-yellow-100 dark:bg-yellow-900/30'
                    : 'bg-blue-100 dark:bg-blue-900/30'
                "
              >
                <component
                  :is="iconComponent"
                  class="w-5 h-5"
                  :class="
                    variant === 'danger'
                      ? 'text-red-600 dark:text-red-400'
                      : variant === 'warning'
                      ? 'text-yellow-600 dark:text-yellow-400'
                      : 'text-blue-600 dark:text-blue-400'
                  "
                />
              </div>
              <div class="flex-1">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ title }}
                </h3>
                <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
                  {{ message }}
                </p>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200 dark:border-gray-700">
              <Button variant="outline" @click="$emit('cancel')" :disabled="loading">
                {{ cancelText }}
              </Button>
              <Button
                :variant="variant === 'danger' ? 'danger' : 'primary'"
                @click="$emit('confirm')"
                :loading="loading"
              >
                {{ confirmText }}
              </Button>
            </div>
          </div>
        </Card>
      </Transition>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import Card from './Card.vue'
import Button from './Button.vue'
import { AlertTriangle, Trash2, Info } from 'lucide-vue-next'

interface Props {
  show: boolean
  title?: string
  message: string
  variant?: 'danger' | 'warning' | 'info'
  confirmText?: string
  cancelText?: string
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Confirm Action',
  variant: 'danger',
  confirmText: 'Confirm',
  cancelText: 'Cancel',
  loading: false
})

defineEmits<{
  confirm: []
  cancel: []
}>()

const iconComponent = computed(() => {
  switch (props.variant) {
    case 'danger':
      return h(Trash2)
    case 'warning':
      return h(AlertTriangle)
    default:
      return h(Info)
  }
})
</script>

