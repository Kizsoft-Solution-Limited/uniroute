<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white" id="page-title">Tunnel Management</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          View and delete tunnels
        </p>
      </div>
      <button
        v-if="selectedIds.length > 0"
        type="button"
        @click="confirmDeleteSelected"
        class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50 transition-colors"
      >
        Delete selected ({{ selectedIds.length }})
      </button>
    </div>

    <Card v-if="loading">
      <div class="text-center py-8" role="status" aria-live="polite">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" aria-hidden="true"></div>
        <p class="text-gray-500 dark:text-gray-400 mt-2">Loading tunnels...</p>
      </div>
    </Card>

    <Card v-else>
      <div class="overflow-x-auto">
        <table class="w-full" role="table" aria-label="Tunnels table">
          <thead>
            <tr class="border-b border-gray-200 dark:border-gray-700">
              <th class="px-4 py-3 text-left w-10">
                <input
                  ref="selectAllCheckbox"
                  type="checkbox"
                  :checked="selectedIds.length === selectableCount && selectableCount > 0"
                  :disabled="selectableCount === 0"
                  @change="toggleSelectAll"
                  class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                  aria-label="Select all tunnels"
                />
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Subdomain</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Public URL</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Protocol</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">User</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="t in tunnels" :key="t.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
              <td class="px-4 py-3">
                <input
                  type="checkbox"
                  :checked="selectedIds.includes(t.id)"
                  @change="toggleSelect(t.id)"
                  class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                  :aria-label="`Select ${t.subdomain}`"
                />
              </td>
              <td class="px-4 py-3 text-sm text-gray-900 dark:text-white font-mono">{{ t.subdomain }}</td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400 truncate max-w-[200px]" :title="t.public_url">{{ t.public_url }}</td>
              <td class="px-4 py-3">
                <span
                  :class="[
                    'px-2 py-1 text-xs font-medium rounded-full',
                    t.status === 'active'
                      ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400'
                      : 'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-400'
                  ]"
                >
                  {{ t.status }}
                </span>
              </td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">{{ t.protocol || 'http' }}</td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400 truncate max-w-[160px]" :title="(t.user_display || t.user_id) || ''">{{ t.user_display || t.user_id || '-' }}</td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                {{ formatDate(t.created_at) }}
              </td>
              <td class="px-4 py-3">
                <button
                  @click="confirmDeleteOne(t)"
                  class="text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300 text-sm font-medium"
                  :aria-label="`Delete tunnel ${t.subdomain}`"
                >
                  Delete
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="total > 0" class="mt-4 flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-700">
        <div class="text-sm text-gray-600 dark:text-gray-400">
          Showing {{ offset + 1 }} to {{ Math.min(offset + count, total) }} of {{ total }} tunnels
        </div>
        <div class="flex space-x-2">
          <button
            @click="previousPage"
            :disabled="offset === 0"
            class="px-3 py-1 text-sm rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 dark:hover:bg-gray-800"
            aria-label="Previous page"
          >
            Previous
          </button>
          <button
            @click="nextPage"
            :disabled="offset + count >= total"
            class="px-3 py-1 text-sm rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 dark:hover:bg-gray-800"
            aria-label="Next page"
          >
            Next
          </button>
        </div>
      </div>
    </Card>

    <div
      v-if="showDeleteModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
      @click.self="showDeleteModal = false"
      role="dialog"
      aria-labelledby="delete-modal-title"
      aria-modal="true"
    >
      <Card class="w-full max-w-md mx-4">
        <h2 id="delete-modal-title" class="text-xl font-bold text-gray-900 dark:text-white mb-2">Delete tunnel(s)?</h2>
        <p class="text-gray-600 dark:text-gray-400 mb-4">
          This will permanently delete {{ selectedIds.length }} tunnel(s). This cannot be undone.
        </p>
        <div class="flex space-x-3">
          <button
            @click="deleteSelected"
            :disabled="deleting"
            class="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
          <button
            @click="showDeleteModal = false"
            :disabled="deleting"
            class="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
          >
            Cancel
          </button>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import Card from '@/components/ui/Card.vue'
import { tunnelsApi, type Tunnel } from '@/services/api/tunnels'
import { useToast } from '@/composables/useToast'

const { showToast } = useToast()

const loading = ref(false)
const tunnels = ref<Tunnel[]>([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)
const count = ref(0)

const selectedIds = ref<string[]>([])
const showDeleteModal = ref(false)
const deleting = ref(false)

const selectableCount = computed(() => tunnels.value.length)
const selectAllCheckbox = ref<HTMLInputElement | null>(null)

watch([selectedIds, selectableCount], () => {
  const el = selectAllCheckbox.value
  if (!el) return
  el.indeterminate = selectedIds.value.length > 0 && selectedIds.value.length < selectableCount.value
})

onMounted(() => {
  loadTunnels()
})

const loadTunnels = async () => {
  loading.value = true
  try {
    const response = await tunnelsApi.listAdmin(limit.value, offset.value)
    tunnels.value = response.tunnels
    total.value = response.total
    count.value = response.count
  } catch (error: any) {
    showToast(error.message || 'Failed to load tunnels', 'error')
  } finally {
    loading.value = false
  }
}

const previousPage = () => {
  if (offset.value > 0) {
    offset.value = Math.max(0, offset.value - limit.value)
    loadTunnels()
  }
}

const nextPage = () => {
  if (offset.value + count.value < total.value) {
    offset.value += limit.value
    loadTunnels()
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

const toggleSelect = (id: string) => {
  const idx = selectedIds.value.indexOf(id)
  if (idx === -1) selectedIds.value = [...selectedIds.value, id]
  else selectedIds.value = selectedIds.value.filter((x) => x !== id)
}

const toggleSelectAll = () => {
  if (selectedIds.value.length === tunnels.value.length) {
    selectedIds.value = []
  } else {
    selectedIds.value = tunnels.value.map((t) => t.id)
  }
}

const confirmDeleteOne = (t: Tunnel) => {
  selectedIds.value = [t.id]
  showDeleteModal.value = true
}

const confirmDeleteSelected = () => {
  if (selectedIds.value.length === 0) return
  showDeleteModal.value = true
}

const deleteSelected = async () => {
  if (selectedIds.value.length === 0) return
  deleting.value = true
  try {
    const res = await tunnelsApi.deleteManyAdmin(selectedIds.value)
    showToast(
      res.deleted === selectedIds.value.length
        ? `${res.deleted} tunnel(s) deleted`
        : `Deleted ${res.deleted}; ${res.failed?.length ?? 0} failed`,
      res.deleted > 0 ? 'success' : 'error'
    )
    if (res.deleted > 0) {
      showDeleteModal.value = false
      selectedIds.value = []
      await loadTunnels()
    }
  } catch (error: any) {
    showToast(error.response?.data?.error || error.message || 'Failed to delete tunnels', 'error')
  } finally {
    deleting.value = false
  }
}
</script>
