<script setup lang="ts">
import { LucideUpload, LucideLoader2 } from "lucide-vue-next"

const { uploading = false, progress, compact = false } = defineProps<{
  uploading?: boolean
  progress?: { done: number; total: number }
  compact?: boolean
}>()

const emit = defineEmits<{
  selectedFiles: [files: File[]]
}>()

const handleFileSelect = (ev: Event) => {
  const target = ev.target as HTMLInputElement

  const files = target?.files
  if (files && files.length > 0) {
    emit("selectedFiles", Array.from(files))
  }
  target.value = ""
}

const handleDrop = (ev: DragEvent) => {
  if (uploading) return

  const files = ev.dataTransfer?.files
  if (files && files.length > 0) {
    emit("selectedFiles", Array.from(files))
  }
}
</script>

<template>
  <label
    for="dropzone-file"
    class="border-border bg-surface overflow-hidden rounded-xl border-2 border-dashed text-center transition-colors"
    :class="[
      uploading ? 'border-primary/50 pointer-events-none' : 'hover:border-primary/40 hover:bg-surface-2 cursor-pointer',
      compact ? 'flex w-full items-center justify-center gap-2 px-4 py-3' : 'flex aspect-square w-full flex-col items-center justify-center p-2'
    ]"
    @dragover.prevent
    @drop.prevent="handleDrop">
    <!-- Uploading -->
    <template v-if="uploading">
      <LucideLoader2 :class="['text-primary animate-spin', compact ? 'size-5' : 'mb-2 size-8']" />
      <p class="text-primary text-xs font-semibold sm:text-sm">
        {{ progress && progress.total > 1 ? `Uploading ${progress.done}/${progress.total}...` : "Uploading..." }}
      </p>
    </template>

    <!-- Idle -->
    <template v-else-if="compact">
      <LucideUpload class="text-muted size-5" />
      <p class="text-muted text-xs sm:text-sm">
        <span class="text-primary font-semibold">Click to upload</span>
        or drag and drop
      </p>
    </template>
    <template v-else>
      <LucideUpload class="text-muted mb-2 size-7 sm:size-8" />
      <p class="text-muted text-xs leading-tight sm:text-sm">
        <span class="text-primary font-semibold">Click to upload</span>
        or drag and drop
      </p>
      <p class="text-muted mt-1 hidden text-xs opacity-70 sm:block">Multiple files supported</p>
    </template>

    <input @change="handleFileSelect" id="dropzone-file" type="file" multiple class="hidden" :disabled="uploading" />
  </label>
</template>
