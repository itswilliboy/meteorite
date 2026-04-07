<script setup lang="ts">
import type { Component } from "vue"
import type { Colour } from "@/utils/type.ts"
import { XIcon } from "lucide-vue-next"

const { title = "Are you sure?", confirmColour } = defineProps<{
  title?: string
  description?: string
  confirmText: string
  confirmIcon: Component
  confirmColour: Colour
  confirmAction: (...args: unknown[]) => unknown
}>()

const colourMap: Record<Colour, string> = {
  info: "bg-info",
  success: "bg-success",
  warning: "bg-warning",
  danger: "bg-danger"
}
const colour = colourMap[confirmColour]

defineEmits<{
  dismiss: []
}>()
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-[999] flex items-center justify-center">
      <div class="absolute inset-0 z-40 bg-black/75" @click="$emit('dismiss')" />
      <div class="relative z-50 h-60 w-135 rounded-xl bg-white p-6 shadow-xl">
        <h1 class="text-3xl font-semibold">{{ title }}</h1>
        <p class="text-black/75">{{ description }}</p>
        <div
          class="absolute bottom-0 -ml-5 grid h-max w-full grid-cols-2 gap-2 p-5 text-center text-white *:flex *:cursor-pointer *:items-center *:justify-center *:gap-1 *:rounded-xl *:p-3 *:font-semibold">
          <button class="bg-primary" @click="$emit('dismiss')">
            <XIcon class="size-6" />
            Cancel
          </button>
          <button
            :class="colour"
            @click="
              () => {
                confirmAction()
                $emit('dismiss')
              }
            ">
            <component :is="confirmIcon" class="size-5" />
            {{ confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
