<script setup lang="ts">
import type { Image } from "@/utils/type"
import { formatDistance } from "date-fns"
import ImageButtons from "./ImageButtons.vue"
import useClient from "@/composables/useClient"

const { image } = defineProps<{ image: Image }>()

const isVideo = image.mimetype.split("/")[0] == "video"
const client = useClient()

defineEmits<{
  pop: [id: string]
}>()
</script>

<template>
  <div class="group relative size-64 rounded-xl shadow-sm">
    <ImageButtons
      :image="image"
      @delete="
        id => {
          client.deleteImage(id)
          $emit('pop', id)
        }
      "
      class="pointer-events-none absolute top-2 right-2 opacity-0 transition-opacity duration-200 group-hover:pointer-events-auto group-hover:opacity-100" />
    <video v-if="isVideo" :src="image.url" muted autoplay loop>Failed to load video...</video>
    <img v-else :src="image.url" class="size-64 rounded-t-xl object-cover" />
    <div class="bg-background absolute bottom-0 flex min-h-12 w-full justify-between p-1">
      <span class="text-xl font-semibold">{{ image.id }}</span>
      <span class="flex flex-col text-right text-sm font-bold text-gray-400">
        <span>{{ formatDistance(new Date(), new Date(image.date)).replace("about ", "") }} ago</span>
        <span>{{ image.mimetype.split("/")[1].toUpperCase() }}</span>
      </span>
    </div>
  </div>
</template>
