<script setup lang="ts">
import Card from "@/components/Card.vue"
import ImageCard from "@/components/ImageCard.vue"
import PageContainer from "@/components/PageContainer.vue"
import FileDrop from "@/components/FileDrop.vue"
import ConfirmDialogue from "@/components/ConfirmDialogue.vue"
import useClient from "@/composables/useClient"
import useToaster from "@/composables/useToaster"
import { bumpMediaVersion } from "@/composables/useMediaVersion"
import type { Image, PaginatedResponse } from "@/utils/type"
import { computed, onMounted, ref } from "vue"
import { ChevronLeft, ChevronRight, ListChecksIcon, Trash2Icon, XIcon } from "lucide-vue-next"

defineOptions({ name: "DashImageView" })

const client = useClient()
const { push } = useToaster()

const response = ref<PaginatedResponse<Image[]> | null>(null)
const uploading = ref(false)
const uploadProgress = ref({ done: 0, total: 0 })

const selectMode = ref(false)
const selected = ref<Set<string>>(new Set())
const bulkConfirmOpen = ref(false)

const allSelectedOnPage = computed(() => {
  const data = response.value?.data ?? []
  return data.length > 0 && data.every(img => selected.value.has(img.id))
})

// This is the only view that ever mutates media, and it already keeps its own
// list in sync locally on upload/delete — so a plain onMounted (kept alive by
// <KeepAlive> in App.vue) is enough, no need to refetch on every revisit.
onMounted(async () => {
  response.value = await client.getImages(0)
})

const setPage = async (page: number) => {
  response.value = await client.getImages(page)
}

const toggleSelectMode = () => {
  selectMode.value = !selectMode.value
  selected.value.clear()
}

const toggleSelect = (id: string) => {
  if (selected.value.has(id)) selected.value.delete(id)
  else selected.value.add(id)
}

const toggleSelectAllOnPage = () => {
  const data = response.value?.data ?? []
  if (allSelectedOnPage.value) {
    data.forEach(img => selected.value.delete(img.id))
  } else {
    data.forEach(img => selected.value.add(img.id))
  }
}

const bulkDelete = async () => {
  const ids = [...selected.value]
  await Promise.all(ids.map(id => client.deleteImage(id)))

  if (response.value) {
    response.value.data = response.value.data.filter(img => !selected.value.has(img.id))
  }

  push({ title: `Deleted ${ids.length} ${ids.length === 1 ? "item" : "items"}`, delay: 4000, colour: "info" })
  selected.value.clear()
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

  // TODO: account for page
  response.value = await client.getImages(0)
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
      :description="`This will permanently delete ${selected.size} ${selected.size === 1 ? 'item' : 'items'}.`"
      confirm-text="Delete"
      confirm-colour="danger"
      :confirm-icon="Trash2Icon"
      :confirm-action="() => bulkDelete()" />

    <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
      <div class="flex items-center gap-1">
        <button
          :disabled="!(response?.hasPrev ?? false)"
          @click="setPage((response?.page ?? 0) - 1)"
          class="border-border bg-surface text-muted hover:bg-surface-2 hover:text-foreground grid size-9 place-items-center rounded-lg border transition hover:cursor-pointer disabled:pointer-events-none disabled:opacity-40">
          <ChevronLeft :size="18" />
        </button>

        <p class="text-muted px-3 text-sm font-medium tabular-nums">
          Page
          <span class="text-foreground font-semibold">{{ (response?.page ?? 0) + 1 }}</span>
        </p>

        <button
          :disabled="!(response?.hasNext ?? false)"
          @click="setPage((response?.page ?? 0) + 1)"
          class="border-border bg-surface text-muted hover:bg-surface-2 hover:text-foreground grid size-9 place-items-center rounded-lg border transition hover:cursor-pointer disabled:pointer-events-none disabled:opacity-40">
          <ChevronRight :size="18" />
        </button>
      </div>

      <button
        @click="toggleSelectMode"
        :class="[
          'flex items-center gap-1.5 rounded-lg border px-3 py-2 text-sm font-medium transition hover:cursor-pointer',
          selectMode
            ? 'bg-primary border-primary text-white'
            : 'border-border bg-surface text-muted hover:bg-surface-2 hover:text-foreground',
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
            <input type="checkbox" :checked="allSelectedOnPage" @change="toggleSelectAllOnPage" class="accent-primary size-4" />
            Select all on page
          </label>
          <span class="text-muted text-sm">{{ selected.size }} selected</span>
        </div>

        <button
          :disabled="selected.size === 0"
          @click="bulkConfirmOpen = true"
          class="bg-danger flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-semibold text-white transition hover:cursor-pointer hover:opacity-90 disabled:pointer-events-none disabled:opacity-40">
          <Trash2Icon :size="16" />
          Delete selected
        </button>
      </div>
    </Transition>

    <Card class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8">
      <FileDrop v-if="!selectMode" :uploading="uploading" :progress="uploadProgress" @selected-files="uploadFiles" />

      <template v-if="response">
        <ImageCard
          v-for="image in response!.data"
          :image="image"
          :key="image.id"
          :selectable="selectMode"
          :selected="selected.has(image.id)"
          @toggle-select="toggleSelect"
          @pop="
            id => {
              const images = response!.data
              images.splice(
                images.findIndex(img => img.id === id),
                1
              )
              bumpMediaVersion()
            }
          " />
      </template>
    </Card>
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
