<script setup lang="ts">
import clsx from "clsx"
import { useSlots } from "vue"

defineOptions({ inheritAttrs: false })

defineProps<{
  label: string
  id?: string
  className?: string
}>()

const slots = useSlots()
</script>

<template>
  <div class="space-y-1.5">
    <label v-if="label" :for="id" class="text-foreground block text-sm font-medium">{{ label }}</label>

    <div class="relative">
      <input
        :id="id"
        v-bind="$attrs"
        :class="
          clsx(
            'border-border bg-surface text-foreground focus:border-primary focus:ring-primary/20 placeholder-muted block w-full rounded-lg border px-3 py-2.5 transition focus:ring-2 focus:outline-none',
            slots.suffix && 'pr-11',
            className
          )
        " />

      <div v-if="slots.suffix" class="absolute inset-y-0 right-0 flex items-center pr-1.5">
        <slot name="suffix" />
      </div>
    </div>
  </div>
</template>
