<script setup lang="ts">
import useClient from "@/composables/useClient"
import { useKeydown } from "@/composables/useKeydown"
import type { BreadcrumbEntry, Folder } from "@/utils/type"
import { ChevronRightIcon, FolderIcon, HomeIcon, MoveIcon, XIcon } from "lucide-vue-next"
import { onMounted, ref } from "vue"

const { excludeFolderId } = defineProps<{
  excludeFolderId?: string
}>()

const emit = defineEmits<{
  dismiss: []
  moved: [folderId: string | null]
}>()

const client = useClient()

const currentFolderId = ref<string | null>(null)
const folders = ref<Folder[]>([])
const breadcrumb = ref<BreadcrumbEntry[]>([])
const loading = ref(false)

const load = async (folderId: string | null) => {
  loading.value = true
  try {
    const resp = await client.getFolders(folderId)
    folders.value = resp.folders.filter(f => f.id !== excludeFolderId)
    breadcrumb.value = resp.breadcrumb
    currentFolderId.value = folderId
  } finally {
    loading.value = false
  }
}

onMounted(() => load(null))

useKeydown(e => {
  if (e.key === "Escape") emit("dismiss")
})
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-[999] flex items-center justify-center p-4">
      <Transition name="backdrop" appear>
        <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="$emit('dismiss')" />
      </Transition>

      <Transition name="panel" appear>
        <div class="bg-surface border-border/60 relative z-50 flex h-100 w-120 max-w-full flex-col rounded-2xl border p-6 shadow-2xl">
          <div class="flex items-center justify-between gap-2">
            <h1 class="text-xl font-semibold">Move to folder</h1>
            <button
              @click="$emit('dismiss')"
              class="text-muted hover:bg-surface-2 hover:text-foreground grid size-8 shrink-0 place-items-center rounded-full transition hover:cursor-pointer">
              <XIcon :size="18" />
            </button>
          </div>

          <div class="text-muted mt-3 flex flex-wrap items-center gap-1 text-sm">
            <button
              @click="load(null)"
              :class="[
                'flex items-center gap-1 rounded-md px-1.5 py-0.5 hover:cursor-pointer',
                currentFolderId === null ? 'text-foreground font-semibold' : 'hover:text-foreground'
              ]">
              <HomeIcon :size="14" />
              My files
            </button>
            <template v-for="crumb in breadcrumb" :key="crumb.id">
              <ChevronRightIcon :size="14" class="shrink-0" />
              <button
                @click="load(crumb.id)"
                :class="[
                  'truncate rounded-md px-1.5 py-0.5 hover:cursor-pointer',
                  currentFolderId === crumb.id ? 'text-foreground font-semibold' : 'hover:text-foreground'
                ]">
                {{ crumb.name }}
              </button>
            </template>
          </div>

          <div class="border-border mt-3 min-h-0 flex-1 overflow-y-auto rounded-lg border">
            <div v-if="loading" class="text-muted grid h-full place-items-center text-sm">Loading...</div>
            <div v-else-if="folders.length === 0" class="text-muted grid h-full place-items-center text-sm">
              No subfolders here
            </div>
            <button
              v-else
              v-for="folder in folders"
              :key="folder.id"
              @click="load(folder.id)"
              class="hover:bg-surface-2 flex w-full items-center gap-2 border-b border-border px-3 py-2.5 text-left text-sm transition last:border-b-0 hover:cursor-pointer">
              <FolderIcon :size="18" class="text-primary shrink-0" />
              <span class="truncate">{{ folder.name }}</span>
            </button>
          </div>

          <button
            @click="$emit('moved', currentFolderId)"
            class="bg-primary mt-4 flex w-full shrink-0 items-center justify-center gap-1.5 rounded-xl p-3 text-sm font-semibold text-white transition hover:cursor-pointer hover:opacity-90">
            <MoveIcon class="size-4" />
            Move here
          </button>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>
