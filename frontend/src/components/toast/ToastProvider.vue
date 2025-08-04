<script setup lang="ts">
import { TransitionGroup } from "vue"
import Toast from "./Toast.vue"
import useToaster from "@/composables/useToaster"

const { toasts, pop } = useToaster()
</script>

<template>
  <TransitionGroup name="list" tag="div" class="absolute top-4 right-4 flex flex-col gap-2">
    <Toast v-for="toast in toasts" :toast="toast" :key="toast.id" @dismiss="id => pop(id)" />
  </TransitionGroup>
  <slot />
</template>

<style scoped>
.list-move,
.list-enter-active,
.list-leave-active {
  transition: all 0.3s ease;
}

.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateX(60px);
}

.list-leave-active {
  position: absolute;
}
</style>
