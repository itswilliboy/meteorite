<script setup lang="ts">
import clsx from "clsx"
import { useAttrs } from "vue"

defineProps<{
  label: string
  // placeholder?: string
  id?: string
  className?: string
  // name?: string
  // type?: string
  // required?: boolean
  // class?: string
}>()

console.log(useAttrs())
</script>

<template>
  <div>
    <label :for="id" class="sr-only" v-if="label">{{ label }}</label>

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
