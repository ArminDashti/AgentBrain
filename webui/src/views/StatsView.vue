<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import ContextMeter from '@/components/ContextMeter.vue'
import { getStats, type StatsResponse } from '@/lib/api'

const stats = ref<StatsResponse | null>(null)
const error = ref('')
const loading = ref(false)

const hasSessions = computed(() => (stats.value?.session_count ?? 0) > 0)

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
      <h1 class="page-title">Stats</h1>
      <p class="page-lede">Overview of long-term memory sessions.</p>
    </div>

    <p v-if="error" class="text-sm text-danger">{{ error }}</p>
    <p v-if="loading" class="text-sm text-muted-foreground">Loading…</p>

    <template v-if="stats">
      <RouterLink
        to="/sessions"
        class="block max-w-xl rounded-md border border-border bg-card/80 p-6 transition-colors hover:border-primary"
      >
        <p class="text-sm text-muted-foreground">Sessions archived</p>
        <p class="mt-2 font-display text-5xl font-semibold tracking-tight text-foreground">
          {{ stats.session_count }}
        </p>
        <div v-if="hasSessions" class="mt-6">
          <p class="mb-2 text-sm text-muted-foreground">Average context fill</p>
          <ContextMeter :percent="stats.avg_context_usage_percent" size="lg" />
        </div>
        <p class="mt-4 text-sm font-medium text-primary">Browse sessions</p>
      </RouterLink>
    </template>
  </div>
</template>
