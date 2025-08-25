<script setup lang="ts">
import type { Image } from "@/utils/type"
import { Trash2 } from "lucide-vue-next"

defineEmits<{
  delete: [id: string]
}>()
defineProps<{
  image: Image
}>()
</script>

<template>
  <div
    class="bg-background flex w-max items-center justify-center rounded-md p-1 *:cursor-pointer *:rounded-md *:p-1 *:hover:bg-black/10">
    <button @click="$emit('delete', image.id)">
      <Trash2 class="stroke-danger" />
    </button>
  </div>
</template>
