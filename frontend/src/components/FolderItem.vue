<script setup lang="ts">
import type { Folder } from "@/utils/type"
import { CheckIcon, FolderIcon, MoreVerticalIcon, MoveIcon, CheckSquareIcon, PencilIcon, Trash2Icon } from "lucide-vue-next"
import { computed, onBeforeUnmount, ref, useTemplateRef, watch } from "vue"

const { folder, mode, selectable = false, selected = false } = defineProps<{
  folder: Folder
  mode: "grid" | "list"
  selectable?: boolean
  selected?: boolean
}>()

const emit = defineEmits<{
  open: []
  rename: []
  delete: []
  select: [id: string]
  move: [id: string]
  "drop-image": [id: string]
}>()

const onTileClick = () => {
  if (!selectable) emit("open")
}

const menuOpen = ref(false)
const menuButtonRef = useTemplateRef<HTMLElement>("menuButtonRef")
const menuPanelRef = useTemplateRef<HTMLElement>("menuPanelRef")
const menuPosition = ref({ x: 0, y: 0 })

const MENU_WIDTH = 144

const openMenuAt = (x: number, y: number) => {
  menuPosition.value = {
    x: Math.min(x, window.innerWidth - MENU_WIDTH - 8),
    y: Math.min(y, window.innerHeight - 8)
  }
  menuOpen.value = true
}

const openMenuFromButton = (e: MouseEvent) => {
  if (menuOpen.value) {
    menuOpen.value = false
    return
  }

  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  openMenuAt(rect.right - MENU_WIDTH, rect.bottom + 4)
}

const onContextMenu = (e: MouseEvent) => {
  if (selectable) return
  e.preventDefault()
  openMenuAt(e.clientX, e.clientY)
}

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
  const target = e.target as Node
  if (menuButtonRef.value?.contains(target)) return
  if (menuPanelRef.value && !menuPanelRef.value.contains(target)) menuOpen.value = false
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
    @click="onTileClick"
    @contextmenu="onContextMenu"
    @dragover="onDragOver"
    @dragenter="onDragEnter"
    @dragleave="onDragLeave"
    @drop.prevent="onDrop"
    :class="[
      'bg-surface-2 group hover:bg-surface-2/70 relative flex aspect-square w-full cursor-pointer flex-col items-center justify-center gap-2 rounded-xl p-3 shadow-sm transition',
      isDragOver && 'ring-primary bg-primary/10 ring-2 ring-offset-2',
      selected && 'ring-primary ring-3 ring-offset-2'
    ]">
    <FolderIcon :size="40" class="text-primary shrink-0" :stroke-width="1.5" />
    <span class="text-foreground w-full truncate text-center text-sm font-medium" :title="folder.name">
      {{ folder.name }}
    </span>

    <div
      v-if="selectable"
      :class="[
        'absolute top-2 left-2 z-10 grid size-6 place-items-center rounded-full border-2 shadow-sm transition',
        selected ? 'bg-primary border-primary text-white' : 'border-white/80 bg-black/30'
      ]">
      <CheckIcon v-if="selected" :size="15" :stroke-width="3" />
    </div>

    <button
      v-if="!selectable"
      ref="menuButtonRef"
      @click.stop="openMenuFromButton"
      class="text-muted absolute top-1.5 right-1.5 grid size-7 place-items-center rounded-full opacity-0 transition hover:cursor-pointer hover:bg-black/10 group-hover:opacity-100 dark:hover:bg-white/10">
      <MoreVerticalIcon :size="16" />
    </button>
  </div>

  <div
    v-else
    @click="onTileClick"
    @contextmenu="onContextMenu"
    @dragover="onDragOver"
    @dragenter="onDragEnter"
    @dragleave="onDragLeave"
    @drop.prevent="onDrop"
    :class="[
      'bg-surface hover:bg-surface-2 group relative flex w-full cursor-pointer items-center gap-3 border-b border-border px-2 py-2 transition last:border-b-0',
      isDragOver && 'ring-primary bg-primary/10 ring-2 ring-inset',
      selected && 'bg-primary/10'
    ]">
    <div
      v-if="selectable"
      :class="[
        'grid size-6 shrink-0 place-items-center rounded-full border-2 shadow-sm transition',
        selected ? 'bg-primary border-primary text-white' : 'border-border bg-surface-2'
      ]">
      <CheckIcon v-if="selected" :size="15" :stroke-width="3" />
    </div>

    <div class="bg-surface-2 grid size-10 shrink-0 place-items-center rounded-lg">
      <FolderIcon :size="20" class="text-primary" />
    </div>

    <div class="min-w-0 flex-1">
      <span class="text-foreground block truncate text-sm font-medium group-hover:underline" :title="folder.name">
        {{ folder.name }}
      </span>
    </div>

    <button
      v-if="!selectable"
      ref="menuButtonRef"
      @click.stop="openMenuFromButton"
      class="text-muted grid size-8 shrink-0 place-items-center rounded-full opacity-0 transition hover:cursor-pointer hover:bg-black/10 group-hover:opacity-100 dark:hover:bg-white/10">
      <MoreVerticalIcon :size="16" />
    </button>
  </div>

  <Teleport to="body">
    <div
      v-if="menuOpen"
      ref="menuPanelRef"
      class="border-border bg-surface fixed z-[999] w-36 overflow-hidden rounded-lg border py-1 shadow-lg"
      :style="{ left: menuPosition.x + 'px', top: menuPosition.y + 'px' }">
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
            emit('select', folder.id)
          }
        "
        class="hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:cursor-pointer">
        <CheckSquareIcon :size="14" /> Select
      </button>
      <button
        @click.stop="
          () => {
            menuOpen = false
            emit('move', folder.id)
          }
        "
        class="hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:cursor-pointer">
        <MoveIcon :size="14" /> Move to folder
      </button>

      <div class="border-border my-1 border-t" />

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
  </Teleport>
</template>
