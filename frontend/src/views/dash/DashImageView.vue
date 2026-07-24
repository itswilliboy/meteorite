<script setup lang="ts">
import Card from "@/components/Card.vue"
import { Button } from "@/components/common"
import ImageCard from "@/components/ImageCard.vue"
import PageContainer from "@/components/PageContainer.vue"
import FileDrop from "@/components/FileDrop.vue"
import ConfirmDialogue from "@/components/ConfirmDialogue.vue"
import useClient from "@/composables/useClient"
import useToaster from "@/composables/useToaster"
import { bumpMediaVersion } from "@/composables/useMediaVersion"
import { useInfiniteScroll } from "@/composables/useInfiniteScroll"
import { formatBytes } from "@/utils/format"
import type { Image } from "@/utils/type"
import { computed, nextTick, onBeforeUnmount, ref, useTemplateRef } from "vue"
import {
  ListChecksIcon,
  LucideLoader2,
  Trash2Icon,
  XIcon,
  Grid2x2Icon,
  Grid3x3Icon,
  LayoutGridIcon,
  CropIcon,
  ExpandIcon
} from "lucide-vue-next"

defineOptions({ name: "DashImageView" })

const client = useClient()
const { push } = useToaster()

const images = ref<Image[]>([])
const page = ref(-1)
const hasNext = ref(true)
const loadingMore = ref(false)

const uploading = ref(false)
const uploadProgress = ref({ done: 0, total: 0 })

const selectMode = ref(false)
const selected = ref<Set<string>>(new Set())
const selectedSizes = ref<Map<string, number>>(new Map())
const bulkConfirmOpen = ref(false)

const allSelectedLoaded = computed(() => images.value.length > 0 && images.value.every(img => selected.value.has(img.id)))

const selectedSize = computed(() => [...selectedSizes.value.values()].reduce((total, size) => total + size, 0))

const sortOptions = [
  { value: "date_desc", label: "Newest first" },
  { value: "date_asc", label: "Oldest first" },
  { value: "size_desc", label: "Largest first" },
  { value: "size_asc", label: "Smallest first" },
  { value: "views_desc", label: "Most viewed" },
  { value: "views_asc", label: "Least viewed" },
  { value: "name_asc", label: "Name (A–Z)" },
  { value: "name_desc", label: "Name (Z–A)" }
] as const

const sort = ref<(typeof sortOptions)[number]["value"]>("date_desc")

type GridSize = "small" | "medium" | "large"
const gridSizeOptions: { value: GridSize; icon: typeof Grid2x2Icon; title: string }[] = [
  { value: "large", icon: Grid2x2Icon, title: "Large previews" },
  { value: "medium", icon: LayoutGridIcon, title: "Medium previews" },
  { value: "small", icon: Grid3x3Icon, title: "Compact grid" }
]
const gridClasses: Record<GridSize, string> = {
  large: "grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5",
  medium: "grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8",
  small: "grid-cols-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8 xl:grid-cols-10 2xl:grid-cols-12"
}

const gridSize = ref<GridSize>((localStorage.getItem("gallery-grid-size") as GridSize) || "medium")
const setGridSize = (value: GridSize) => {
  gridSize.value = value
  localStorage.setItem("gallery-grid-size", value)
}

type FitMode = "cover" | "contain"
const fitModeOptions: { value: FitMode; icon: typeof CropIcon; title: string }[] = [
  { value: "cover", icon: CropIcon, title: "Fill (crop to square)" },
  { value: "contain", icon: ExpandIcon, title: "Fit (show whole image)" }
]
const fitMode = ref<FitMode>((localStorage.getItem("gallery-fit-mode") as FitMode) || "cover")
const setFitMode = (value: FitMode) => {
  fitMode.value = value
  localStorage.setItem("gallery-fit-mode", value)
}

const baseThumbnailWidth: Record<GridSize, number> = { small: 240, medium: 400, large: 600 }
const thumbnailWidth = computed(() =>
  Math.round(baseThumbnailWidth[gridSize.value] * Math.min(window.devicePixelRatio || 1, 2))
)

const loadMore = async (): Promise<boolean> => {
  if (loadingMore.value || !hasNext.value) return false

  loadingMore.value = true
  try {
    const resp = await client.getImages(page.value + 1, sort.value)
    images.value.push(...resp.data)
    page.value = resp.page
    hasNext.value = resp.hasNext
    return true
  } finally {
    loadingMore.value = false
  }
}

const sentinel = useTemplateRef<HTMLDivElement>("sentinel")
const { refresh } = useInfiniteScroll(sentinel, loadMore)

const resetAndReload = async () => {
  images.value = []
  page.value = -1
  hasNext.value = true
  await nextTick()
  refresh()
}

const onSortChange = () => {
  resetAndReload()
}

const resetDragState = () => {
  window.removeEventListener("pointermove", onGridPointerMove)
  window.removeEventListener("pointerup", onGridPointerUp)
  if (rafId !== null) {
    cancelAnimationFrame(rafId)
    rafId = null
  }
  dragStart.value = null
  dragCurrent.value = null
  isDragging.value = false
  dragPendingId.value = null
}

const toggleSelectMode = () => {
  selectMode.value = !selectMode.value
  selected.value.clear()
  selectedSizes.value.clear()
  lastSelectedIndex.value = null
  resetDragState()
}

const applySelection = (id: string, select: boolean) => {
  if (select) {
    selected.value.add(id)
    const image = images.value.find(img => img.id === id)
    if (image) selectedSizes.value.set(id, image.size)
  } else {
    selected.value.delete(id)
    selectedSizes.value.delete(id)
  }
}

const lastSelectedIndex = ref<number | null>(null)

const toggleSingle = (id: string) => {
  const idx = images.value.findIndex(img => img.id === id)
  applySelection(id, !selected.value.has(id))
  if (idx !== -1) lastSelectedIndex.value = idx
}

const selectRange = (fromIdx: number, toIdx: number) => {
  const [start, end] = [fromIdx, toIdx].sort((a, b) => a - b)
  for (let i = start; i <= end; i++) applySelection(images.value[i].id, true)
}

const cardRefs = new Map<string, HTMLElement>()
const setTileRef = (id: string, el: Element | null) => {
  if (el instanceof HTMLElement) cardRefs.set(id, el)
  else cardRefs.delete(id)
}

const DRAG_THRESHOLD = 4

const dragStart = ref<{ x: number; y: number } | null>(null)
const dragCurrent = ref<{ x: number; y: number } | null>(null)
const isDragging = ref(false)
const dragPendingId = ref<string | null>(null)
const dragAdditive = ref(false)
const dragSnapshot = ref<Set<string>>(new Set())
let rafId: number | null = null

const marqueeRect = computed(() => {
  if (!isDragging.value || !dragStart.value || !dragCurrent.value) return null
  const x1 = Math.min(dragStart.value.x, dragCurrent.value.x)
  const y1 = Math.min(dragStart.value.y, dragCurrent.value.y)
  const x2 = Math.max(dragStart.value.x, dragCurrent.value.x)
  const y2 = Math.max(dragStart.value.y, dragCurrent.value.y)
  return { left: x1, top: y1, width: x2 - x1, height: y2 - y1, x1, y1, x2, y2 }
})

const updateMarqueeSelection = () => {
  const rect = marqueeRect.value
  if (!rect) return

  const intersecting = new Set<string>()
  for (const [id, el] of cardRefs) {
    const r = el.getBoundingClientRect()
    if (r.left < rect.x2 && r.right > rect.x1 && r.top < rect.y2 && r.bottom > rect.y1) {
      intersecting.add(id)
    }
  }

  const next = dragAdditive.value ? new Set([...dragSnapshot.value, ...intersecting]) : intersecting
  selected.value = next
  selectedSizes.value = new Map([...next].map(id => [id, images.value.find(img => img.id === id)?.size ?? 0]))
}

const onGridPointerMove = (e: PointerEvent) => {
  if (!dragStart.value) return
  dragCurrent.value = { x: e.clientX, y: e.clientY }

  if (!isDragging.value) {
    const dx = e.clientX - dragStart.value.x
    const dy = e.clientY - dragStart.value.y
    if (Math.hypot(dx, dy) < DRAG_THRESHOLD) return
    isDragging.value = true
  }

  if (rafId !== null) return
  rafId = requestAnimationFrame(() => {
    rafId = null
    updateMarqueeSelection()
  })
}

const onGridPointerUp = () => {
  if (isDragging.value) {
    updateMarqueeSelection()
  } else if (dragPendingId.value) {
    toggleSingle(dragPendingId.value)
  } else {
    selected.value.clear()
    selectedSizes.value.clear()
  }

  resetDragState()
}

const onGridPointerDown = (e: PointerEvent) => {
  if (!selectMode.value || e.button !== 0) return

  const id = (e.target as HTMLElement).closest<HTMLElement>("[data-image-id]")?.dataset.imageId ?? null

  if (e.shiftKey && id && lastSelectedIndex.value !== null) {
    const idx = images.value.findIndex(img => img.id === id)
    if (idx !== -1) {
      selectRange(lastSelectedIndex.value, idx)
      lastSelectedIndex.value = idx
    }
    return
  }

  e.preventDefault()
  dragStart.value = { x: e.clientX, y: e.clientY }
  dragCurrent.value = { x: e.clientX, y: e.clientY }
  dragPendingId.value = id
  dragAdditive.value = e.shiftKey
  dragSnapshot.value = new Set(selected.value)
  isDragging.value = false

  window.addEventListener("pointermove", onGridPointerMove)
  window.addEventListener("pointerup", onGridPointerUp)
}

onBeforeUnmount(resetDragState)

const toggleSelectAllLoaded = () => {
  if (allSelectedLoaded.value) {
    images.value.forEach(img => {
      selected.value.delete(img.id)
      selectedSizes.value.delete(img.id)
    })
  } else {
    images.value.forEach(img => {
      selected.value.add(img.id)
      selectedSizes.value.set(img.id, img.size)
    })
  }
}

const bulkDelete = async () => {
  const ids = [...selected.value]
  await Promise.all(ids.map(id => client.deleteImage(id)))

  images.value = images.value.filter(img => !selected.value.has(img.id))

  push({ title: `Deleted ${ids.length} ${ids.length === 1 ? "item" : "items"}`, delay: 4000, colour: "info" })
  selected.value.clear()
  selectedSizes.value.clear()
  selectMode.value = false
  bumpMediaVersion()
}

const uploadFiles = async (files: File[]) => {
  if (uploading.value || files.length === 0) return

  uploading.value = true
  uploadProgress.value = { done: 0, total: files.length }

  let succeeded = 0
  const failed: string[] = []

  for (const file of files) {
    try {
      await client.uploadImage(file)
      succeeded++
    } catch {
      failed.push(file.name)
    } finally {
      uploadProgress.value = { ...uploadProgress.value, done: uploadProgress.value.done + 1 }
    }
  }

  if (succeeded > 0) {
    push({ title: `Uploaded ${succeeded} ${succeeded === 1 ? "file" : "files"}`, colour: "success", delay: 4000 })
  }
  if (failed.length > 0) {
    push({ title: `Failed to upload: ${failed.join(", ")}`, colour: "danger", delay: 6000 })
  }

  await resetAndReload()
  uploading.value = false
  if (succeeded > 0) bumpMediaVersion()
}
</script>

<template>
  <PageContainer title="Gallery">
    <ConfirmDialogue
      v-if="bulkConfirmOpen"
      @dismiss="bulkConfirmOpen = false"
      title="Delete selected?"
      :description="`This will permanently delete ${selected.size} ${selected.size === 1 ? 'item' : 'items'} (${formatBytes(selectedSize)}).`"
      confirm-text="Delete"
      confirm-colour="danger"
      :confirm-icon="Trash2Icon"
      :confirm-action="() => bulkDelete()" />

    <div class="mb-4 flex flex-wrap items-center justify-end gap-2">
      <select
        v-model="sort"
        @change="onSortChange"
        class="border-border bg-surface text-muted hover:text-foreground rounded-lg border px-3 py-2 text-sm font-medium transition hover:cursor-pointer">
        <option v-for="opt in sortOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
      </select>

      <div class="border-border bg-surface flex items-center gap-0.5 rounded-lg border p-0.5">
        <button
          v-for="opt in gridSizeOptions"
          :key="opt.value"
          @click="setGridSize(opt.value)"
          :title="opt.title"
          :class="[
            'flex items-center rounded-md p-1.5 transition hover:cursor-pointer',
            gridSize === opt.value ? 'bg-primary text-white' : 'text-muted hover:bg-surface-2 hover:text-foreground'
          ]">
          <component :is="opt.icon" :size="16" />
        </button>
      </div>

      <div class="border-border bg-surface flex items-center gap-0.5 rounded-lg border p-0.5">
        <button
          v-for="opt in fitModeOptions"
          :key="opt.value"
          @click="setFitMode(opt.value)"
          :title="opt.title"
          :class="[
            'flex items-center rounded-md p-1.5 transition hover:cursor-pointer',
            fitMode === opt.value ? 'bg-primary text-white' : 'text-muted hover:bg-surface-2 hover:text-foreground'
          ]">
          <component :is="opt.icon" :size="16" />
        </button>
      </div>

      <button
        @click="toggleSelectMode"
        :class="[
          'flex items-center gap-1.5 rounded-lg border px-3 py-2 text-sm font-medium transition hover:cursor-pointer',
          selectMode
            ? 'bg-primary border-primary text-white'
            : 'border-border bg-surface text-muted hover:bg-surface-2 hover:text-foreground'
        ]">
        <XIcon v-if="selectMode" :size="16" />
        <ListChecksIcon v-else :size="16" />
        {{ selectMode ? "Cancel" : "Select" }}
      </button>
    </div>

    <Transition>
      <div
        v-if="selectMode"
        class="bg-surface border-border mb-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border p-3">
        <div class="flex items-center gap-3">
          <label class="flex items-center gap-2 text-sm font-medium hover:cursor-pointer">
            <input
              type="checkbox"
              :checked="allSelectedLoaded"
              @change="toggleSelectAllLoaded"
              class="accent-primary size-4" />
            Select all loaded
          </label>
          <span class="text-muted text-sm">
            {{ selected.size }} selected
            <template v-if="selected.size > 0">&middot; {{ formatBytes(selectedSize) }}</template>
          </span>
        </div>

        <Button variant="danger" :icon="Trash2Icon" :disabled="selected.size === 0" @click="bulkConfirmOpen = true">
          Delete selected
        </Button>
      </div>
    </Transition>

    <Card :class="['grid gap-3', gridClasses[gridSize]]" @pointerdown="onGridPointerDown">
      <FileDrop v-if="!selectMode" :uploading="uploading" :progress="uploadProgress" @selected-files="uploadFiles" />

      <div
        v-for="image in images"
        :key="image.id"
        :data-image-id="image.id"
        :ref="el => setTileRef(image.id, el as Element | null)">
        <ImageCard
          :image="image"
          :selectable="selectMode"
          :selected="selected.has(image.id)"
          :fit="fitMode"
          :thumbnail-width="thumbnailWidth"
          @pop="
            id => {
              images.splice(
                images.findIndex(img => img.id === id),
                1
              )
              bumpMediaVersion()
            }
          " />
      </div>
    </Card>

    <div ref="sentinel" class="flex h-10 items-center justify-center">
      <LucideLoader2 v-if="loadingMore" :size="20" class="text-muted animate-spin" />
    </div>

    <div
      v-if="marqueeRect"
      class="border-primary bg-primary/10 pointer-events-none fixed z-50 rounded-sm border"
      :style="{
        left: marqueeRect.left + 'px',
        top: marqueeRect.top + 'px',
        width: marqueeRect.width + 'px',
        height: marqueeRect.height + 'px'
      }" />
  </PageContainer>
</template>

<style scoped>
.v-enter-active,
.v-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
