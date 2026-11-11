<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'
import { clearToken, getToken } from '@/lib/api'

const route = useRoute()
const router = useRouter()
// Read route first so the computed always tracks navigation (avoid short-circuit
// skipping route when localStorage has no token yet on the login screen).
const authed = computed(() => route.name !== 'login' && !!getToken())

function navLinkClass(active: boolean) {
  return [
    'nav-link relative block rounded-md px-3 py-2 text-sm transition-colors',
    active ? 'nav-link-active font-semibold text-foreground' : 'text-muted-foreground hover:text-foreground',
  ]
}

function logout() {
  clearToken()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen">
    <template v-if="authed">
      <div class="flex min-h-screen">
        <aside
          class="flex w-56 shrink-0 flex-col border-r border-border bg-card/90 backdrop-blur-sm"
          aria-label="Sidebar"
        >
          <RouterLink to="/stats" class="border-b border-border px-4 py-5">
            <span class="font-display text-xl font-semibold tracking-tight text-foreground">
              Agent Brain
            </span>
            <span class="mt-1 block text-xs text-muted-foreground">Ink archive</span>
          </RouterLink>

          <nav class="flex flex-1 flex-col gap-1 p-2">
            <RouterLink to="/stats" :class="navLinkClass(route.path === '/stats')">
              <span
                v-if="route.path === '/stats'"
                class="absolute inset-y-1.5 left-0 w-0.5 rounded-full bg-primary"
                aria-hidden="true"
              />
              Stats
            </RouterLink>
            <RouterLink
              to="/sessions"
              :class="navLinkClass(route.path.startsWith('/sessions'))"
            >
              <span
                v-if="route.path.startsWith('/sessions')"
                class="absolute inset-y-1.5 left-0 w-0.5 rounded-full bg-primary"
                aria-hidden="true"
              />
              Sessions
            </RouterLink>
            <RouterLink to="/prompts" :class="navLinkClass(route.path.startsWith('/prompts'))">
              <span
                v-if="route.path.startsWith('/prompts')"
                class="absolute inset-y-1.5 left-0 w-0.5 rounded-full bg-primary"
                aria-hidden="true"
              />
              Prompts
            </RouterLink>
          </nav>

          <button
            type="button"
            class="border-t border-border px-4 py-3 text-left text-sm text-muted-foreground transition-colors hover:text-foreground"
            @click="logout"
          >
            Sign out
          </button>
        </aside>

        <main class="page-enter min-w-0 flex-1 px-6 py-7 sm:px-8">
          <RouterView />
        </main>
      </div>
    </template>

    <template v-else>
      <RouterView />
    </template>
  </div>
</template>
