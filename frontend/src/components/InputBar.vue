<script setup lang="ts">
import clsx from "clsx"

defineProps<{
  label: string
  id?: string
  className?: string
}>()
</script>

<template>
  <div>
    <input
      :id="id"
      v-bind="$attrs"
      :class="
        clsx(
          'block w-full appearance-none rounded-none border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-blue-500 focus:ring-blue-500 focus:outline-none sm:text-sm',
          className
        )
      " />
  </div>
</template>
