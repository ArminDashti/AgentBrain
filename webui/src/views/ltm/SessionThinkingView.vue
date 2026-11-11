<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import SessionTabs from '@/components/SessionTabs.vue'
import { fmtNum, getSessionThinking } from '@/lib/api'

const route = useRoute()
const uuid = computed(() => route.params.uuid as string)
const items = ref<Array<Record<string, unknown>>>([])
const error = ref('')

onMounted(async () => {
  try {
    const data = await getSessionThinking(uuid.value)
    items.value = data.thinking
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed'
  }
})
</script>

<template>
  <div>
    <div>
      <h1 class="page-title">Thinking</h1>
      <p class="page-lede">Model reasoning steps extracted for this session.</p>
    </div>

    <SessionTabs :uuid="uuid" />

    <p v-if="error" class="mt-6 text-danger">{{ error }}</p>

    <ol v-else class="mt-6 max-w-3xl space-y-4">
      <li v-for="(t, i) in items" :key="String(t.bubble_id)" class="read-block">
        <div class="flex flex-wrap gap-3 text-xs text-muted-foreground">
          <span class="font-mono">#{{ i + 1 }}</span>
          <span>{{ fmtNum(t.thinking_duration_ms as number | null) }} ms</span>
        </div>
        <pre class="mt-3">{{
          (t.thinking_text as string) || '(no thinking text in extract)'
        }}</pre>
      </li>
      <li v-if="items.length === 0" class="empty-hint rounded-md border border-border bg-card">
        No thinking steps for this session.
      </li>
    </ol>
  </div>
</template>
