<script setup lang="ts">
import { useHover } from "@/composables/useHover"
import type { Image } from "@/utils/type"
import { useTemplateRef, watch } from "vue"

const { image, preview = true } = defineProps<{ image: Image; preview?: boolean }>()

const isVideo = image.mimetype.split("/")[0] == "video"

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
    <video ref="v" v-if="isVideo" :src="image.url + '?d=true'" :controls="!preview">Failed to load video...</video>
    <img v-else :src="image.url + '?d=true'" />
  </span>
</template>
