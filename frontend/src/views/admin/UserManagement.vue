<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white" id="page-title">User Management</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          Manage users and their roles
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

    <!-- Loading State -->
    <Card v-if="loading">
      <div class="text-center py-8" role="status" aria-live="polite">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" aria-hidden="true"></div>
        <p class="text-gray-500 dark:text-gray-400 mt-2">Loading users...</p>
      </div>
    </Card>

    <!-- Users Table -->
    <Card v-else>
      <div class="overflow-x-auto">
        <table class="w-full" role="table" aria-label="Users table">
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
                  aria-label="Select all users"
                />
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Email</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Roles</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Email Verified</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="user in users" :key="user.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
              <td class="px-4 py-3">
                <input
                  type="checkbox"
                  :checked="selectedIds.includes(user.id)"
                  :disabled="user.id === currentUserId"
                  @change="toggleSelect(user.id)"
                  class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500 disabled:opacity-50"
                  :aria-label="`Select ${user.email}`"
                />
              </td>
              <td class="px-4 py-3 text-sm text-gray-900 dark:text-white">{{ user.email }}</td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">{{ user.name || '-' }}</td>
              <td class="px-4 py-3">
                <div class="flex flex-wrap gap-2">
                  <span
                    v-for="role in user.roles"
                    :key="role"
                    :class="[
                      'px-2 py-1 text-xs font-medium rounded-full',
                      role === 'admin'
                        ? 'bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-400'
                        : 'bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400'
                    ]"
                  >
                    {{ role }}
                  </span>
                </div>
              </td>
              <td class="px-4 py-3">
                <span
                  :class="[
                    'px-2 py-1 text-xs font-medium rounded-full',
                    user.email_verified
                      ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400'
                      : 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-400'
                  ]"
                >
                  {{ user.email_verified ? 'Verified' : 'Unverified' }}
                </span>
              </td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                {{ formatDate(user.created_at) }}
              </td>
              <td class="px-4 py-3">
                <button
                  @click="openEditModal(user)"
                  class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 text-sm font-medium"
                  :aria-label="`Edit roles for ${user.email}`"
                >
                  Edit Roles
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="total > 0" class="mt-4 flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-700">
        <div class="text-sm text-gray-600 dark:text-gray-400">
          Showing {{ offset + 1 }} to {{ Math.min(offset + count, total) }} of {{ total }} users
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

    <!-- Edit Roles Modal -->
    <div
      v-if="showEditModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
      @click.self="closeEditModal"
      role="dialog"
      aria-labelledby="modal-title"
      aria-modal="true"
    >
      <Card class="w-full max-w-md mx-4">
        <h2 id="modal-title" class="text-xl font-bold text-gray-900 dark:text-white mb-4">Edit User Roles</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              User: {{ editingUser?.email }}
            </label>
            <div class="space-y-2">
              <label class="flex items-center space-x-2">
                <input
                  type="checkbox"
                  v-model="selectedRoles"
                  value="user"
                  class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                />
                <span class="text-sm text-gray-700 dark:text-gray-300">User</span>
              </label>
              <label class="flex items-center space-x-2">
                <input
                  type="checkbox"
                  v-model="selectedRoles"
                  value="admin"
                  class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                />
                <span class="text-sm text-gray-700 dark:text-gray-300">Admin</span>
              </label>
            </div>
          </div>
          <div class="flex space-x-3">
            <button
              @click="saveRoles"
              :disabled="saving || selectedRoles.length === 0"
              class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
            <button
              @click="closeEditModal"
              class="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      </Card>
    </div>

    <!-- Delete confirmation modal -->
    <div
      v-if="showDeleteModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
      @click.self="showDeleteModal = false"
      role="dialog"
      aria-labelledby="delete-modal-title"
      aria-modal="true"
    >
      <Card class="w-full max-w-md mx-4">
        <h2 id="delete-modal-title" class="text-xl font-bold text-gray-900 dark:text-white mb-2">Delete users?</h2>
        <p class="text-gray-600 dark:text-gray-400 mb-4">
          This will permanently delete {{ selectedIds.length }} user(s) and all their data (API keys, tunnels, conversations, provider keys, etc.). This cannot be undone.
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
import { usersApi, type User } from '@/services/api/users'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'

const { showToast } = useToast()
const authStore = useAuthStore()

const loading = ref(false)
const users = ref<User[]>([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)
const count = ref(0)

const showEditModal = ref(false)
const editingUser = ref<User | null>(null)
const selectedRoles = ref<string[]>([])
const saving = ref(false)

const selectedIds = ref<string[]>([])
const showDeleteModal = ref(false)
const deleting = ref(false)

const currentUserId = computed(() => authStore.user?.id ?? null)
const selectableCount = computed(() => users.value.filter((u) => u.id !== currentUserId.value).length)
const allSelectedDisabled = computed(() => selectableCount.value === 0)
const selectAllCheckbox = ref<HTMLInputElement | null>(null)

watch([selectedIds, selectableCount], () => {
  const el = selectAllCheckbox.value
  if (!el) return
  el.indeterminate = selectedIds.value.length > 0 && selectedIds.value.length < selectableCount.value
})

onMounted(() => {
  loadUsers()
})

const loadUsers = async () => {
  loading.value = true
  try {
    const response = await usersApi.list(limit.value, offset.value)
    users.value = response.users
    total.value = response.total
    count.value = response.count
  } catch (error: any) {
    showToast(error.message || 'Failed to load users', 'error')
  } finally {
    loading.value = false
  }
}

const openEditModal = (user: User) => {
  editingUser.value = user
  selectedRoles.value = [...user.roles]
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
  editingUser.value = null
  selectedRoles.value = []
}

const saveRoles = async () => {
  if (!editingUser.value || selectedRoles.value.length === 0) return

  saving.value = true
  try {
    await usersApi.updateRoles(editingUser.value.id, selectedRoles.value)
    showToast('User roles updated successfully', 'success')
    closeEditModal()
    await loadUsers()
  } catch (error: any) {
    showToast(error.message || 'Failed to update user roles', 'error')
  } finally {
    saving.value = false
  }
}

const previousPage = () => {
  if (offset.value > 0) {
    offset.value = Math.max(0, offset.value - limit.value)
    loadUsers()
  }
}

const nextPage = () => {
  if (offset.value + count.value < total.value) {
    offset.value += limit.value
    loadUsers()
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
  const selectable = users.value.filter((u) => u.id !== currentUserId.value).map((u) => u.id)
  if (selectedIds.value.length === selectable.length) {
    selectedIds.value = []
  } else {
    selectedIds.value = [...selectable]
  }
}

const confirmDeleteSelected = () => {
  if (selectedIds.value.length === 0) return
  showDeleteModal.value = true
}

const deleteSelected = async () => {
  if (selectedIds.value.length === 0) return
  deleting.value = true
  try {
    const res = await usersApi.deleteMany(selectedIds.value)
    showToast(
      res.deleted === selectedIds.value.length
        ? `${res.deleted} user(s) deleted`
        : `Deleted ${res.deleted}; ${res.failed?.length ?? 0} failed${res.error ? `: ${res.error}` : ''}`,
      res.deleted > 0 ? 'success' : 'error'
    )
    if (res.deleted > 0) {
      showDeleteModal.value = false
      selectedIds.value = []
      await loadUsers()
    }
  } catch (error: any) {
    showToast(error.response?.data?.error || error.message || 'Failed to delete users', 'error')
  } finally {
    deleting.value = false
  }
}
</script>

