<script setup lang="ts">
import type { Image } from "@/utils/type"
import { formatDistance } from "date-fns"
import ImageButtons from "./ImageButtons.vue"
import ImageCardFull from "./ImageCardFull.vue"
import useClient from "@/composables/useClient"
import { ref } from "vue"
import useToaster from "@/composables/useToaster"
import ImageOrVideo from "./ImageOrVideo.vue"
import { Trash2Icon } from "lucide-vue-next"
import ConfirmDialogue from "./ConfirmDialogue.vue"

const { image } = defineProps<{ image: Image }>()

const client = useClient()
const { push } = useToaster()

const imageOpen = ref<boolean>(false)
const confirmOpen = ref<boolean>(false)

const emit = defineEmits<{
  pop: [id: string]
}>()

const deleteImage = () => {
  client.deleteImage(image.id)
  emit("pop", image.id)
  push({ title: `Deleted ${image.id}`, delay: 4000, colour: "info" })
}
</script>

<template>
  <div class="group relative size-64 rounded-xl shadow-sm">
    <Transition>
      <ImageCardFull v-if="imageOpen" :image="image" @dismiss="imageOpen = false" @pop="$emit('pop', image.id)" />
    </Transition>
    <ConfirmDialogue
      v-if="confirmOpen"
      @dismiss="confirmOpen = false"
      title="Delete image?"
      description="This action is irreversible."
      confirm-text="Delete"
      confirm-colour="danger"
      :confirm-icon="Trash2Icon"
      :confirm-action="() => deleteImage()" />
    <ImageButtons
      :image="image"
      @delete="confirmOpen = true"
      class="pointer-events-none absolute top-2 right-2 z-10 opacity-0 transition-opacity duration-200 group-hover:pointer-events-auto group-hover:opacity-100" />
    <div @click="imageOpen = true" class="cursor-pointer transition-all duration-200 hover:brightness-75">
      <ImageOrVideo :image="image" />
    </div>
    <div class="bg-background absolute bottom-0 flex min-h-12 w-full justify-between p-1">
      <a :href="image.url" class="text-xl font-semibold hover:underline" target="_blank">{{ image.id }}</a>
      <div class="flex flex-col text-right text-sm font-bold text-gray-400">
        <p class="line-clamp-1">
          {{ formatDistance(new Date(), new Date(image.date)).replace("about ", "") }}
          ago
        </p>
        <p>{{ image.views }} views &middot; {{ image.mimetype.split("/")[1].toUpperCase() }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.v-enter-active,
.v-leave-active {
  transition: opacity 0.2s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
}
</style>
