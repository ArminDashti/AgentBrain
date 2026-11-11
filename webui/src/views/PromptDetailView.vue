<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fmtDateTime, getPrompt, updatePrompt } from '@/lib/api'

const route = useRoute()
const id = computed(() => (typeof route.params.id === 'string' ? route.params.id : ''))

const prompt = ref('')
const technicallyRevised = ref('')
const englishRevised = ref('')
const harness = ref('')
const updatedAt = ref('')
const error = ref('')
const loading = ref(false)
const saving = ref(false)
const saved = ref(false)

async function load() {
  if (!id.value) return
  loading.value = true
  error.value = ''
  saved.value = false
  try {
    const row = await getPrompt(id.value)
    prompt.value = row.prompt
    technicallyRevised.value = row.technically_revised
    englishRevised.value = row.english_revised
    harness.value = row.harness
    updatedAt.value = row.updated_at
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!id.value) return
  saving.value = true
  error.value = ''
  saved.value = false
  try {
    await updatePrompt(id.value, {
      prompt: prompt.value,
      technically_revised: technicallyRevised.value,
      english_revised: englishRevised.value,
      harness: harness.value,
    })
    saved.value = true
    const row = await getPrompt(id.value)
    updatedAt.value = row.updated_at
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Save failed'
  } finally {
    saving.value = false
  }
}

onMounted(load)
watch(id, load)
</script>

<template>
  <div>
    <RouterLink to="/prompts" class="text-sm text-primary hover:underline">← Prompts</RouterLink>
    <div class="mt-3 flex flex-wrap items-end gap-3">
      <div>
        <h1 class="page-title">Prompt</h1>
        <p v-if="updatedAt" class="page-lede">
          Updated {{ fmtDateTime(updatedAt) }}
          <span v-if="harness"> · Harness: {{ harness }}</span>
        </p>
      </div>
      <button
        type="button"
        class="btn-primary ml-auto"
        :disabled="saving || loading"
        @click="save"
      >
        {{ saving ? 'Saving…' : 'Save' }}
      </button>
    </div>
    <p v-if="saved" class="mt-2 text-sm text-success">Saved.</p>
    <p v-if="error" class="mt-2 text-sm text-danger">{{ error }}</p>
    <p v-else-if="loading" class="mt-4 text-sm text-muted-foreground">Loading…</p>
    <form v-else class="mt-6 flex max-w-3xl flex-col gap-4" @submit.prevent="save">
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-muted-foreground">Prompt</span>
        <textarea v-model="prompt" rows="6" class="field-input font-mono" />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-muted-foreground">Technically revised</span>
        <textarea v-model="technicallyRevised" rows="6" class="field-input font-mono" />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="font-medium text-muted-foreground">English revised</span>
        <textarea v-model="englishRevised" rows="6" class="field-input font-mono" />
      </label>
    </form>
  </div>
</template>
