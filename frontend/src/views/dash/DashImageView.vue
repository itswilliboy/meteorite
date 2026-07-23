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
import { computed, nextTick, ref, useTemplateRef } from "vue"
import { ListChecksIcon, LucideLoader2, Trash2Icon, XIcon } from "lucide-vue-next"

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

const loadMore = async (): Promise<boolean> => {
  if (loadingMore.value || !hasNext.value) return false

  loadingMore.value = true
  try {
    const resp = await client.getImages(page.value + 1)
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

const toggleSelectMode = () => {
  selectMode.value = !selectMode.value
  selected.value.clear()
  selectedSizes.value.clear()
}

const toggleSelect = (id: string) => {
  if (selected.value.has(id)) {
    selected.value.delete(id)
    selectedSizes.value.delete(id)
  } else {
    selected.value.add(id)
    const image = images.value.find(img => img.id === id)
    if (image) selectedSizes.value.set(id, image.size)
  }
}

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

    <Card class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8">
      <FileDrop v-if="!selectMode" :uploading="uploading" :progress="uploadProgress" @selected-files="uploadFiles" />

      <ImageCard
        v-for="image in images"
        :image="image"
        :key="image.id"
        :selectable="selectMode"
        :selected="selected.has(image.id)"
        @toggle-select="toggleSelect"
        @pop="
          id => {
            images.splice(
              images.findIndex(img => img.id === id),
              1
            )
            bumpMediaVersion()
          }
        " />
    </Card>

    <div ref="sentinel" class="flex h-10 items-center justify-center">
      <LucideLoader2 v-if="loadingMore" :size="20" class="text-muted animate-spin" />
    </div>
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
