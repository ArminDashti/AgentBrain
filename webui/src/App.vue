<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'
import { clearToken, getToken } from '@/lib/api'

const route = useRoute()
const router = useRouter()
const authed = computed(() => !!getToken() && route.name !== 'login')

const sessionUuid = computed(() =>
  typeof route.params.uuid === 'string' ? route.params.uuid : null,
)

function navLinkClass(active: boolean) {
  return [
    'block rounded-md px-3 py-2 text-sm transition-colors',
    active ? 'bg-muted text-foreground font-medium' : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground',
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
          class="flex w-56 shrink-0 flex-col border-r border-border bg-card/80 backdrop-blur"
          aria-label="Sidebar"
        >
          <RouterLink
            to="/stats"
            class="border-b border-border px-4 py-4 text-sm font-semibold tracking-wide text-primary"
          >
            Agent Brain
          </RouterLink>

          <nav class="flex flex-1 flex-col gap-1 p-2">
            <RouterLink to="/stats" :class="navLinkClass(route.path === '/stats')">
              Stats
            </RouterLink>
            <RouterLink
              to="/sessions"
              :class="navLinkClass(route.path.startsWith('/sessions') && !sessionUuid)"
            >
              Sessions
            </RouterLink>
            <RouterLink to="/prompts" :class="navLinkClass(route.path.startsWith('/prompts'))">
              Prompts
            </RouterLink>

            <template v-if="sessionUuid">
              <p class="mt-4 px-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Session
              </p>
              <RouterLink
                :to="`/sessions/${sessionUuid}`"
                :class="navLinkClass(route.name === 'session')"
              >
                Overview
              </RouterLink>
              <RouterLink
                :to="`/sessions/${sessionUuid}/thinking`"
                :class="navLinkClass(route.name === 'session-thinking')"
              >
                Thinking
              </RouterLink>
              <RouterLink
                :to="`/sessions/${sessionUuid}/turns`"
                :class="navLinkClass(route.name === 'session-turns')"
              >
                Turns
              </RouterLink>
            </template>
          </nav>

          <button
            type="button"
            class="border-t border-border px-4 py-3 text-left text-sm text-muted-foreground hover:text-foreground"
            @click="logout"
          >
            Sign out
          </button>
        </aside>

        <main class="min-w-0 flex-1 px-6 py-6">
          <RouterView />
        </main>
      </div>
    </template>

    <template v-else>
      <RouterView />
    </template>
  </div>
</template>
