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
        <h1 class="text-2xl font-semibold">Prompts</h1>
        <p class="text-sm text-muted-foreground">Raw prompts and revised variants</p>
      </div>
      <button
        type="button"
        class="ml-auto rounded-md bg-primary px-3 py-2 text-sm text-primary-foreground disabled:opacity-50"
        :disabled="creating"
        @click="addPrompt"
      >
        {{ creating ? 'Creating…' : 'New prompt' }}
      </button>
    </div>
    <p v-if="error" class="mt-4 text-sm text-red-400">{{ error }}</p>
    <p v-else-if="loading" class="mt-4 text-sm text-muted-foreground">Loading…</p>
    <div v-else class="mt-6 overflow-x-auto rounded-lg border border-border">
      <table class="min-w-full text-left text-sm">
        <thead class="bg-muted/40 text-muted-foreground">
          <tr>
            <th class="px-3 py-2 font-medium">Prompt</th>
            <th class="px-3 py-2 font-medium">Technically revised</th>
            <th class="px-3 py-2 font-medium">English revised</th>
            <th class="px-3 py-2 font-medium">Harness</th>
            <th class="px-3 py-2 font-medium">DateTime</th>
            <th class="px-3 py-2 font-medium w-20">Delete</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="p in prompts"
            :key="p.id"
            class="cursor-pointer border-t border-border hover:bg-muted/20"
            @click="openPrompt(p.id)"
          >
            <td class="max-w-xs px-3 py-2">{{ clipPromptCell(p.prompt) }}</td>
            <td class="max-w-xs px-3 py-2">{{ clipPromptCell(p.technically_revised) }}</td>
            <td class="max-w-xs px-3 py-2">{{ clipPromptCell(p.english_revised) }}</td>
            <td class="px-3 py-2">{{ p.harness || '—' }}</td>
            <td class="whitespace-nowrap px-3 py-2">{{ fmtDateTime(p.updated_at) }}</td>
            <td class="px-3 py-2">
              <button
                type="button"
                class="text-red-400 hover:text-red-300"
                @click="onDelete($event, p.id)"
              >
                Delete
              </button>
            </td>
          </tr>
          <tr v-if="prompts.length === 0">
            <td colspan="6" class="px-3 py-8 text-center text-muted-foreground">
              No prompts yet. Click New prompt to add one.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
