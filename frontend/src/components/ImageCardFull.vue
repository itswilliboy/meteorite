<script setup lang="ts">
import { useKeydown } from "@/composables/useKeydown"
import type { Image } from "@/utils/type"
import ImageOrVideo from "./ImageOrVideo.vue"
import { FileImageIcon, FileVideoIcon, LinkIcon, Trash2Icon, XIcon } from "lucide-vue-next"
import { format } from "date-fns"
import useToaster from "@/composables/useToaster"
import useClient from "@/composables/useClient"
import { ref } from "vue"
import ConfirmDialogue from "./ConfirmDialogue.vue"

const { image } = defineProps<{ image: Image }>()

const client = useClient()
const { push } = useToaster()

const confirmOpen = ref<boolean>(false)

const isVideo = image.mimetype.split("/")[0] == "video"

const copyToClipboard = async () => {
  await navigator.clipboard.writeText(image.url)
  push({ title: "Copied to clipboard!", colour: "info", delay: 4000 })
}

const deleteImage = () => {
  client.deleteImage(image.id)
  emit("pop", image.id)
  push({ title: `Deleted ${image.id}`, delay: 4000, colour: "info" })
}

useKeydown(e => {
  if (e.key === "Escape") emit("dismiss")
})

const emit = defineEmits<{
  dismiss: []
  pop: [id: string]
}>()
</script>

<template>
  <div class="fixed inset-0 z-40 flex min-h-screen w-screen items-center justify-center">
    <ConfirmDialogue
      v-if="confirmOpen"
      @dismiss="confirmOpen = false"
      title="Delete image?"
      description="This action is irreversible."
      confirm-text="Delete"
      confirm-colour="danger"
      :confirm-icon="Trash2Icon"
      :confirm-action="() => deleteImage()" />
    <button @click="$emit('dismiss')" class="absolute top-4 left-4 z-40 cursor-pointer">
      <XIcon class="stroke-white" />
    </button>
    <div class="absolute h-full w-full bg-black/75" @click="$emit('dismiss')" />
    <div class="z-50 grid h-5/6 w-[95%] grid-cols-10 grid-rows-1 gap-4">
      <div
        class="col-span-8 flex h-full w-full items-center justify-center rounded-xl bg-white bg-[url(/grid.svg)] bg-size-[auto_250px] bg-center bg-repeat lg:bg-size-[auto_350px]">
        <ImageOrVideo :image="image" :preview="false" />
      </div>
      <div class="relative col-span-2 h-full w-full rounded-xl bg-white p-5">
        <div class="flex w-full items-center gap-2">
          <FileVideoIcon v-if="isVideo" class="size-8" />
          <FileImageIcon v-else class="size-10" />
          <a :href="image.url" class="text-4xl font-semibold hover:underline" target="_blank">{{ image.id }}</a>
        </div>
        <p class="pt-0.5 text-sm font-semibold text-black/50">
          Uploaded {{ format(new Date(image.date), "yyyy-MM-dd HH:mm:ss") }}
          <br />
          {{ image.mimetype.split("/")[1].toUpperCase() }}
          <br />
          {{ image.views }} views
        </p>
        <div
          class="absolute bottom-0 -ml-5 grid h-max w-full grid-cols-2 gap-2 p-5 text-center text-white *:flex *:cursor-pointer *:items-center *:justify-center *:gap-1 *:rounded-xl *:p-3 *:font-semibold">
          <button class="bg-primary" @click="copyToClipboard">
            <LinkIcon class="size-5" />
            Copy Link
          </button>
          <button class="bg-danger" @click="() => (confirmOpen = true)">
            <Trash2Icon />
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
