<script setup lang="ts">
import type { Component } from "vue"
import Card from "@/components/Card.vue"

defineProps<{
  label: string
  icon: Component
  value: string | null
}>()
</script>

<template>
  <Card class="space-y-3">
    <div class="flex items-center justify-between">
      <span class="text-muted text-xs font-semibold tracking-wide uppercase">{{ label }}</span>
      <component :is="icon" :size="16" class="text-primary" />
    </div>

    <h2 v-if="value !== null" class="text-2xl font-bold leading-none">{{ value }}</h2>
    <div v-else class="bg-surface-2 h-6 w-2/3 animate-pulse rounded"></div>
  </Card>
</template>
