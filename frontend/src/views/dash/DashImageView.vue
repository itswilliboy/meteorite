<script setup lang="ts">
import Card from "@/components/Card.vue"
import ImageCard from "@/components/ImageCard.vue"
import PageContainer from "@/components/PageContainer.vue"
import useClient from "@/composables/useClient"
import type { Image, PaginatedResponse } from "@/utils/type"
import { onMounted, ref } from "vue"

const client = useClient()

const response = ref<PaginatedResponse<Image[]> | null>(null)

onMounted(async () => {
  response.value = await client.getImages(0)
})

const setPage = async (page: number) => {
  response.value = await client.getImages(page)
}
</script>

<template>
  <PageContainer title="Gallery">
    <div class="*:bg-primary flex gap-2 text-white *:cursor-pointer *:rounded-xl *:p-2 *:disabled:cursor-not-allowed">
      <button :disabled="!(response?.hasPrev ?? false)" @click="setPage((response?.page ?? 0) - 1)">Left</button>
      <p>Page: {{ response?.page ?? 0 }}</p>
      <button :disabled="!(response?.hasNext ?? false)" @click="setPage((response?.page ?? 0) + 1)">Right</button>
    </div>
    <Card
      class="grid grid-cols-1 place-items-center gap-y-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-6"
      v-if="response">
      <ImageCard
        v-for="image in response!.data"
        :image="image"
        :key="image.id"
        @pop="
          id => {
            const images = response!.data
            images.splice(
              images.findIndex(img => img.id === id),
              1
            )
          }
        " />
    </Card>
  </PageContainer>
</template>
