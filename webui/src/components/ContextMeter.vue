<script setup lang="ts">
import { computed } from 'vue'
import { fmtPct } from '@/lib/api'

const props = withDefaults(
  defineProps<{
    percent: number | null | undefined
    size?: 'sm' | 'lg'
    showLabel?: boolean
  }>(),
  {
    size: 'sm',
    showLabel: true,
  },
)

const clamped = computed(() => {
  if (props.percent == null || Number.isNaN(props.percent)) return null
  return Math.max(0, Math.min(100, props.percent))
})

const fillStyle = computed(() => ({
  width: clamped.value == null ? '0%' : `${clamped.value}%`,
}))
</script>

<template>
  <div class="ink-meter" :class="size === 'lg' ? 'ink-meter-lg' : 'ink-meter-sm'">
    <div class="ink-meter-track" role="progressbar" :aria-valuenow="clamped ?? undefined" aria-valuemin="0" aria-valuemax="100">
      <div class="ink-meter-fill" :style="fillStyle" />
    </div>
    <span v-if="showLabel" class="ink-meter-label font-mono">{{ fmtPct(percent) }}</span>
  </div>
</template>

<style scoped>
.ink-meter {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  min-width: 0;
}

.ink-meter-track {
  flex: 1;
  min-width: 4.5rem;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--muted) 80%, white);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--border) 70%, transparent);
}

.ink-meter-sm .ink-meter-track {
  height: 0.5rem;
}

.ink-meter-lg .ink-meter-track {
  height: 0.875rem;
}

.ink-meter-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #0a6b62 0%, var(--primary) 100%);
  transition: width 320ms ease;
}

@media (prefers-reduced-motion: reduce) {
  .ink-meter-fill {
    transition: none;
  }
}

.ink-meter-label {
  flex-shrink: 0;
  font-size: 0.75rem;
  color: var(--muted-foreground);
  font-variant-numeric: tabular-nums;
}

.ink-meter-lg .ink-meter-label {
  font-size: 0.875rem;
  color: var(--foreground);
  font-weight: 500;
}
</style>
