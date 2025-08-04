<script setup lang="ts">
import type { Toast } from "@/utils/type.ts"
import { defineProps } from "vue"
import { X } from "lucide-vue-next"

const { toast } = defineProps<{ toast: Toast }>()
const stripeColour = `bg-${toast.colour ?? "info"}`

defineEmits<{
  dismiss: [id: number]
}>()
</script>

<template>
  <div>
    <div
      class="relative min-h-10 w-80 overflow-hidden rounded-md bg-white p-4 shadow-xl"
      :style="`--duration: ${toast.delay}ms`">
      <span class="flex justify-between">
        <span :id="toast.delay ? 'stripe' : ''" class="absolute bottom-0 left-0 h-full w-1" :class="`${stripeColour}`" />
        <h1 class="line-clamp-1 font-medium">{{ toast.title }}</h1>
        <button @click="$emit('dismiss', toast.id)" class="cursor-pointer">
          <X :size="16" />
        </button>
      </span>
      <p>{{ toast.desc }}</p>
    </div>
  </div>
</template>

<style scoped>
@keyframes stripe {
  from {
    height: 100%;
  }

  to {
    height: 0%;
  }
}

#stripe {
  animation-name: stripe;
  animation-timing-function: linear;
  animation-fill-mode: forwards;
  transform-origin: bottom;
  animation-duration: var(--duration);
}
</style>
