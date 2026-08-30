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
    <div class="fixed inset-0 z-[999] flex items-center justify-center p-4">
      <Transition name="backdrop" appear>
        <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="$emit('dismiss')" />
      </Transition>

      <Transition name="panel" appear>
        <div class="bg-surface border-border/60 relative z-50 w-135 max-w-full rounded-2xl border p-6 shadow-2xl">
          <h1 class="text-2xl font-semibold">{{ title }}</h1>
          <p v-if="description" class="text-muted mt-1.5 text-sm">{{ description }}</p>

          <div
            class="mt-6 grid w-full grid-cols-2 gap-2 text-center *:flex *:cursor-pointer *:items-center *:justify-center *:gap-1.5 *:rounded-xl *:p-3 *:text-sm *:font-semibold *:transition *:hover:opacity-90">
            <button class="bg-surface-2 text-foreground" @click="$emit('dismiss')">
              <XIcon class="size-4" />
              Cancel
            </button>
            <button
              :class="[colour, 'text-white']"
              @click="
                () => {
                  confirmAction()
                  $emit('dismiss')
                }
              ">
              <component :is="confirmIcon" class="size-4" />
              {{ confirmText }}
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>
