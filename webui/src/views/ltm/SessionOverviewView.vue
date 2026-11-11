<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import ContextMeter from '@/components/ContextMeter.vue'
import SessionTabs from '@/components/SessionTabs.vue'
import { fmtNum, getSession, type SessionSummary } from '@/lib/api'

const route = useRoute()
const uuid = computed(() => route.params.uuid as string)
const session = ref<SessionSummary | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    session.value = await getSession(uuid.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed'
  }
})
</script>

<template>
  <div v-if="error" class="text-danger">{{ error }}</div>
  <div v-else-if="!session" class="text-muted-foreground">Loading…</div>
  <div v-else>
    <div>
      <h1 class="page-title">{{ session.name || session.uuid }}</h1>
      <p class="page-lede">{{ session.subtitle }}</p>
    </div>

    <SessionTabs :uuid="uuid" />

    <section class="mt-8 max-w-2xl">
      <p class="text-sm text-muted-foreground">Context fill</p>
      <div class="mt-3">
        <ContextMeter :percent="session.context_usage_percent" size="lg" />
      </div>
    </section>

    <dl class="mt-8 grid max-w-3xl gap-x-8 gap-y-4 text-sm sm:grid-cols-2">
      <div>
        <dt class="text-muted-foreground">Input tokens</dt>
        <dd class="mt-0.5 font-mono">{{ fmtNum(session.input_tokens) }}</dd>
      </div>
      <div>
        <dt class="text-muted-foreground">Output tokens</dt>
        <dd class="mt-0.5 font-mono">{{ fmtNum(session.output_tokens) }}</dd>
      </div>
      <div>
        <dt class="text-muted-foreground">Cache read / write</dt>
        <dd class="mt-0.5 font-mono">
          {{ fmtNum(session.cache_read_tokens) }} / {{ fmtNum(session.cache_write_tokens) }}
        </dd>
      </div>
      <div>
        <dt class="text-muted-foreground">Mode</dt>
        <dd class="mt-0.5">{{ session.unified_mode || '—' }}</dd>
      </div>
      <div>
        <dt class="text-muted-foreground">Status</dt>
        <dd class="mt-0.5">{{ session.status || '—' }}</dd>
      </div>
      <div>
        <dt class="text-muted-foreground">Workspace</dt>
        <dd class="mt-0.5 break-all">{{ session.workspace_path || session.workspace_id || '—' }}</dd>
      </div>
      <div>
        <dt class="text-muted-foreground">Turns / thinking</dt>
        <dd class="mt-0.5">{{ session.turn_count ?? 0 }} / {{ session.thinking_count ?? 0 }}</dd>
      </div>
      <div>
        <dt class="text-muted-foreground">Lines + / −</dt>
        <dd class="mt-0.5 font-mono">
          +{{ session.total_lines_added }} / −{{ session.total_lines_removed }}
        </dd>
      </div>
      <div>
        <dt class="text-muted-foreground">Files changed</dt>
        <dd class="mt-0.5">{{ session.files_changed_count }}</dd>
      </div>
    </dl>
  </div>
</template>
