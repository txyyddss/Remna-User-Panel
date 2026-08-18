<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import { adminOperationsApi, type BackupRun } from '@/api/adminOperations'
import { ApiError } from '@/api/http'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { localizedError, useI18n } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { formatBytes } from '@/utils/format'

const emit = defineEmits<{ uploaded: [backup: BackupRun] }>()
const { t } = useI18n()
const file = shallowRef<globalThis.File | null>(null)
const sha256 = shallowRef('')
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const completed = shallowRef<BackupRun | null>(null)
let uploadKey: string | undefined

const normalizedSHA = computed(() => sha256.value.trim().toLowerCase())
const hashValid = computed(() => normalizedSHA.value === '' || /^[a-f0-9]{64}$/.test(normalizedSHA.value))
const canUpload = computed(() => file.value !== null && hashValid.value && !busy.value)

watch([file, sha256], () => {
  uploadKey = undefined
  completed.value = null
  error.value = null
})

async function upload(): Promise<void> {
  const selected = file.value
  if (!selected || !canUpload.value) return
  busy.value = true
  error.value = null
  uploadKey ??= createUuid()
  try {
    const backup = await adminOperationsApi.uploadBackup(selected, normalizedSHA.value, uploadKey)
    uploadKey = undefined
    completed.value = backup
    file.value = null
    sha256.value = ''
    emit('uploaded', backup)
  } catch (caught) {
    if (caught instanceof ApiError) uploadKey = undefined
    error.value = localizedError(caught, 'backupUpload.failed')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="backup-upload">
    <div class="admin-profile-section__heading">
      <div><h3>{{ t('backupUpload.title') }}</h3><p>{{ t('backupUpload.copy') }}</p></div>
    </div>
    <UFileUpload v-model="file" accept=".db,.sqlite,.sqlite3,application/vnd.sqlite3,application/x-sqlite3" variant="area" :label="t('backupUpload.choose')" :description="t('backupUpload.chooseHint')" :disabled="busy" />
    <p v-if="file" class="admin-profile-empty">{{ t('backupUpload.selected', { name: file.name, size: formatBytes(file.size) }) }}</p>
    <div class="backup-upload__actions">
      <UFormField name="sha256" :label="t('backupUpload.sha256')" :hint="t('backupUpload.optional')" :error="hashValid ? undefined : t('backupUpload.invalidHash')">
        <UInput v-model.trim="sha256" class="w-full" autocomplete="off" spellcheck="false" :disabled="busy" />
      </UFormField>
      <UButton icon="i-ph-upload-simple" :label="busy ? t('backupUpload.uploading') : t('backupUpload.upload')" :loading="busy" :disabled="!canUpload" @click="upload" />
    </div>
    <UProgress v-if="busy" animation="carousel" :aria-label="t('backupUpload.uploading')" />
    <InlineNotice v-if="completed" tone="success" :title="t('backupUpload.complete')">{{ t('backupUpload.completeHint', { id: completed.id }) }}</InlineNotice>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
  </section>
</template>
