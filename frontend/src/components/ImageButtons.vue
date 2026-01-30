<script setup lang="ts">
import useToaster from "@/composables/useToaster"
import type { Image } from "@/utils/type"
import { Trash2Icon, SquareArrowOutUpRightIcon, CopyIcon } from "lucide-vue-next"

const { image } = defineProps<{
  image: Image
}>()

const { push } = useToaster()

const copyToClipboard = () => {
  navigator.clipboard.writeText(image.url)
  push({ title: "Copied to clipboard!", colour: "info", delay: 4000 })
}

defineEmits<{
  delete: [id: string]
}>()
</script>

<template>
  <div
    class="bg-background flex w-max items-center justify-center rounded-md p-1 *:cursor-pointer *:rounded-md *:p-1 *:hover:bg-black/10">
    <button @click="copyToClipboard"><CopyIcon class="stroke-black/80" /></button>
    <a :href="image.url" target="_blank">
      <SquareArrowOutUpRightIcon class="stroke-black/80" />
    </a>
    <button
      @click="
        () => {
          $emit('delete', image.id)
          console.log('hi')
        }
      ">
      <Trash2Icon class="stroke-danger" />
    </button>
  </div>
</template>
