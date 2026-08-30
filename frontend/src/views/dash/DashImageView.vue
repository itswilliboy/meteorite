<script setup lang="ts">
import Card from "@/components/Card.vue"
import { Button } from "@/components/common"
import ImageCard from "@/components/ImageCard.vue"
import ImageListRow from "@/components/ImageListRow.vue"
import FolderItem from "@/components/FolderItem.vue"
import FolderNameDialog from "@/components/FolderNameDialog.vue"
import MoveToFolderDialog from "@/components/MoveToFolderDialog.vue"
import PageContainer from "@/components/PageContainer.vue"
import FileDrop from "@/components/FileDrop.vue"
import ConfirmDialogue from "@/components/ConfirmDialogue.vue"
import useClient from "@/composables/useClient"
import { HTTPException } from "@/utils/client"
import useToaster from "@/composables/useToaster"
import { bumpMediaVersion } from "@/composables/useMediaVersion"
import { useInfiniteScroll } from "@/composables/useInfiniteScroll"
import { formatBytes } from "@/utils/format"
import type { BreadcrumbEntry, Folder, Image } from "@/utils/type"
import { computed, nextTick, onBeforeUnmount, ref, useTemplateRef, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import {
  ListChecksIcon,
  LucideLoader2,
  Trash2Icon,
  XIcon,
  Grid2x2Icon,
  Grid3x3Icon,
  LayoutGridIcon,
  LayoutListIcon,
  CropIcon,
  ExpandIcon,
  FolderPlusIcon,
  HomeIcon,
  ChevronRightIcon,
  MoveIcon
} from "lucide-vue-next"

defineOptions({ name: "DashImageView" })

const client = useClient()
const { push } = useToaster()
const route = useRoute()
const router = useRouter()

const currentFolderId = computed<string | null>(() => (typeof route.query.folder === "string" ? route.query.folder : null))

const folders = ref<Folder[]>([])
const breadcrumb = ref<BreadcrumbEntry[]>([])

const loadFolders = async () => {
  const requestedFolderId = currentFolderId.value
  const resp = await client.getFolders(requestedFolderId)
  if (currentFolderId.value !== requestedFolderId) return
  folders.value = resp.folders
  breadcrumb.value = resp.breadcrumb
}

const openFolder = (id: string | null) => {
  router.push({ query: { ...route.query, folder: id ?? undefined } })
}

const newFolderDialogOpen = ref(false)
const renameFolderTarget = ref<Folder | null>(null)
const deleteFolderTarget = ref<Folder | null>(null)
const moveDialogOpen = ref(false)

const createFolder = async (name: string) => {
  try {
    await client.createFolder(name, currentFolderId.value)
    push({ title: `Created folder "${name}"`, colour: "success", delay: 4000 })
    await loadFolders()
  } catch (e) {
    push({ title: e instanceof HTTPException ? e.message : "Could not create folder", colour: "danger", delay: 6000 })
  }
}

const renameFolder = async (name: string) => {
  if (!renameFolderTarget.value) return
  try {
    await client.renameFolder(renameFolderTarget.value.id, name)
    await loadFolders()
  } catch (e) {
    push({ title: e instanceof HTTPException ? e.message : "Could not rename folder", colour: "danger", delay: 6000 })
  }
}

const deleteFolder = async () => {
  if (!deleteFolderTarget.value) return
  const name = deleteFolderTarget.value.name
  try {
    await client.deleteFolder(deleteFolderTarget.value.id)
    push({ title: `Deleted folder "${name}"`, colour: "info", delay: 4000 })
    await loadFolders()
    bumpMediaVersion()
  } catch (e) {
    push({ title: e instanceof HTTPException ? e.message : "Could not delete folder", colour: "danger", delay: 6000 })
  }
}

const bulkMove = async (folderId: string | null) => {
  const ids = [...selected.value]
  const folderIds = ids.filter(id => folders.value.some(f => f.id === id))
  const imageIds = ids.filter(id => !folderIds.includes(id))

  let movedImageIds: string[] = []
  let imagesFailed = false
  if (imageIds.length > 0) {
    try {
      await client.bulkMoveImages(imageIds, folderId)
      movedImageIds = imageIds
    } catch (e) {
      imagesFailed = true
      push({ title: e instanceof HTTPException ? e.message : "Could not move selected files", colour: "danger", delay: 6000 })
    }
  }

  let movedFolderIds: string[] = []
  if (folderIds.length > 0) {
    const results = await Promise.allSettled(folderIds.map(id => client.moveFolder(id, folderId)))
    movedFolderIds = folderIds.filter((_, i) => results[i].status === "fulfilled")
    const failedCount = folderIds.length - movedFolderIds.length
    if (failedCount > 0) {
      push({ title: `Could not move ${failedCount} ${failedCount === 1 ? "folder" : "folders"}`, colour: "danger", delay: 6000 })
    }
  }

  if (folderId !== currentFolderId.value) {
    images.value = images.value.filter(img => !movedImageIds.includes(img.id))
    folders.value = folders.value.filter(f => !movedFolderIds.includes(f.id))
  }

  const movedCount = movedImageIds.length + movedFolderIds.length
  if (movedCount > 0) {
    push({ title: `Moved ${movedCount} ${movedCount === 1 ? "item" : "items"}`, delay: 4000, colour: "info" })
    bumpMediaVersion()
  }

  for (const id of [...movedImageIds, ...movedFolderIds]) {
    selected.value.delete(id)
    selectedSizes.value.delete(id)
  }
  if (!imagesFailed && movedFolderIds.length === folderIds.length) {
    selectMode.value = false
    moveDialogOpen.value = false
  }
}

const moveImageToFolder = async (id: string, folderId: string | null, folderName: string) => {
  if (folderId === currentFolderId.value) return

  try {
    await client.bulkMoveImages([id], folderId)
  } catch (e) {
    push({ title: e instanceof HTTPException ? e.message : "Could not move item", colour: "danger", delay: 6000 })
    return
  }

  images.value = images.value.filter(img => img.id !== id)
  selected.value.delete(id)
  selectedSizes.value.delete(id)
  push({ title: `Moved to "${folderName}"`, colour: "info", delay: 4000 })
  bumpMediaVersion()
}

const moveItemViaMenu = (id: string) => {
  const item = allItems.value.find(i => i.id === id)
  if (!item) return

  selected.value = new Set([id])
  selectedSizes.value = new Map([[id, item.size]])
  moveDialogOpen.value = true
}

const draggingImageId = ref<string | null>(null)

const onImageDragStart = (e: DragEvent, id: string) => {
  draggingImageId.value = id
  e.dataTransfer?.setData("text/plain", id)
  if (e.dataTransfer) e.dataTransfer.effectAllowed = "move"
}

const onImageDragEnd = () => {
  draggingImageId.value = null
}

const onBreadcrumbDrop = (e: DragEvent, folderId: string | null, folderName: string) => {
  breadcrumbDragOverDepth.value = {}
  const id = e.dataTransfer?.getData("text/plain")
  if (id) moveImageToFolder(id, folderId, folderName)
}

const ROOT_BREADCRUMB_KEY = "__root__"
const breadcrumbDragOverDepth = ref<Record<string, number>>({})

const isBreadcrumbDragOver = (key: string) => (breadcrumbDragOverDepth.value[key] ?? 0) > 0

const onBreadcrumbDragOver = (e: DragEvent) => {
  if (!e.dataTransfer?.types.includes("text/plain")) return
  e.preventDefault()
  e.dataTransfer.dropEffect = "move"
}

const onBreadcrumbDragEnter = (e: DragEvent, key: string) => {
  if (!e.dataTransfer?.types.includes("text/plain")) return
  breadcrumbDragOverDepth.value = { ...breadcrumbDragOverDepth.value, [key]: (breadcrumbDragOverDepth.value[key] ?? 0) + 1 }
}

const onBreadcrumbDragLeave = (key: string) => {
  const depth = breadcrumbDragOverDepth.value[key] ?? 0
  if (depth > 0) breadcrumbDragOverDepth.value = { ...breadcrumbDragOverDepth.value, [key]: depth - 1 }
}

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

const allItems = computed(() => [
  ...folders.value.map(f => ({ id: f.id, size: 0 })),
  ...images.value.map(img => ({ id: img.id, size: img.size }))
])
const selectedFolderIds = computed(() => [...selected.value].filter(id => folders.value.some(f => f.id === id)))

const allSelectedLoaded = computed(() => allItems.value.length > 0 && allItems.value.every(item => selected.value.has(item.id)))

const selectedSize = computed(() => [...selectedSizes.value.values()].reduce((total, size) => total + size, 0))

const selectedSizeLabel = computed(() => {
  const folderCount = selectedFolderIds.value.length
  const fileCount = selected.value.size - folderCount

  const parts: string[] = []
  if (fileCount > 0) parts.push(`${fileCount} ${fileCount === 1 ? "file" : "files"} (${formatBytes(selectedSize.value)})`)
  if (folderCount > 0) parts.push(`${folderCount} ${folderCount === 1 ? "folder" : "folders"}`)
  return parts.join(" + ")
})

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

type ViewMode = "grid" | "list"
const viewModeOptions: { value: ViewMode; icon: typeof LayoutGridIcon; title: string }[] = [
  { value: "grid", icon: LayoutGridIcon, title: "Grid view" },
  { value: "list", icon: LayoutListIcon, title: "List view" }
]
const viewMode = ref<ViewMode>((localStorage.getItem("gallery-view-mode") as ViewMode) || "grid")
const setViewMode = (value: ViewMode) => {
  viewMode.value = value
  localStorage.setItem("gallery-view-mode", value)
}

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

  const requestedFolderId = currentFolderId.value
  const requestedSort = sort.value
  const requestedPage = page.value + 1

  loadingMore.value = true
  try {
    const resp = await client.getImages(requestedPage, requestedSort, requestedFolderId)
    if (currentFolderId.value !== requestedFolderId || sort.value !== requestedSort) return false

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

watch(
  currentFolderId,
  () => {
    selectMode.value = false
    selected.value.clear()
    selectedSizes.value.clear()
    loadFolders()
    resetAndReload()
  },
  { immediate: true }
)

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
    const item = allItems.value.find(i => i.id === id)
    if (item) selectedSizes.value.set(id, item.size)
  } else {
    selected.value.delete(id)
    selectedSizes.value.delete(id)
  }
}

const lastSelectedIndex = ref<number | null>(null)

const toggleSingle = (id: string) => {
  const idx = allItems.value.findIndex(i => i.id === id)
  applySelection(id, !selected.value.has(id))
  if (idx !== -1) lastSelectedIndex.value = idx
}

const selectRange = (fromIdx: number, toIdx: number) => {
  const [start, end] = [fromIdx, toIdx].sort((a, b) => a - b)
  for (let i = start; i <= end; i++) applySelection(allItems.value[i].id, true)
}

const selectItemViaMenu = (id: string) => {
  selectMode.value = true
  applySelection(id, true)
  const idx = allItems.value.findIndex(i => i.id === id)
  if (idx !== -1) lastSelectedIndex.value = idx
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

  const next = new Set([...dragSnapshot.value, ...intersecting])
  selected.value = next
  selectedSizes.value = new Map([...next].map(id => [id, allItems.value.find(i => i.id === id)?.size ?? 0]))
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

  const id = (e.target as HTMLElement).closest<HTMLElement>("[data-item-id]")?.dataset.itemId ?? null

  if (e.shiftKey && id && lastSelectedIndex.value !== null) {
    const idx = allItems.value.findIndex(i => i.id === id)
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
  dragSnapshot.value = new Set(selected.value)
  isDragging.value = false

  window.addEventListener("pointermove", onGridPointerMove)
  window.addEventListener("pointerup", onGridPointerUp)
}

onBeforeUnmount(resetDragState)

const toggleSelectAllLoaded = () => {
  if (allSelectedLoaded.value) {
    allItems.value.forEach(item => {
      selected.value.delete(item.id)
      selectedSizes.value.delete(item.id)
    })
  } else {
    allItems.value.forEach(item => {
      selected.value.add(item.id)
      selectedSizes.value.set(item.id, item.size)
    })
  }
}

const bulkDelete = async () => {
  const ids = [...selected.value]
  const folderIds = ids.filter(id => folders.value.some(f => f.id === id))
  const imageIds = ids.filter(id => !folderIds.includes(id))

  let deletedImageIds: string[] = []
  if (imageIds.length > 0) {
    try {
      await client.bulkDeleteImages(imageIds)
      deletedImageIds = imageIds
    } catch (e) {
      push({ title: e instanceof HTTPException ? e.message : "Could not delete selected files", colour: "danger", delay: 6000 })
    }
  }

  let deletedFolderIds: string[] = []
  if (folderIds.length > 0) {
    const results = await Promise.allSettled(folderIds.map(id => client.deleteFolder(id)))
    deletedFolderIds = folderIds.filter((_, i) => results[i].status === "fulfilled")
    const failedCount = folderIds.length - deletedFolderIds.length
    if (failedCount > 0) {
      push({ title: `Could not delete ${failedCount} ${failedCount === 1 ? "folder" : "folders"}`, colour: "danger", delay: 6000 })
    }
  }

  images.value = images.value.filter(img => !deletedImageIds.includes(img.id))
  folders.value = folders.value.filter(f => !deletedFolderIds.includes(f.id))

  const deletedCount = deletedImageIds.length + deletedFolderIds.length
  if (deletedCount > 0) {
    push({ title: `Deleted ${deletedCount} ${deletedCount === 1 ? "item" : "items"}`, delay: 4000, colour: "info" })
    bumpMediaVersion()
  }

  for (const id of [...deletedImageIds, ...deletedFolderIds]) {
    selected.value.delete(id)
    selectedSizes.value.delete(id)
  }
  if (deletedCount === ids.length) selectMode.value = false
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
      :description="`This will permanently delete ${selectedSizeLabel}.`"
      confirm-text="Delete"
      confirm-colour="danger"
      :confirm-icon="Trash2Icon"
      :confirm-action="() => bulkDelete()" />

    <FolderNameDialog
      v-if="newFolderDialogOpen"
      title="New folder"
      confirm-text="Create"
      @dismiss="newFolderDialogOpen = false"
      :confirm-action="createFolder" />

    <FolderNameDialog
      v-if="renameFolderTarget"
      title="Rename folder"
      confirm-text="Rename"
      :initial-value="renameFolderTarget.name"
      @dismiss="renameFolderTarget = null"
      :confirm-action="renameFolder" />

    <ConfirmDialogue
      v-if="deleteFolderTarget"
      @dismiss="deleteFolderTarget = null"
      title="Delete folder?"
      :description="`This will permanently delete '${deleteFolderTarget.name}' and everything inside it.`"
      confirm-text="Delete"
      confirm-colour="danger"
      :confirm-icon="Trash2Icon"
      :confirm-action="() => deleteFolder()" />

    <MoveToFolderDialog
      v-if="moveDialogOpen"
      :exclude-folder-ids="selectedFolderIds"
      @dismiss="moveDialogOpen = false"
      @moved="bulkMove" />

    <div class="mb-3 flex flex-wrap items-center gap-1 text-sm">
      <button
        @click="openFolder(null)"
        @dragover="onBreadcrumbDragOver"
        @dragenter="onBreadcrumbDragEnter($event, ROOT_BREADCRUMB_KEY)"
        @dragleave="onBreadcrumbDragLeave(ROOT_BREADCRUMB_KEY)"
        @drop.prevent="onBreadcrumbDrop($event, null, 'My files')"
        :class="[
          'flex items-center gap-1 rounded-md px-1.5 py-0.5 font-medium transition hover:cursor-pointer',
          isBreadcrumbDragOver(ROOT_BREADCRUMB_KEY)
            ? 'ring-primary bg-primary/10 text-foreground ring-2 ring-inset'
            : 'text-muted hover:text-foreground',
          currentFolderId === null && 'text-foreground'
        ]">
        <HomeIcon :size="14" />
        My files
      </button>
      <template v-for="crumb in breadcrumb" :key="crumb.id">
        <ChevronRightIcon :size="14" class="text-muted shrink-0" />
        <button
          @click="openFolder(crumb.id)"
          @dragover="onBreadcrumbDragOver"
          @dragenter="onBreadcrumbDragEnter($event, crumb.id)"
          @dragleave="onBreadcrumbDragLeave(crumb.id)"
          @drop.prevent="onBreadcrumbDrop($event, crumb.id, crumb.name)"
          :class="[
            'truncate rounded-md px-1.5 py-0.5 font-medium transition hover:cursor-pointer',
            isBreadcrumbDragOver(crumb.id)
              ? 'ring-primary bg-primary/10 text-foreground ring-2 ring-inset'
              : 'text-muted hover:text-foreground',
            currentFolderId === crumb.id && 'text-foreground'
          ]">
          {{ crumb.name }}
        </button>
      </template>
    </div>

    <div class="mb-4 flex flex-wrap items-center justify-end gap-2">
      <select
        v-model="sort"
        @change="onSortChange"
        class="border-border bg-surface text-muted hover:text-foreground rounded-lg border px-3 py-2 text-sm font-medium transition hover:cursor-pointer">
        <option v-for="opt in sortOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
      </select>

      <div v-if="viewMode === 'grid'" class="border-border bg-surface flex items-center gap-0.5 rounded-lg border p-0.5">
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

      <div v-if="viewMode === 'grid'" class="border-border bg-surface flex items-center gap-0.5 rounded-lg border p-0.5">
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

      <div class="border-border bg-surface flex items-center gap-0.5 rounded-lg border p-0.5">
        <button
          v-for="opt in viewModeOptions"
          :key="opt.value"
          @click="setViewMode(opt.value)"
          :title="opt.title"
          :class="[
            'flex items-center rounded-md p-1.5 transition hover:cursor-pointer',
            viewMode === opt.value ? 'bg-primary text-white' : 'text-muted hover:bg-surface-2 hover:text-foreground'
          ]">
          <component :is="opt.icon" :size="16" />
        </button>
      </div>

      <button
        v-if="!selectMode"
        @click="newFolderDialogOpen = true"
        class="border-border bg-surface text-muted hover:bg-surface-2 hover:text-foreground flex items-center gap-1.5 rounded-lg border px-3 py-2 text-sm font-medium transition hover:cursor-pointer">
        <FolderPlusIcon :size="16" />
        New folder
      </button>

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
            <template v-if="selected.size > 0">&middot; {{ selectedSizeLabel }}</template>
          </span>
        </div>

        <div class="flex items-center gap-2">
          <Button variant="secondary" :icon="MoveIcon" :disabled="selected.size === 0" @click="moveDialogOpen = true">
            Move
          </Button>
          <Button variant="danger" :icon="Trash2Icon" :disabled="selected.size === 0" @click="bulkConfirmOpen = true">
            Delete selected
          </Button>
        </div>
      </div>
    </Transition>

    <div class="min-h-[16rem]" @pointerdown="onGridPointerDown">
      <Card
        v-if="viewMode === 'grid' && folders.length > 0"
        :class="['mb-3 grid gap-3', gridClasses[gridSize]]">
        <div
          v-for="folder in folders"
          :key="folder.id"
          :data-item-id="folder.id"
          :ref="el => setTileRef(folder.id, el as Element | null)">
          <FolderItem
            :folder="folder"
            mode="grid"
            :selectable="selectMode"
            :selected="selected.has(folder.id)"
            @open="openFolder(folder.id)"
            @rename="renameFolderTarget = folder"
            @delete="deleteFolderTarget = folder"
            @select="selectItemViaMenu"
            @move="moveItemViaMenu"
            @drop-image="id => moveImageToFolder(id, folder.id, folder.name)" />
        </div>
      </Card>

      <Card v-if="viewMode === 'grid'" :class="['grid gap-3', gridClasses[gridSize]]">
        <FileDrop v-if="!selectMode" :uploading="uploading" :progress="uploadProgress" @selected-files="uploadFiles" />

        <div
          v-for="image in images"
          :key="image.id"
          :data-item-id="image.id"
          :draggable="!selectMode"
          @dragstart="onImageDragStart($event, image.id)"
          @dragend="onImageDragEnd"
          :class="draggingImageId === image.id && 'opacity-40'"
          :ref="el => setTileRef(image.id, el as Element | null)">
          <ImageCard
            :image="image"
            :selectable="selectMode"
            :selected="selected.has(image.id)"
            :fit="fitMode"
            :thumbnail-width="thumbnailWidth"
            @select="selectItemViaMenu"
            @move="moveItemViaMenu"
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

      <template v-else>
        <FileDrop
          v-if="!selectMode"
          compact
          class="mb-3"
          :uploading="uploading"
          :progress="uploadProgress"
          @selected-files="uploadFiles" />

        <Card v-if="folders.length > 0" class="mb-3 overflow-hidden !p-0">
          <div
            v-for="folder in folders"
            :key="folder.id"
            :data-item-id="folder.id"
            :ref="el => setTileRef(folder.id, el as Element | null)">
            <FolderItem
              :folder="folder"
              mode="list"
              :selectable="selectMode"
              :selected="selected.has(folder.id)"
              @open="openFolder(folder.id)"
              @rename="renameFolderTarget = folder"
              @delete="deleteFolderTarget = folder"
              @select="selectItemViaMenu"
              @move="moveItemViaMenu"
              @drop-image="id => moveImageToFolder(id, folder.id, folder.name)" />
          </div>
        </Card>

        <Card class="overflow-hidden !p-0">
          <div
            v-for="image in images"
            :key="image.id"
            :data-item-id="image.id"
            :draggable="!selectMode"
            @dragstart="onImageDragStart($event, image.id)"
            @dragend="onImageDragEnd"
            :class="draggingImageId === image.id && 'opacity-40'"
            :ref="el => setTileRef(image.id, el as Element | null)">
            <ImageListRow
              :image="image"
              :selectable="selectMode"
              :selected="selected.has(image.id)"
              @select="selectItemViaMenu"
              @move="moveItemViaMenu"
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
      </template>
    </div>

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
