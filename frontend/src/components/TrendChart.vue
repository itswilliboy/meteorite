<script setup lang="ts">
import { computed, ref } from "vue"
import { format, parseISO } from "date-fns"

const {
  labels,
  values,
  variant = "area",
  formatValue = (v: number) => v.toLocaleString("en-GB"),
} = defineProps<{
  labels: string[]
  values: number[]
  variant?: "area" | "bar"
  formatValue?: (v: number) => string
}>()

const width = 600
const height = 160
const padding = 4

const max = computed(() => Math.max(...values, 1))

const points = computed(() =>
  values.map((v, i) => {
    const x = values.length > 1 ? (i / (values.length - 1)) * (width - padding * 2) + padding : width / 2
    const y = height - padding - (v / max.value) * (height - padding * 2)
    return { x, y, v, label: labels[i] }
  })
)

const linePath = computed(() => points.value.map((p, i) => `${i === 0 ? "M" : "L"} ${p.x} ${p.y}`).join(" "))

const areaPath = computed(() => {
  if (points.value.length === 0) return ""
  const first = points.value[0]
  const last = points.value[points.value.length - 1]
  return `${linePath.value} L ${last.x} ${height - padding} L ${first.x} ${height - padding} Z`
})

const barWidth = computed(() => (width - padding * 2) / values.length)

const hovered = ref<number | null>(null)
const active = computed(() => (hovered.value === null ? null : points.value[hovered.value]))

const formatLabel = (iso: string) => format(parseISO(iso), "MMM d")
</script>

<template>
  <div class="relative w-full">
    <svg :viewBox="`0 0 ${width} ${height}`" preserveAspectRatio="none" class="h-40 w-full overflow-visible">
      <defs>
        <linearGradient id="trend-fill" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" class="[stop-color:var(--color-primary)]" stop-opacity="0.35" />
          <stop offset="100%" class="[stop-color:var(--color-primary)]" stop-opacity="0" />
        </linearGradient>
      </defs>

      <template v-if="variant === 'area'">
        <path :d="areaPath" fill="url(#trend-fill)" />
        <path :d="linePath" fill="none" class="stroke-primary" stroke-width="2" stroke-linejoin="round" stroke-linecap="round" />
      </template>

      <template v-else>
        <rect
          v-for="(p, i) in points"
          :key="i"
          :x="p.x - barWidth / 2 + 1"
          :y="p.v === 0 ? height - padding - 1 : p.y"
          :width="Math.max(barWidth - 2, 1)"
          :height="p.v === 0 ? 1 : height - padding - p.y"
          rx="1.5"
          :class="hovered === i ? 'fill-primary' : 'fill-primary/40'" />
      </template>

      <!-- Hover targets -->
      <rect
        v-for="(p, i) in points"
        :key="`hit-${i}`"
        :x="i * barWidth"
        y="0"
        :width="barWidth"
        :height="height"
        fill="transparent"
        @mouseenter="hovered = i"
        @mouseleave="hovered = null" />

      <line
        v-if="active"
        :x1="active.x"
        y1="0"
        :x2="active.x"
        :y2="height - padding"
        class="stroke-border"
        stroke-width="1"
        stroke-dasharray="3 3" />
      <circle v-if="active" :cx="active.x" :cy="active.y" r="3.5" class="fill-primary stroke-surface" stroke-width="2" />
    </svg>

    <div
      v-if="active"
      class="bg-surface border-border pointer-events-none absolute top-0 -translate-x-1/2 rounded-lg border px-2.5 py-1.5 text-xs shadow-lg"
      :style="{ left: `${(active.x / width) * 100}%` }">
      <p class="text-muted font-medium">{{ formatLabel(active.label) }}</p>
      <p class="font-semibold">{{ formatValue(active.v) }}</p>
    </div>
  </div>
</template>
