<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import SessionTabs from '@/components/SessionTabs.vue'
import { fmtNum, getSessionTurns } from '@/lib/api'

const route = useRoute()
const uuid = computed(() => route.params.uuid as string)
const items = ref<Array<Record<string, unknown>>>([])
const error = ref('')

onMounted(async () => {
  try {
    const data = await getSessionTurns(uuid.value)
    items.value = data.turns
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed'
  }
})
</script>

<template>
  <div>
    <div>
      <h1 class="page-title">Turns</h1>
      <p class="page-lede">Conversation and tool turns for this session.</p>
    </div>

    <SessionTabs :uuid="uuid" />

    <p v-if="error" class="mt-6 text-danger">{{ error }}</p>

    <div v-else class="mt-6 max-w-3xl space-y-4">
      <article v-for="t in items" :key="String(t.bubble_id)" class="read-block text-sm">
        <div class="flex flex-wrap gap-3 text-xs text-muted-foreground">
          <span class="rounded bg-muted px-2 py-0.5 font-medium text-foreground">{{
            t.turn_type
          }}</span>
          <span v-if="t.tool_name">tool: {{ t.tool_name }}</span>
          <span v-if="t.mcp_name">mcp: {{ t.mcp_name }}</span>
          <span class="font-mono">
            in/out {{ fmtNum(t.input_tokens as number | null) }}/{{
              fmtNum(t.output_tokens as number | null)
            }}
          </span>
          <span v-if="t.duration_ms != null" class="font-mono">
            {{ fmtNum(t.duration_ms as number | null) }} ms
          </span>
        </div>
        <pre v-if="t.text" class="mt-3 max-h-56 overflow-auto">{{ t.text }}</pre>
        <pre
          v-if="t.has_thinking && t.thinking_text"
          class="mt-3 border-t border-border pt-3 text-muted-foreground"
          >{{ t.thinking_text }}</pre
        >
      </article>
      <p v-if="items.length === 0" class="empty-hint rounded-md border border-border bg-card">
        No turns.
      </p>
    </div>
  </div>
</template>
