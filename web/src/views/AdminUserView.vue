<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AdminShell from '@/components/admin/AdminShell.vue'
import AdminUserProfilePage from '@/components/admin/users/AdminUserProfilePage.vue'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'

const route = useRoute()
const router = useRouter()
const userId = computed(() => String(route.params.userId ?? ''))
const ownsBack = computed(() => userId.value !== '')

function backToUsers(): void {
  void router.push('/admin/users')
}

useTelegramBackButton(ownsBack, backToUsers)
</script>

<template>
  <div class="page page--admin">
    <AdminShell compact>
      <AdminUserProfilePage :user-id="userId" @back="backToUsers" />
    </AdminShell>
  </div>
</template>
