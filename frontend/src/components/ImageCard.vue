<script setup lang="ts">
import type { Image } from "@/utils/type"
import { formatDistance } from "date-fns"
import { formatBytes } from "@/utils/format"
import ImageButtons from "./ImageButtons.vue"
import ImageCardFull from "./ImageCardFull.vue"
import useClient from "@/composables/useClient"
import { ref } from "vue"
import useToaster from "@/composables/useToaster"
import ImageOrVideo from "./ImageOrVideo.vue"
import { CheckIcon, Trash2Icon } from "lucide-vue-next"
import ConfirmDialogue from "./ConfirmDialogue.vue"

const { image, selectable = false, selected = false } = defineProps<{
  image: Image
  selectable?: boolean
  selected?: boolean
}>()

const client = useClient()
const { push } = useToaster()

const imageOpen = ref<boolean>(false)
const confirmOpen = ref<boolean>(false)

const emit = defineEmits<{
  pop: [id: string]
  "toggle-select": [id: string]
}>()

const deleteImage = () => {
  client.deleteImage(image.id)
  emit("pop", image.id)
  push({ title: `Deleted ${image.id}`, delay: 4000, colour: "info" })
}

const onTileClick = () => {
  if (selectable) emit("toggle-select", image.id)
  else imageOpen.value = true
}
</script>

<template>
  <div
    :class="[
      'group relative aspect-square w-full overflow-hidden rounded-xl shadow-sm transition',
      selected && 'ring-primary ring-3 ring-offset-2',
    ]">
    <ImageCardFull v-if="imageOpen" :image="image" @dismiss="imageOpen = false" @pop="$emit('pop', image.id)" />
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
      v-if="!selectable"
      :image="image"
      @delete="confirmOpen = true"
      class="pointer-events-none absolute top-2 right-2 z-10 opacity-0 transition-opacity duration-200 group-hover:pointer-events-auto group-hover:opacity-100" />
    <div
      v-if="selectable"
      :class="[
        'absolute top-2 left-2 z-10 grid size-6 place-items-center rounded-full border-2 shadow-sm transition',
        selected ? 'bg-primary border-primary text-white' : 'border-white/80 bg-black/30',
      ]">
      <CheckIcon v-if="selected" :size="15" :stroke-width="3" />
    </div>
    <div
      @click="onTileClick"
      class="size-full cursor-pointer transition-all duration-200 hover:brightness-90">
      <ImageOrVideo :image="image" :thumbnail="true" class="!block size-full [&>*]:!size-full [&>*]:!rounded-none" />
    </div>

    <div class="pointer-events-none absolute inset-x-0 bottom-0 bg-linear-to-t from-black/75 via-black/35 to-transparent p-2.5 pt-8">
      <a
        :href="image.url"
        target="_blank"
        :title="image.filename ?? image.id"
        class="pointer-events-auto block truncate text-sm font-semibold text-white hover:underline">
        {{ image.filename ?? image.id }}
      </a>
      <div class="mt-0.5 flex items-center justify-between gap-2 text-[11px] font-medium text-white/70">
        <span class="truncate">
          {{ formatDistance(new Date(), new Date(image.date)).replace("about ", "") }} ago
        </span>
        <span class="shrink-0">
          {{ image.views }} views &middot; {{ image.mimetype.split("/")[1].toUpperCase() }} &middot; {{ formatBytes(image.size) }}
        </span>
      </div>
    </div>
  </div>
</template>
