<script setup lang="ts">
import type { Component } from "vue"
import Card from "@/components/Card.vue"

defineProps<{
  icon: Component
  title: string
  description?: string
}>()
</script>

<template>
  <Card class="space-y-4">
    <div class="flex items-center gap-2">
      <component :is="icon" :size="18" class="text-primary" />
      <h2 class="text-lg font-bold">{{ title }}</h2>
    </div>

    <p v-if="description" class="text-muted text-sm">{{ description }}</p>

    <slot />
  </Card>
</template>
