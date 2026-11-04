<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getStats, type StatsResponse } from '@/lib/api'

const stats = ref<StatsResponse | null>(null)
const error = ref('')
const loading = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    stats.value = await getStats()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-2xl font-semibold">Stats</h1>
      <p class="text-sm text-muted-foreground">Overview of long-term memory sessions.</p>
    </div>

    <p v-if="error" class="text-sm text-red-400">{{ error }}</p>
    <p v-if="loading" class="text-sm text-muted-foreground">Loading…</p>

    <template v-if="stats">
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <RouterLink
          to="/sessions"
          class="rounded-lg border border-border bg-card p-4 hover:border-primary"
        >
          <p class="text-xs uppercase tracking-wide text-muted-foreground">LTM sessions</p>
          <p class="mt-2 text-3xl font-semibold text-primary">{{ stats.session_count }}</p>
        </RouterLink>
      </div>
    </template>
  </div>
</template>
