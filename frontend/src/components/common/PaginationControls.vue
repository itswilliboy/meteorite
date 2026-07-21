<script setup lang="ts">
import { ChevronLeft, ChevronRight } from "lucide-vue-next"

const { page, hasPrev, hasNext } = defineProps<{
  page: number
  hasPrev: boolean
  hasNext: boolean
}>()

const emit = defineEmits<{
  "update:page": [page: number]
}>()
</script>

<template>
  <div class="flex items-center gap-1">
    <button
      :disabled="!hasPrev"
      @click="emit('update:page', page - 1)"
      class="border-border bg-surface text-muted hover:bg-surface-2 hover:text-foreground grid size-9 place-items-center rounded-lg border transition hover:cursor-pointer disabled:pointer-events-none disabled:opacity-40">
      <ChevronLeft :size="18" />
    </button>

    <p class="text-muted px-3 text-sm font-medium tabular-nums">
      Page <span class="text-foreground font-semibold">{{ page + 1 }}</span>
    </p>

    <button
      :disabled="!hasNext"
      @click="emit('update:page', page + 1)"
      class="border-border bg-surface text-muted hover:bg-surface-2 hover:text-foreground grid size-9 place-items-center rounded-lg border transition hover:cursor-pointer disabled:pointer-events-none disabled:opacity-40">
      <ChevronRight :size="18" />
    </button>
  </div>
</template>
