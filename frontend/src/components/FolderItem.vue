<script setup lang="ts">
import type { Folder } from "@/utils/type"
import { FolderIcon, MoreVerticalIcon, PencilIcon, Trash2Icon } from "lucide-vue-next"
import { computed, onBeforeUnmount, ref, useTemplateRef, watch } from "vue"

const { folder, mode } = defineProps<{
  folder: Folder
  mode: "grid" | "list"
}>()

const emit = defineEmits<{
  open: []
  rename: []
  delete: []
  "drop-image": [id: string]
}>()

const menuOpen = ref(false)
const menuRef = useTemplateRef<HTMLElement>("menuRef")

const dragOverDepth = ref(0)
const isDragOver = computed(() => dragOverDepth.value > 0)

const onDragOver = (e: DragEvent) => {
  if (!e.dataTransfer?.types.includes("text/plain")) return
  e.preventDefault()
  e.dataTransfer.dropEffect = "move"
}

const onDragEnter = (e: DragEvent) => {
  if (!e.dataTransfer?.types.includes("text/plain")) return
  dragOverDepth.value++
}

const onDragLeave = () => {
  if (dragOverDepth.value > 0) dragOverDepth.value--
}

const onDrop = (e: DragEvent) => {
  dragOverDepth.value = 0
  const id = e.dataTransfer?.getData("text/plain")
  if (id) emit("drop-image", id)
}

const onDocumentClick = (e: MouseEvent) => {
  if (menuRef.value && !menuRef.value.contains(e.target as Node)) menuOpen.value = false
}

watch(menuOpen, open => {
  if (open) window.addEventListener("click", onDocumentClick)
  else window.removeEventListener("click", onDocumentClick)
})

onBeforeUnmount(() => window.removeEventListener("click", onDocumentClick))
</script>

<template>
  <div
    v-if="mode === 'grid'"
    @click="emit('open')"
    @dragover="onDragOver"
    @dragenter="onDragEnter"
    @dragleave="onDragLeave"
    @drop.prevent="onDrop"
    :class="[
      'bg-surface-2 group hover:bg-surface-2/70 relative flex aspect-square w-full cursor-pointer flex-col items-center justify-center gap-2 rounded-xl p-3 shadow-sm transition',
      isDragOver && 'ring-primary bg-primary/10 ring-2 ring-offset-2'
    ]">
    <FolderIcon :size="40" class="text-primary shrink-0" :stroke-width="1.5" />
    <span class="text-foreground w-full truncate text-center text-sm font-medium" :title="folder.name">
      {{ folder.name }}
    </span>

    <div ref="menuRef" class="absolute top-1.5 right-1.5">
      <button
        @click.stop="menuOpen = !menuOpen"
        class="text-muted grid size-7 place-items-center rounded-full opacity-0 transition hover:cursor-pointer hover:bg-black/10 group-hover:opacity-100 dark:hover:bg-white/10">
        <MoreVerticalIcon :size="16" />
      </button>
      <div
        v-if="menuOpen"
        class="border-border bg-surface absolute top-8 right-0 z-10 w-36 overflow-hidden rounded-lg border py-1 shadow-lg">
        <button
          @click.stop="
            () => {
              menuOpen = false
              emit('rename')
            }
          "
          class="hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:cursor-pointer">
          <PencilIcon :size="14" /> Rename
        </button>
        <button
          @click.stop="
            () => {
              menuOpen = false
              emit('delete')
            }
          "
          class="text-danger hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:cursor-pointer">
          <Trash2Icon :size="14" /> Delete
        </button>
      </div>
    </div>
  </div>

  <div
    v-else
    @click="emit('open')"
    @dragover="onDragOver"
    @dragenter="onDragEnter"
    @dragleave="onDragLeave"
    @drop.prevent="onDrop"
    :class="[
      'bg-surface hover:bg-surface-2 group relative flex w-full cursor-pointer items-center gap-3 border-b border-border px-2 py-2 transition last:border-b-0',
      isDragOver && 'bg-primary/10'
    ]">
    <div class="bg-surface-2 grid size-10 shrink-0 place-items-center rounded-lg">
      <FolderIcon :size="20" class="text-primary" />
    </div>

    <div class="min-w-0 flex-1">
      <span class="text-foreground block truncate text-sm font-medium group-hover:underline" :title="folder.name">
        {{ folder.name }}
      </span>
    </div>

    <div ref="menuRef" class="relative shrink-0">
      <button
        @click.stop="menuOpen = !menuOpen"
        class="text-muted grid size-8 place-items-center rounded-full opacity-0 transition hover:cursor-pointer hover:bg-black/10 group-hover:opacity-100 dark:hover:bg-white/10">
        <MoreVerticalIcon :size="16" />
      </button>
      <div
        v-if="menuOpen"
        class="border-border bg-surface absolute top-9 right-0 z-10 w-36 overflow-hidden rounded-lg border py-1 shadow-lg">
        <button
          @click.stop="
            () => {
              menuOpen = false
              emit('rename')
            }
          "
          class="hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:cursor-pointer">
          <PencilIcon :size="14" /> Rename
        </button>
        <button
          @click.stop="
            () => {
              menuOpen = false
              emit('delete')
            }
          "
          class="text-danger hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:cursor-pointer">
          <Trash2Icon :size="14" /> Delete
        </button>
      </div>
    </div>
  </div>
</template>
