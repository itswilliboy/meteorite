<script setup lang="ts">
import { useKeydown } from "@/composables/useKeydown"
import { useMediaType } from "@/composables/useMediaType"
import type { Image } from "@/utils/type"
import ImageOrVideo from "./ImageOrVideo.vue"
import {
  FileImageIcon,
  FileVideoIcon,
  FileMusicIcon,
  FileTextIcon,
  FileIcon,
  DownloadIcon,
  LinkIcon,
  MoveIcon,
  PencilIcon,
  Trash2Icon,
  XIcon
} from "lucide-vue-next"
import { format } from "date-fns"
import { formatBytes, formatDuration, formatBitrate } from "@/utils/format"
import useToaster from "@/composables/useToaster"
import useClient from "@/composables/useClient"
import { computed, ref } from "vue"
import ConfirmDialogue from "./ConfirmDialogue.vue"
import { Button } from "@/components/common"

const { image } = defineProps<{ image: Image }>()

const client = useClient()
const { push } = useToaster()

const confirmOpen = ref<boolean>(false)

const { isImage, isVideo, isAudio, isText } = useMediaType(image.mimetype)

const copyToClipboard = async () => {
  await navigator.clipboard.writeText(image.url)
  push({ title: "Copied to clipboard!", colour: "info", delay: 4000 })
}

const downloadFile = () => {
  const url = new URL(image.url)
  url.searchParams.set("download", "true")
  window.location.href = url.toString()
}

const deleteImage = () => {
  client.deleteImage(image.id)
  emit("pop", image.id)
  push({ title: `Deleted ${image.id}`, delay: 4000, colour: "info" })
}

const channelLabel = (channels: number): string => {
  if (channels === 1) return "Mono"
  if (channels === 2) return "Stereo"
  return `${channels} channels`
}

const metadataRows = computed(() => {
  const rows: { label: string; value: string }[] = []

  rows.push({ label: "Type", value: image.mimetype.split("/")[1]?.toUpperCase() ?? image.mimetype })
  rows.push({ label: "Size", value: formatBytes(image.size) })

  if (image.width && image.height) rows.push({ label: "Dimensions", value: `${image.width} × ${image.height}` })
  if (image.duration_ms) rows.push({ label: "Duration", value: formatDuration(image.duration_ms) })
  if (isVideo && image.framerate) rows.push({ label: "Framerate", value: `${image.framerate.toFixed(2)} fps` })
  if (image.codec) rows.push({ label: "Codec", value: image.codec.toUpperCase() })
  if (image.bitrate) rows.push({ label: "Bitrate", value: formatBitrate(image.bitrate) })
  if (isAudio && image.sample_rate)
    rows.push({ label: "Sample Rate", value: `${(image.sample_rate / 1000).toFixed(1)} kHz` })
  if (isAudio && image.channels) rows.push({ label: "Channels", value: channelLabel(image.channels) })

  rows.push({ label: "Views", value: image.views.toLocaleString() })
  rows.push({ label: "Uploaded", value: format(new Date(image.date), "yyyy-MM-dd HH:mm:ss") })

  return rows
})

useKeydown(e => {
  if (e.key === "Escape") emit("dismiss")
})

const emit = defineEmits<{
  dismiss: []
  pop: [id: string]
  rename: [id: string]
  move: [id: string]
}>()
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-60 flex h-dvh w-screen items-center justify-center overflow-hidden">
      <ConfirmDialogue
        v-if="confirmOpen"
        @dismiss="confirmOpen = false"
        title="Delete image?"
        description="This action is irreversible."
        confirm-text="Delete"
        confirm-colour="danger"
        :confirm-icon="Trash2Icon"
        :confirm-action="() => deleteImage()" />

      <Transition name="backdrop" appear>
        <div class="absolute h-full w-full bg-black/70 backdrop-blur-sm" @click="$emit('dismiss')" />
      </Transition>

      <button
        @click="$emit('dismiss')"
        class="absolute top-4 left-4 z-40 grid size-9 cursor-pointer place-items-center rounded-full bg-black/40 text-white backdrop-blur-sm transition hover:bg-black/60">
        <XIcon :size="18" />
      </button>

      <Transition name="panel" appear>
        <div class="z-50 flex h-[92%] w-[95%] flex-col gap-3 lg:grid lg:h-5/6 lg:grid-cols-10 lg:grid-rows-1 lg:gap-4">
          <div
            :class="[
              'bg-surface border-border/60 flex min-h-0 flex-1 items-center justify-center overflow-hidden rounded-2xl border shadow-2xl lg:col-span-8',
              isImage && 'bg-[url(/grid.svg)] bg-size-[auto_250px] bg-center bg-repeat lg:bg-size-[auto_350px]'
            ]">
            <ImageOrVideo :image="image" :preview="false" />
          </div>
          <div
            class="bg-surface border-border/60 relative flex max-h-[45%] min-h-0 shrink-0 flex-col rounded-2xl border p-5 shadow-2xl lg:col-span-2 lg:h-full lg:max-h-none">
            <div class="flex w-full items-center gap-2">
              <FileVideoIcon v-if="isVideo" class="text-muted size-6 shrink-0" />
              <FileMusicIcon v-else-if="isAudio" class="text-muted size-6 shrink-0" />
              <FileImageIcon v-else-if="isImage" class="text-muted size-6 shrink-0" />
              <FileTextIcon v-else-if="isText" class="text-muted size-6 shrink-0" />
              <FileIcon v-else class="text-muted size-6 shrink-0" />
              <a
                :href="image.url"
                class="truncate text-xl font-semibold hover:underline lg:text-2xl"
                target="_blank"
                :title="image.filename ?? image.id">
                {{ image.filename ?? image.id }}
              </a>
            </div>
            <p v-if="image.filename" class="text-muted truncate pt-0.5 text-xs" :title="image.id">
              {{ image.id }}
            </p>

            <dl class="mt-5 min-h-0 flex-1 space-y-2.5 overflow-y-auto pr-1 text-sm">
              <div
                v-for="row in metadataRows"
                :key="row.label"
                class="border-border/70 flex items-baseline justify-between gap-3 border-b pb-2.5 last:border-b-0">
                <dt class="text-muted shrink-0">{{ row.label }}</dt>
                <dd class="truncate text-right font-medium">{{ row.value }}</dd>
              </div>
            </dl>

            <div class="mt-5 grid w-full shrink-0 grid-cols-2 gap-2">
              <Button :icon="LinkIcon" @click="copyToClipboard">Copy</Button>
              <Button variant="secondary" :icon="DownloadIcon" @click="downloadFile">Download</Button>
              <Button variant="secondary" :icon="PencilIcon" @click="$emit('rename', image.id)">Rename</Button>
              <Button variant="secondary" :icon="MoveIcon" @click="$emit('move', image.id)">Move</Button>
              <Button
                variant="danger"
                :icon="Trash2Icon"
                class="col-span-2"
                @click="() => (confirmOpen = true)">
                Delete
              </Button>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>
