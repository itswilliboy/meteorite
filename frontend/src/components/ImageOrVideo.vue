<script setup lang="ts">
import { useHover } from "@/composables/useHover"
import { useMediaType } from "@/composables/useMediaType"
import { useActiveMedia } from "@/composables/useActiveMedia"
import type { Image } from "@/utils/type"
import { computed, onMounted, ref, useTemplateRef, watch } from "vue"
import { LucideMusic, LucidePlay, LucidePause, Volume2, VolumeX, FileText, File as LucideFile } from "lucide-vue-next"
import { formatDuration } from "@/utils/format"

const {
  image,
  preview = true,
  thumbnail = false,
  fit = "cover",
  thumbnailWidth = 320,
  compact = false
} = defineProps<{
  image: Image
  preview?: boolean
  thumbnail?: boolean
  fit?: "cover" | "contain"
  thumbnailWidth?: number
  compact?: boolean
}>()

const previewIconSize = compact ? 18 : 56
const playButtonSizeClass = compact ? "size-5" : "size-10"
const playIconSize = compact ? 11 : 18

const { isImage, isVideo, isAudio, isText, isOther } = useMediaType(image.mimetype)
const { setActive, clearActive } = useActiveMedia()

const mediaUrl = computed(() => {
  if (isVideo || isAudio) return image.url

  const url = new URL(image.url)
  const params = new URLSearchParams()

  if (preview) params.append("d", "true")
  if (thumbnail && isImage) params.append("width", String(thumbnailWidth))
  url.search = params.toString()

  return url.toString()
})

const textContent = ref<string | null>(null)
const textLoading = ref(false)
const textError = ref(false)

onMounted(() => {
  if (!isText || preview) return

  textLoading.value = true
  fetch(mediaUrl.value)
    .then(resp => {
      if (!resp.ok) throw new Error("Failed to fetch text preview")
      return resp.text()
    })
    .then(text => {
      const maxLen = 200_000
      textContent.value = text.length > maxLen ? text.slice(0, maxLen) + "\n\n… (truncated)" : text
    })
    .catch(() => {
      textError.value = true
    })
    .finally(() => {
      textLoading.value = false
    })
})

const coverUrl = computed(() => {
  if (!isAudio || !image.has_cover) return null

  const url = new URL(image.url)
  url.search = new URLSearchParams({ cover: "true", width: thumbnail ? "160" : "512" }).toString()
  return url.toString()
})

const aref = useTemplateRef<HTMLAudioElement>("a")
const vref = useTemplateRef<HTMLVideoElement>("v")
const mref = computed<HTMLMediaElement | null>(() => aref.value ?? vref.value)

const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const volume = ref(0.5)
const muted = ref(false)

watch(mref, m => {
  if (m) m.volume = volume.value
})

const progress = computed(() => (duration.value ? (currentTime.value / duration.value) * 100 : 0))

const durationLabel = computed(() => formatDuration((duration.value || (image.duration_ms ?? 0) / 1000) * 1000))
const currentTimeLabel = computed(() => formatDuration(currentTime.value * 1000))

const togglePlay = () => {
  const m = mref.value
  if (!m) return
  if (m.paused) m.play()
  else m.pause()
}

const onTime = () => {
  const m = mref.value
  if (m) currentTime.value = m.currentTime
}

const onLoadedMeta = () => {
  const m = mref.value
  if (m && isFinite(m.duration)) duration.value = m.duration
}

const onPlay = () => {
  isPlaying.value = true
  const m = mref.value
  // hover preview shouldn't interrupt
  if (m && !(isVideo && preview)) setActive(m)
}

const onPause = () => {
  isPlaying.value = false
  const m = mref.value
  if (m) clearActive(m)
}

const onEnded = () => {
  isPlaying.value = false
  currentTime.value = 0
  const m = mref.value
  if (m) clearActive(m)
}

const onSeek = (e: Event) => {
  const m = mref.value
  const total = m?.duration || (image.duration_ms ?? 0) / 1000
  if (!m || !total) return

  const pct = Number((e.target as HTMLInputElement).value)
  m.currentTime = (pct / 100) * total
  currentTime.value = m.currentTime
}

const onVolumeInput = (e: Event) => {
  const m = mref.value
  const v = Number((e.target as HTMLInputElement).value)
  volume.value = v
  muted.value = v === 0
  if (m) {
    m.volume = v
    m.muted = v === 0
  }
}

const toggleMute = () => {
  const m = mref.value
  if (!m) return
  m.muted = !m.muted
  muted.value = m.muted
}

const isHovered = useHover(vref)

watch(isHovered, after => {
  if (!preview) return
  if (after) {
    const v = vref.value!
    v.currentTime = 0
    v.muted = true
    v.play()
  } else {
    vref.value?.pause()
    vref.value!.muted = false
  }
})
</script>

<template>
  <span
    :class="
      preview
        ? `*:size-64 *:rounded-t-xl ${fit === 'contain' ? '*:object-contain' : '*:object-cover'}`
        : 'flex size-full items-center justify-center overflow-hidden *:max-h-full *:max-w-full *:object-scale-down'
    ">
    <video ref="v" v-if="isVideo && preview" :src="mediaUrl" preload="metadata">Failed to load video...</video>

    <div v-else-if="isVideo" class="relative flex size-full items-center justify-center overflow-hidden rounded-xl bg-black">
      <video
        ref="v"
        :src="mediaUrl"
        preload="metadata"
        @click="togglePlay"
        @timeupdate="onTime"
        @loadedmetadata="onLoadedMeta"
        @play="onPlay"
        @pause="onPause"
        @ended="onEnded"
        class="max-h-full max-w-full cursor-pointer object-contain" />

      <div
        class="pointer-events-none absolute inset-x-0 bottom-0 bg-linear-to-t from-black/85 via-black/40 to-transparent pt-10">
        <div class="pointer-events-auto flex w-full items-center gap-3 px-4 pb-3">
          <button
            @click="togglePlay"
            :title="isPlaying ? 'Pause' : 'Play'"
            class="bg-primary grid size-9 shrink-0 place-items-center rounded-full text-white shadow-sm transition hover:cursor-pointer hover:opacity-90">
            <LucidePause v-if="isPlaying" :size="16" />
            <LucidePlay v-else :size="16" class="ml-0.5" />
          </button>

          <span class="w-9 shrink-0 text-right text-xs text-white/80 tabular-nums">{{ currentTimeLabel }}</span>

          <div class="group relative h-1.5 flex-1 rounded-full bg-white/25">
            <div
              class="bg-primary pointer-events-none absolute inset-y-0 left-0 rounded-full"
              :style="{ width: progress + '%' }" />
            <div
              class="bg-primary pointer-events-none absolute top-1/2 size-3 -translate-y-1/2 rounded-full shadow transition-transform group-hover:scale-110"
              :style="{ left: `calc(${progress}% - 6px)` }" />
            <input
              type="range"
              min="0"
              max="100"
              step="0.1"
              :value="progress"
              @input="onSeek"
              class="absolute inset-0 size-full cursor-pointer appearance-none opacity-0 [&::-moz-range-thumb]:size-0 [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:border-0 [&::-webkit-slider-thumb]:size-0 [&::-webkit-slider-thumb]:appearance-none" />
          </div>

          <span class="w-9 shrink-0 text-xs text-white/80 tabular-nums">{{ durationLabel }}</span>

          <button
            @click="toggleMute"
            :title="muted ? 'Unmute' : 'Mute'"
            class="shrink-0 text-white/80 hover:cursor-pointer hover:text-white">
            <VolumeX v-if="muted || volume === 0" :size="16" />
            <Volume2 v-else :size="16" />
          </button>

          <div class="relative h-1.5 w-16 shrink-0 rounded-full bg-white/25">
            <div
              class="pointer-events-none absolute inset-y-0 left-0 rounded-full bg-white/70"
              :style="{ width: (muted ? 0 : volume * 100) + '%' }" />
            <input
              type="range"
              min="0"
              max="1"
              step="0.01"
              :value="muted ? 0 : volume"
              @input="onVolumeInput"
              class="absolute inset-0 size-full cursor-pointer appearance-none opacity-0 [&::-moz-range-thumb]:size-0 [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:border-0 [&::-webkit-slider-thumb]:size-0 [&::-webkit-slider-thumb]:appearance-none" />
          </div>
        </div>
      </div>
    </div>

    <img
      v-else-if="isImage"
      :src="mediaUrl"
      :loading="thumbnail ? 'lazy' : undefined"
      :decoding="thumbnail ? 'async' : undefined" />

    <!-- Text files -->
    <div v-else-if="isText" class="relative flex size-full items-center justify-center overflow-hidden">
      <div v-if="preview" class="from-primary/15 to-primary/5 grid size-full place-items-center bg-linear-to-br">
        <FileText class="text-primary/30" :size="previewIconSize" :stroke-width="1.5" />
      </div>
      <div v-else class="bg-surface-2 flex size-full flex-col overflow-hidden rounded-xl">
        <div v-if="textLoading" class="text-muted grid size-full place-items-center text-sm">Loading preview...</div>
        <div v-else-if="textError" class="text-muted grid size-full place-items-center text-sm">
          Couldn't load a preview.
        </div>
        <pre
          v-else
          class="text-foreground size-full overflow-auto p-4 font-mono text-xs leading-relaxed whitespace-pre-wrap"
          >{{ textContent }}</pre
        >
      </div>
    </div>

    <!-- Anything else without a preview -->
    <div v-else-if="isOther" class="relative flex size-full items-center justify-center overflow-hidden">
      <div v-if="preview" class="from-primary/15 to-primary/5 grid size-full place-items-center bg-linear-to-br">
        <LucideFile class="text-primary/30" :size="previewIconSize" :stroke-width="1.5" />
      </div>
      <div v-else class="text-muted flex size-full flex-col items-center justify-center gap-3 text-sm">
        <LucideFile :size="56" :stroke-width="1.2" class="text-muted/40" />
        No preview available
      </div>
    </div>

    <div v-else class="relative flex size-full items-center justify-center overflow-hidden">
      <!-- Mini preview tile -->
      <div
        v-if="preview"
        class="from-primary/15 to-primary/5 relative grid size-full place-items-center overflow-hidden bg-linear-to-br">
        <img
          v-if="coverUrl"
          :src="coverUrl"
          :loading="thumbnail ? 'lazy' : undefined"
          class="pointer-events-none absolute inset-0 size-full object-cover brightness-75" />
        <LucideMusic
          v-else
          class="text-primary/25 pointer-events-none absolute"
          :size="previewIconSize"
          :stroke-width="1.5" />

        <button
          v-if="!compact"
          @click.stop.prevent="togglePlay"
          :title="isPlaying ? 'Pause' : 'Play'"
          :class="[
            'bg-primary/90 relative z-10 grid place-items-center rounded-full text-white shadow-md transition hover:scale-105 hover:cursor-pointer',
            playButtonSizeClass
          ]">
          <LucidePause v-if="isPlaying" :size="playIconSize" />
          <LucidePlay v-else :size="playIconSize" class="ml-0.5" />
        </button>

        <div class="pointer-events-none absolute inset-x-0 bottom-0 h-1 bg-black/10">
          <div class="bg-primary h-full" :style="{ width: progress + '%' }" />
        </div>
      </div>

      <!-- Full "now playing" view -->
      <div
        v-else
        class="from-surface-2 to-surface flex size-full flex-col items-center justify-center gap-6 overflow-y-auto bg-linear-to-b p-6 sm:gap-8 sm:p-10">
        <div class="relative aspect-square w-full max-w-[min(52vh,32rem)] shrink overflow-hidden rounded-3xl shadow-2xl">
          <img v-if="coverUrl" :src="coverUrl" class="size-full object-cover" />
          <div v-else class="from-primary/25 to-primary/5 flex size-full items-center justify-center bg-linear-to-br">
            <LucideMusic :size="112" :stroke-width="1.2" class="text-primary/40" />
          </div>
        </div>

        <div class="w-full max-w-lg shrink-0">
          <p v-if="image.filename" class="text-foreground w-full truncate text-center text-lg font-semibold sm:text-xl">
            {{ image.filename }}
          </p>

          <div class="mt-6 flex w-full items-center gap-3 sm:mt-8">
            <span class="text-muted w-11 shrink-0 text-right text-sm tabular-nums">{{ currentTimeLabel }}</span>

            <div class="bg-surface-2 group relative h-2 flex-1 rounded-full">
              <div
                class="bg-primary pointer-events-none absolute inset-y-0 left-0 rounded-full"
                :style="{ width: progress + '%' }" />
              <div
                class="bg-primary pointer-events-none absolute top-1/2 size-4 -translate-y-1/2 rounded-full shadow transition-transform group-hover:scale-110"
                :style="{ left: `calc(${progress}% - 8px)` }" />
              <input
                type="range"
                min="0"
                max="100"
                step="0.1"
                :value="progress"
                @input="onSeek"
                class="absolute inset-0 size-full cursor-pointer appearance-none opacity-0 [&::-moz-range-thumb]:size-0 [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:border-0 [&::-webkit-slider-thumb]:size-0 [&::-webkit-slider-thumb]:appearance-none" />
            </div>

            <span class="text-muted w-11 shrink-0 text-sm tabular-nums">{{ durationLabel }}</span>
          </div>

          <div class="mt-6 flex w-full items-center justify-center gap-6 sm:mt-8 sm:gap-8">
            <div class="flex items-center gap-2">
              <button
                @click="toggleMute"
                :title="muted ? 'Unmute' : 'Mute'"
                class="text-muted hover:text-foreground shrink-0 hover:cursor-pointer">
                <VolumeX v-if="muted || volume === 0" :size="18" />
                <Volume2 v-else :size="18" />
              </button>

              <div class="bg-surface-2 relative h-2 w-24 rounded-full sm:w-28">
                <div
                  class="bg-muted pointer-events-none absolute inset-y-0 left-0 rounded-full"
                  :style="{ width: (muted ? 0 : volume * 100) + '%' }" />
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.01"
                  :value="muted ? 0 : volume"
                  @input="onVolumeInput"
                  class="absolute inset-0 size-full cursor-pointer appearance-none opacity-0 [&::-moz-range-thumb]:size-0 [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:border-0 [&::-webkit-slider-thumb]:size-0 [&::-webkit-slider-thumb]:appearance-none" />
              </div>
            </div>

            <button
              @click="togglePlay"
              :title="isPlaying ? 'Pause' : 'Play'"
              class="bg-primary grid size-16 shrink-0 place-items-center rounded-full text-white shadow-lg transition hover:scale-105 hover:cursor-pointer hover:opacity-90 sm:size-20">
              <LucidePause v-if="isPlaying" :size="26" />
              <LucidePlay v-else :size="26" class="ml-0.5" />
            </button>

            <div class="w-25 sm:w-28" />
          </div>
        </div>
      </div>

      <audio
        ref="a"
        :src="mediaUrl"
        :preload="preview ? 'none' : 'metadata'"
        @timeupdate="onTime"
        @loadedmetadata="onLoadedMeta"
        @play="onPlay"
        @pause="onPause"
        @ended="onEnded" />
    </div>
  </span>
</template>
