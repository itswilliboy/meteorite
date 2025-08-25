<script setup lang="ts">
import Card from "@/components/Card.vue"
import ImageCard from "@/components/ImageCard.vue"
import PageContainer from "@/components/PageContainer.vue"
import useClient from "@/composables/useClient"
import type { Image } from "@/utils/type"
import { onMounted, ref } from "vue"

const images = ref<Image[]>([])

const client = useClient()
onMounted(async () => {
  images.value = await client.getImages()
})
</script>

<template>
  <PageContainer title="Gallery">
    <Card class="flex flex-wrap justify-evenly gap-4">
      <ImageCard
        v-for="image in images"
        :image="image"
        @pop="
          id => {
            console.log('slice', id)
            images.splice(
              images.findIndex(img => img.id === id),
              1
            )
          }
        " />
    </Card>
  </PageContainer>
</template>
