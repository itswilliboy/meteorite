<script setup lang="ts">
import type { Image } from "@/utils/type"
import { formatDistance } from "date-fns"
import { formatBytes } from "@/utils/format"
import ItemMenu from "./ItemMenu.vue"
import ImageCardFull from "./ImageCardFull.vue"
import useClient from "@/composables/useClient"
import { ref, useTemplateRef } from "vue"
import useToaster from "@/composables/useToaster"
import ImageOrVideo from "./ImageOrVideo.vue"
import { CheckIcon, MoreVerticalIcon, Trash2Icon } from "lucide-vue-next"
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
  select: [id: string]
  move: [id: string]
}>()

const deleteImage = () => {
  client.deleteImage(image.id)
  emit("pop", image.id)
  push({ title: `Deleted ${image.id}`, delay: 4000, colour: "info" })
}

const onRowClick = () => {
  if (!selectable) imageOpen.value = true
}

const menuRef = useTemplateRef<InstanceType<typeof ItemMenu>>("menuRef")

const onMenuButtonClick = (e: MouseEvent) => {
  menuRef.value?.openAtElement(e.currentTarget as HTMLElement)
}

const onContextMenu = (e: MouseEvent) => {
  if (selectable) return
  e.preventDefault()
  menuRef.value?.openAt(e.clientX, e.clientY)
}
</script>

<template>
  <div
    @click="onRowClick"
    @contextmenu="onContextMenu"
    :class="[
      'bg-surface hover:bg-surface-2 group relative flex w-full cursor-pointer items-center gap-3 border-b border-border px-2 py-2 transition last:border-b-0',
      selectable && 'select-none',
      selected && 'bg-primary/10'
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

    <div
      v-if="selectable"
      :class="[
        'grid size-6 shrink-0 place-items-center rounded-full border-2 shadow-sm transition',
        selected ? 'bg-primary border-primary text-white' : 'border-border bg-surface-2'
      ]">
      <CheckIcon v-if="selected" :size="15" :stroke-width="3" />
    </div>

    <div class="bg-surface-2 size-10 shrink-0 overflow-hidden rounded-lg">
      <ImageOrVideo
        :image="image"
        :thumbnail="true"
        :compact="true"
        fit="cover"
        :thumbnail-width="80"
        class="*:size-full! block! size-full *:rounded-none!" />
    </div>

    <div class="min-w-0 flex-1">
      <span
        :title="image.filename ?? image.id"
        :class="['text-foreground block truncate text-sm font-medium', !selectable && 'group-hover:underline']">
        {{ image.filename ?? image.id }}
      </span>
    </div>

    <div class="text-muted hidden w-20 shrink-0 truncate text-xs font-medium sm:block">
      {{ image.mimetype.split("/")[1]?.toUpperCase() }}
    </div>

    <div class="text-muted hidden w-20 shrink-0 text-right text-xs font-medium sm:block">
      {{ formatBytes(image.size) }}
    </div>

    <div class="text-muted hidden w-20 shrink-0 text-right text-xs font-medium md:block">{{ image.views }} views</div>

    <div class="text-muted hidden w-28 shrink-0 text-right text-xs font-medium md:block">
      {{ formatDistance(new Date(), new Date(image.date)).replace("about ", "") }} ago
    </div>

    <ItemMenu
      ref="menuRef"
      :image="image"
      @select="$emit('select', image.id)"
      @move="$emit('move', image.id)"
      @delete="confirmOpen = true" />
    <button
      v-if="!selectable"
      @click.stop="onMenuButtonClick"
      class="text-muted hover:bg-surface-2 hover:text-foreground grid size-8 shrink-0 place-items-center rounded-full opacity-0 transition-opacity duration-200 hover:cursor-pointer group-hover:opacity-100">
      <MoreVerticalIcon :size="16" />
    </button>
  </div>
</template>
