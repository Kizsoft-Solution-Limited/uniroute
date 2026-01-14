<template>
  <div class="space-y-6">
    <div>
      <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-1">
        Profile Settings
      </h2>
      <p class="text-gray-600 dark:text-gray-400">
        Manage your account information and preferences
      </p>
    </div>

    <Card>
      <form @submit.prevent="updateProfile" class="space-y-6">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Email
          </label>
          <Input
            v-model="profile.email"
            type="email"
            required
            disabled
            class="bg-gray-50 dark:bg-gray-800"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            Email cannot be changed
          </p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Full Name
          </label>
          <Input
            v-model="profile.name"
            placeholder="John Doe"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Roles
          </label>
          <Input
            :value="profile.roles?.join(', ') || 'user'"
            disabled
            class="bg-gray-50 dark:bg-gray-800"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            Roles cannot be changed from profile settings
          </p>
        </div>

        <div class="flex items-center space-x-4 pt-4">
          <Button type="submit" :loading="saving">
            Save Changes
          </Button>
          <Button variant="outline" @click="resetForm">
            Cancel
          </Button>
        </div>
      </form>
    </Card>

    <!-- Change Password Section -->
    <Card>
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Change Password
      </h3>
      <form @submit.prevent="changePassword" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Current Password
          </label>
          <Input
            v-model="passwordForm.currentPassword"
            type="password"
            required
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            New Password
          </label>
          <Input
            v-model="passwordForm.newPassword"
            type="password"
            required
            minlength="8"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            Must be at least 8 characters
          </p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Confirm New Password
          </label>
          <Input
            v-model="passwordForm.confirmPassword"
            type="password"
            required
          />
        </div>
        <div class="flex items-center space-x-4 pt-4">
          <Button type="submit" :loading="changingPassword">
            Change Password
          </Button>
        </div>
      </form>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { authApi } from '@/services/api/auth'

const authStore = useAuthStore()
const { showToast } = useToast()

const saving = ref(false)
const changingPassword = ref(false)

const profile = ref({
  email: '',
  name: '',
  roles: [] as string[]
})

const passwordForm = ref({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const originalProfile = ref({ ...profile.value })

onMounted(() => {
  loadProfile()
})

const loadProfile = () => {
  if (authStore.user) {
    profile.value = {
      email: authStore.user.email || '',
      name: authStore.user.name || '',
      roles: authStore.user.roles || ['user']
    }
    originalProfile.value = { ...profile.value }
  }
}

const updateProfile = async () => {
  saving.value = true
  try {
    // Only send name - roles are explicitly excluded for security
    const updatedUser = await authApi.updateProfile({
      name: profile.value.name
    })
    
    // Update auth store with new user data
    authStore.user = updatedUser
    
    showToast('Profile updated successfully', 'success')
    originalProfile.value = { ...profile.value }
  } catch (error: any) {
    showToast(error.message || 'Failed to update profile', 'error')
  } finally {
    saving.value = false
  }
}

const changePassword = async () => {
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    showToast('Passwords do not match', 'error')
    return
  }

  if (passwordForm.value.newPassword.length < 8) {
    showToast('Password must be at least 8 characters', 'error')
    return
  }

  changingPassword.value = true
  try {
    await authApi.changePassword(
      passwordForm.value.currentPassword,
      passwordForm.value.newPassword
    )
    showToast('Password changed successfully', 'success')
    passwordForm.value = {
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    }
  } catch (error: any) {
    showToast(error.response?.data?.error || error.message || 'Failed to change password', 'error')
  } finally {
    changingPassword.value = false
  }
}

const resetForm = () => {
  profile.value = { ...originalProfile.value }
}
</script>
