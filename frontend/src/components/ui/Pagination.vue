<template>
  <div
    v-if="total > 0"
    class="flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-700"
    role="navigation"
    aria-label="Pagination"
  >
    <!-- Results Info -->
    <div class="text-sm text-gray-600 dark:text-gray-400">
      Showing {{ startItem }} to {{ endItem }} of {{ total }} {{ itemLabel }}
    </div>

    <!-- Pagination Controls -->
    <div class="flex items-center space-x-2">
      <!-- Previous Button -->
      <button
        @click="previousPage"
        :disabled="currentPage === 1"
        class="px-3 py-1 text-sm rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
        aria-label="Previous page"
        :aria-disabled="currentPage === 1"
      >
        Previous
      </button>

      <!-- Page Numbers (for large datasets) -->
      <template v-if="showPageNumbers">
        <button
          v-for="page in visiblePages"
          :key="page"
          @click="goToPage(page)"
          :class="[
            'px-3 py-1 text-sm rounded-lg border transition-colors',
            page === currentPage
              ? 'bg-blue-600 text-white border-blue-600'
              : 'border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800'
          ]"
          :aria-label="`Go to page ${page}`"
          :aria-current="page === currentPage ? 'page' : undefined"
        >
          {{ page }}
        </button>
      </template>

      <!-- Next Button -->
      <button
        @click="nextPage"
        :disabled="currentPage >= totalPages"
        class="px-3 py-1 text-sm rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
        aria-label="Next page"
        :aria-disabled="currentPage >= totalPages"
      >
        Next
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  limit: number
  offset: number
  total: number
  itemLabel?: string
  showPageNumbers?: boolean
  maxVisiblePages?: number
}

const props = withDefaults(defineProps<Props>(), {
  itemLabel: 'items',
  showPageNumbers: false,
  maxVisiblePages: 5
})

const emit = defineEmits<{
  'update:offset': [offset: number]
  'page-change': [page: number]
}>()

const currentPage = computed(() => Math.floor(props.offset / props.limit) + 1)
const totalPages = computed(() => Math.ceil(props.total / props.limit))
const startItem = computed(() => props.offset + 1)
const endItem = computed(() => Math.min(props.offset + props.limit, props.total))

const visiblePages = computed(() => {
  if (!props.showPageNumbers) return []
  
  const pages: number[] = []
  const maxPages = props.maxVisiblePages
  const current = currentPage.value
  const total = totalPages.value

  if (total <= maxPages) {
    // Show all pages if total is less than max
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // Show pages around current page
    const half = Math.floor(maxPages / 2)
    let start = Math.max(1, current - half)
    let end = Math.min(total, current + half)

    // Adjust if we're near the start or end
    if (end - start < maxPages - 1) {
      if (start === 1) {
        end = Math.min(total, start + maxPages - 1)
      } else {
        start = Math.max(1, end - maxPages + 1)
      }
    }

    for (let i = start; i <= end; i++) {
      pages.push(i)
    }
  }

  return pages
})

const previousPage = () => {
  if (currentPage.value > 1) {
    const newOffset = Math.max(0, props.offset - props.limit)
    emit('update:offset', newOffset)
    emit('page-change', currentPage.value - 1)
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    const newOffset = props.offset + props.limit
    emit('update:offset', newOffset)
    emit('page-change', currentPage.value + 1)
  }
}

const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    const newOffset = (page - 1) * props.limit
    emit('update:offset', newOffset)
    emit('page-change', page)
  }
}
</script>

