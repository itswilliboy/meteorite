<script setup lang="ts">
import type { Component } from "vue"
import { Loader2 } from "lucide-vue-next"

const { variant = "primary", icon, loading = false, disabled = false } = defineProps<{
  variant?: "primary" | "secondary" | "danger"
  icon?: Component
  loading?: boolean
  disabled?: boolean
}>()

const variantClasses: Record<string, string> = {
  primary: "bg-primary text-white",
  secondary: "bg-surface-2 text-foreground",
  danger: "bg-danger text-white"
}
</script>

<template>
  <button
    :disabled="disabled || loading"
    :class="[
      variantClasses[variant],
      'flex items-center justify-center gap-1.5 rounded-lg px-4 py-2.5 text-sm font-semibold transition hover:cursor-pointer hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50',
    ]">
    <component :is="loading ? Loader2 : icon" v-if="loading || icon" :size="16" :class="{ 'animate-spin': loading }" />
    <slot />
  </button>
</template>
