<script setup lang="ts">
import { useHover } from "@/composables/useHover"
import type { Image } from "@/utils/type"
import { computed, useTemplateRef, watch } from "vue"

const { image, preview = true, thumbnail = false } = defineProps<{ image: Image; preview?: boolean; thumbnail?: boolean }>()

const isVideo = image.mimetype.split("/")[0] == "video"
const mediaUrl = computed(() => {
  if (isVideo) return image.url

  const url = new URL(image.url)
  const params = new URLSearchParams()

  if (preview) params.append("d", "true")
  if (thumbnail) params.append("width", "256")
  url.search = params.toString()

  return url.toString()
})

const vref = useTemplateRef<HTMLVideoElement>("v")
const isHovered = useHover(vref)

watch(isHovered, after => {
  if (!preview) return
  if (after) {
    const v = vref.value!
    v.currentTime = 0
    v.play()
  } else {
    vref.value!.pause()
  }
})
</script>

<template>
  <span :class="preview ? '*:size-64 *:rounded-t-xl *:object-cover' : '*:size-full *:object-scale-down'">
    <video ref="v" v-if="isVideo" :src="mediaUrl" :controls="!preview">Failed to load video...</video>
    <img v-else :src="mediaUrl" />
  </span>
</template>
