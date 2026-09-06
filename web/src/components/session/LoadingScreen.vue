<script setup lang="ts">
import { onMounted, onUnmounted, ref, useTemplateRef } from 'vue'

const original = useTemplateRef<globalThis.HTMLSpanElement>('original')
const replacement = useTemplateRef<globalThis.HTMLSpanElement>('replacement')
const wordShift = ref(0)
let observer: globalThis.ResizeObserver | undefined

function alignWords(): void {
  if (!original.value || !replacement.value) return
  wordShift.value = (replacement.value.getBoundingClientRect().width - original.value.getBoundingClientRect().width) / 2
}

onMounted(() => {
  alignWords()

  const ResizeObserverCtor = globalThis.ResizeObserver
  if (typeof ResizeObserverCtor === 'undefined') return

  observer = new ResizeObserverCtor(alignWords)
  if (original.value) observer.observe(original.value)
  if (replacement.value) observer.observe(replacement.value)
})
onUnmounted(() => observer?.disconnect())
</script>

<template>
  <div class="session-loading" role="status" aria-live="polite" aria-busy="true">
    <p class="sr-only">{{ $t('auth.securingRide') }}</p>
    <div class="session-loading__brand" :style="{ '--word-shift': `${wordShift}px` }" aria-hidden="true">
      <span class="session-loading__prefix">
        <span
          v-for="(letter, index) in $t('auth.loadingPrefix')"
          :key="index"
          class="session-loading__letter"
          :data-letter="letter"
          :style="{ '--letter-delay': `${index * 100}ms` }"
        >{{ letter }}</span>
      </span>
      <span class="session-loading__words">
        <span ref="original" class="session-loading__original" lang="zh-CN">
          <span
            v-for="(letter, index) in $t('auth.loadingOriginal')"
            :key="index"
            class="session-loading__letter"
            :data-letter="letter"
            :style="{ '--letter-delay': `${(index + 2) * 100}ms` }"
          >{{ letter }}</span>
        </span>
        <span ref="replacement" class="session-loading__replacement" lang="en">{{ $t('auth.loadingReplacement') }}</span>
      </span>
    </div>
  </div>
</template>
