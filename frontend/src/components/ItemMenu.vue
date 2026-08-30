<script setup lang="ts">
import useToaster from "@/composables/useToaster"
import type { Image } from "@/utils/type"
import {
  CheckSquareIcon,
  CopyIcon,
  DownloadIcon,
  MoveIcon,
  SquareArrowOutUpRightIcon,
  Trash2Icon
} from "lucide-vue-next"
import { computed, ref } from "vue"

const { image } = defineProps<{
  image: Image
}>()

const emit = defineEmits<{
  select: []
  move: []
  delete: []
}>()

const { push } = useToaster()

const open = ref(false)
const rawPosition = ref({ x: 0, y: 0 })

const MENU_WIDTH = 192
const MENU_HEIGHT = 220

const position = computed(() => ({
  x: Math.min(rawPosition.value.x, window.innerWidth - MENU_WIDTH - 8),
  y: Math.min(rawPosition.value.y, window.innerHeight - MENU_HEIGHT - 8)
}))

const openAt = (x: number, y: number) => {
  rawPosition.value = { x, y }
  open.value = true
}

const openAtElement = (el: HTMLElement) => {
  const rect = el.getBoundingClientRect()
  openAt(rect.right - MENU_WIDTH, rect.bottom + 4)
}

const close = () => {
  open.value = false
}

defineExpose({ openAt, openAtElement })

const copyToClipboard = () => {
  navigator.clipboard.writeText(image.url)
  push({ title: "Copied to clipboard!", colour: "info", delay: 4000 })
  close()
}

const downloadFile = () => {
  const url = new URL(image.url)
  url.searchParams.set("download", "true")
  window.location.href = url.toString()
  close()
}
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-[998]" @click="close" @contextmenu.prevent="close" />

    <div
      v-if="open"
      class="border-border bg-surface fixed z-[999] w-48 overflow-hidden rounded-lg border py-1 text-sm shadow-lg"
      :style="{ left: position.x + 'px', top: position.y + 'px' }">
      <a
        :href="image.url"
        target="_blank"
        @click="close"
        class="hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left hover:cursor-pointer">
        <SquareArrowOutUpRightIcon :size="14" /> Open
      </a>
      <button
        @click="copyToClipboard"
        class="hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left hover:cursor-pointer">
        <CopyIcon :size="14" /> Copy link
      </button>
      <button
        @click="downloadFile"
        class="hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left hover:cursor-pointer">
        <DownloadIcon :size="14" /> Download
      </button>

      <div class="border-border my-1 border-t" />

      <button
        @click="
          () => {
            emit('select')
            close()
          }
        "
        class="hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left hover:cursor-pointer">
        <CheckSquareIcon :size="14" /> Select
      </button>
      <button
        @click="
          () => {
            emit('move')
            close()
          }
        "
        class="hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left hover:cursor-pointer">
        <MoveIcon :size="14" /> Move to folder
      </button>

      <div class="border-border my-1 border-t" />

      <button
        @click="
          () => {
            emit('delete')
            close()
          }
        "
        class="text-danger hover:bg-surface-2 flex w-full items-center gap-2 px-3 py-2 text-left hover:cursor-pointer">
        <Trash2Icon :size="14" /> Delete
      </button>
    </div>
  </Teleport>
</template>
