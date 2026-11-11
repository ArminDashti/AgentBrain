<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import ContextMeter from '@/components/ContextMeter.vue'
import { fmtNum, listSessions, type SessionSummary } from '@/lib/api'

const sessions = ref<SessionSummary[]>([])
const q = ref('')
const error = ref('')
const loading = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await listSessions(q.value.trim())
    sessions.value = data.sessions
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex flex-wrap items-end gap-3">
      <div>
        <h1 class="page-title">Sessions</h1>
        <p class="page-lede">Cursor sessions from state.vscdb — context fill, tokens, mode.</p>
      </div>
      <form class="ml-auto flex gap-2" @submit.prevent="load">
        <input v-model="q" placeholder="Search name…" class="field-input w-48" />
        <button class="btn-primary" type="submit">Search</button>
      </form>
    </div>

    <p v-if="error" class="mt-4 text-sm text-danger">{{ error }}</p>
    <p v-else-if="loading" class="mt-4 text-sm text-muted-foreground">Loading…</p>

    <div v-else class="table-shell mt-6">
      <table class="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th class="w-48">Context</th>
            <th>In / Out tokens</th>
            <th>Mode</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in sessions" :key="s.uuid">
            <td>
              <RouterLink class="font-medium text-primary hover:underline" :to="`/sessions/${s.uuid}`">
                {{ s.name || s.uuid }}
              </RouterLink>
              <div class="text-xs text-muted-foreground">{{ s.subtitle }}</div>
            </td>
            <td>
              <ContextMeter :percent="s.context_usage_percent" />
            </td>
            <td class="font-mono text-xs">
              {{ fmtNum(s.input_tokens) }} / {{ fmtNum(s.output_tokens) }}
            </td>
            <td>{{ s.unified_mode || '—' }}</td>
            <td>{{ s.status || '—' }}</td>
          </tr>
          <tr v-if="sessions.length === 0">
            <td colspan="5" class="empty-hint">
              No sessions yet. On Windows run:
              <code class="font-mono text-foreground">agent-brain-dispatcher sync</code>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
