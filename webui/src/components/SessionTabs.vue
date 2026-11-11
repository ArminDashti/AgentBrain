<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const props = defineProps<{
  uuid: string
}>()

const route = useRoute()

const tabs = computed(() => [
  { name: 'session' as const, label: 'Overview', to: `/sessions/${props.uuid}` },
  { name: 'session-thinking' as const, label: 'Thinking', to: `/sessions/${props.uuid}/thinking` },
  { name: 'session-turns' as const, label: 'Turns', to: `/sessions/${props.uuid}/turns` },
])
</script>

<template>
  <nav class="session-tabs" aria-label="Session sections">
    <RouterLink
      v-for="tab in tabs"
      :key="tab.name"
      :to="tab.to"
      class="session-tab"
      :class="{ 'session-tab-active': route.name === tab.name }"
    >
      {{ tab.label }}
    </RouterLink>
  </nav>
</template>
