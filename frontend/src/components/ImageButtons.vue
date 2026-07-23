<script setup lang="ts">
import useToaster from "@/composables/useToaster"
import type { Image } from "@/utils/type"
import { Trash2Icon, SquareArrowOutUpRightIcon, CopyIcon, DownloadIcon } from "lucide-vue-next"

const { image } = defineProps<{
  image: Image
}>()

const { push } = useToaster()

const copyToClipboard = () => {
  navigator.clipboard.writeText(image.url)
  push({ title: "Copied to clipboard!", colour: "info", delay: 4000 })
}

const downloadFile = () => {
  const url = new URL(image.url)
  url.searchParams.set("download", "true")
  window.location.href = url.toString()
}

defineEmits<{
  delete: [id: string]
}>()
</script>

<template>
  <div
    class="bg-surface border-border text-foreground flex w-max items-center justify-center rounded-md border p-1 shadow-sm *:cursor-pointer *:rounded-md *:p-1 *:hover:bg-black/10 dark:*:hover:bg-white/10">
    <button @click="copyToClipboard"><CopyIcon :size="20" /></button>
    <a :href="image.url" target="_blank">
      <SquareArrowOutUpRightIcon :size="20" />
    </a>
    <button @click="downloadFile"><DownloadIcon :size="20" /></button>
    <button
      @click="$emit('delete', image.id)">
      <Trash2Icon class="stroke-danger" />
    </button>
  </div>
</template>
