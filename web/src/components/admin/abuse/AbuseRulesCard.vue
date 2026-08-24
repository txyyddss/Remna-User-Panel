<script setup lang="ts">
import { computed, reactive, shallowRef } from 'vue'
import type { AbuseRule } from '@/api/abuse'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useI18n } from '@/i18n'

const props = defineProps<{ rules: AbuseRule[]; whitelist: string[]; busy: boolean }>()
const emit = defineEmits<{ saveRule: [value: AbuseRule]; deleteRule: [id: string, revision: number]; whitelist: [id: string, enabled: boolean] }>()
const { t } = useI18n()
const draft = reactive<AbuseRule>({ id: '', name: '', expression: '', qpsLimit: 1, enabled: true, revision: 0 })
const remoteID = shallowRef('')
const deleting = shallowRef<Pick<AbuseRule, 'id' | 'name' | 'revision'> | null>(null)
const choices = computed(() => props.rules.map(item => ({ value: item.id, label: item.name })))

function choose(id: string): void {
  const rule = props.rules.find(item => item.id === id)
  if (rule) Object.assign(draft, rule)
}

function addRule(): void {
  Object.assign(draft, { id: '', name: '', expression: '', qpsLimit: 1, enabled: true, revision: 0 })
}

function requestDelete(): void {
  if (draft.id) deleting.value = { id: draft.id, name: draft.name, revision: draft.revision }
}

function deleteRule(): void {
  if (!deleting.value) return
  emit('deleteRule', deleting.value.id, deleting.value.revision)
  deleting.value = null
}
</script>

<template>
  <div class="rules-stack">
    <section class="card">
      <h3>{{ t('adminAbuse.rulesTitle') }}</h3>
      <UForm :state="draft" @submit="emit('saveRule', { ...draft })">
        <UFormField :label="t('adminAbuse.ruleSelect')">
          <USelect :items="choices" value-key="value" @update:model-value="choose" />
        </UFormField>
        <UFormField :label="t('adminAbuse.ruleName')">
          <UInput v-model="draft.name" />
        </UFormField>
        <UFormField :label="t('adminAbuse.expression')">
          <UInput v-model="draft.expression" />
        </UFormField>
        <UFormField :label="t('adminAbuse.ruleLimit')">
          <UInput v-model.number="draft.qpsLimit" type="number" />
        </UFormField>
        <USwitch v-model="draft.enabled" :label="t('adminAbuse.ruleEnabled')" />
        <div class="actions">
          <UButton type="submit" :loading="busy" :label="t('common.save')" />
          <UButton type="button" color="neutral" variant="outline" :label="t('adminAbuse.addRule')" @click="addRule" />
          <UButton v-if="draft.id" type="button" color="error" variant="ghost" :label="t('adminAbuse.deleteRule')" @click="requestDelete" />
        </div>
      </UForm>
    </section>

    <section class="card">
      <h3>{{ t('adminAbuse.whitelistTitle') }}</h3>
      <div class="actions">
        <UInput v-model="remoteID" />
        <UButton :label="t('adminAbuse.addWhitelist')" @click="remoteID && emit('whitelist', remoteID, true)" />
      </div>
      <div v-for="id in whitelist" :key="id" class="whitelist">
        <span>{{ id }}</span>
        <UButton
          color="error"
          variant="ghost"
          icon="i-ph-trash"
          :aria-label="t('adminAbuse.removeWhitelist')"
          @click="emit('whitelist', id, false)"
        />
      </div>
    </section>
    <ConfirmDialog
      :open="Boolean(deleting)"
      :title="t('adminAbuse.deleteRuleTitle', { name: deleting?.name ?? '' })"
      :description="t('adminAbuse.deleteRuleDescription')"
      :confirm-label="t('adminAbuse.deleteRule')"
      :busy="busy"
      danger
      @update:open="!$event && (deleting = null)"
      @confirm="deleteRule"
    />
  </div>
</template>

<style scoped>
.rules-stack,
.card,
.card :deep(form) {
  display: grid;
  gap: 0.75rem;
}

.card {
  padding: 1rem;
  border: 1px solid var(--line);
  border-radius: var(--radius-panel);
  background: var(--surface-raised);
}

.actions,
.whitelist {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.65rem;
}

.actions {
  flex-wrap: wrap;
}

.actions > :first-child {
  min-width: 0;
}

.whitelist {
  min-height: 44px;
  border-top: 1px solid var(--line);
}
</style>
