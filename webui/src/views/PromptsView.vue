<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  clipPromptCell,
  createPrompt,
  deletePrompt,
  fmtDateTime,
  listPrompts,
  type PromptRow,
} from '@/lib/api'

const router = useRouter()
const prompts = ref<PromptRow[]>([])
const error = ref('')
const loading = ref(false)
const creating = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await listPrompts()
    prompts.value = data.prompts
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
}

function openPrompt(id: string) {
  router.push({ name: 'prompt', params: { id } })
}

async function addPrompt() {
  creating.value = true
  error.value = ''
  try {
    const { id } = await createPrompt({
      prompt: '',
      technically_revised: '',
      english_revised: '',
    })
    await router.push({ name: 'prompt', params: { id } })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to create'
  } finally {
    creating.value = false
  }
}

async function onDelete(e: Event, id: string) {
  e.stopPropagation()
  if (!confirm('Delete this prompt?')) return
  error.value = ''
  try {
    await deletePrompt(id)
    prompts.value = prompts.value.filter((p) => p.id !== id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Delete failed'
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex flex-wrap items-end gap-3">
      <div>
        <h1 class="page-title">Prompts</h1>
        <p class="page-lede">Raw prompts and revised variants.</p>
      </div>
      <button type="button" class="btn-primary ml-auto" :disabled="creating" @click="addPrompt">
        {{ creating ? 'Creating…' : 'New prompt' }}
      </button>
    </div>

    <p v-if="error" class="mt-4 text-sm text-danger">{{ error }}</p>
    <p v-else-if="loading" class="mt-4 text-sm text-muted-foreground">Loading…</p>

    <div v-else class="table-shell mt-6">
      <table class="data-table">
        <thead>
          <tr>
            <th>Prompt</th>
            <th>Technically revised</th>
            <th>English revised</th>
            <th>Harness</th>
            <th>DateTime</th>
            <th class="w-20">Delete</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="p in prompts"
            :key="p.id"
            class="cursor-pointer"
            @click="openPrompt(p.id)"
          >
            <td class="max-w-xs">{{ clipPromptCell(p.prompt) }}</td>
            <td class="max-w-xs">{{ clipPromptCell(p.technically_revised) }}</td>
            <td class="max-w-xs">{{ clipPromptCell(p.english_revised) }}</td>
            <td>{{ p.harness || '—' }}</td>
            <td class="whitespace-nowrap font-mono text-xs">{{ fmtDateTime(p.updated_at) }}</td>
            <td>
              <button
                type="button"
                class="text-danger hover:underline"
                @click="onDelete($event, p.id)"
              >
                Delete
              </button>
            </td>
          </tr>
          <tr v-if="prompts.length === 0">
            <td colspan="6" class="empty-hint">No prompts yet. Click New prompt to add one.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
