<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { login } from '@/lib/api'

const router = useRouter()
const route = useRoute()
const username = ref('armin')
const password = ref('dopadopa123')
const error = ref('')
const loading = ref(false)

async function onSubmit() {
  error.value = ''
  loading.value = true
  try {
    await login(username.value, password.value)
    const redirect = (route.query.redirect as string) || '/stats'
    router.replace(redirect)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="relative flex min-h-screen flex-col justify-center px-6 py-12">
    <div
      class="pointer-events-none absolute inset-0 opacity-40"
      aria-hidden="true"
      style="
        background-image:
          linear-gradient(var(--border) 1px, transparent 1px),
          linear-gradient(90deg, var(--border) 1px, transparent 1px);
        background-size: 48px 48px;
        mask-image: radial-gradient(ellipse 70% 60% at 50% 40%, black, transparent);
      "
    />

    <div class="page-enter relative mx-auto w-full max-w-md">
      <p class="font-display text-4xl font-semibold tracking-tight text-foreground sm:text-5xl">
        Agent Brain
      </p>
      <p class="page-lede mt-3 text-base">
        Local vault for agent sessions, context pressure, and prompt revisions.
      </p>

      <form class="mt-10 space-y-4" @submit.prevent="onSubmit">
        <label class="block text-sm">
          <span class="text-muted-foreground">Username</span>
          <input v-model="username" class="field-input mt-1" autocomplete="username" />
        </label>
        <label class="block text-sm">
          <span class="text-muted-foreground">Password</span>
          <input
            v-model="password"
            type="password"
            class="field-input mt-1"
            autocomplete="current-password"
          />
        </label>
        <p v-if="error" class="text-sm text-danger">{{ error }}</p>
        <button class="btn-primary w-full" :disabled="loading" type="submit">
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>
    </div>
  </div>
</template>
