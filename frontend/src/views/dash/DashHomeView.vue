<script setup lang="ts">
import { onMounted, ref } from "vue"

import PageContainer from "@/components/PageContainer.vue"
import Card from "@/components/Card.vue"
import useClient from "@/composables/useClient"

import type { DashboardStats } from "@/utils/type"

const client = useClient()

const data = ref<Option<DashboardStats>>(null)

onMounted(async () => {
  data.value = await client.dashboardStats()
})

const formatBytesToMebibytes = (bytes: number): string => {
  const ONE_MEBIBYTE = 1_048_576
  return new Intl.NumberFormat("en-GB").format(bytes / ONE_MEBIBYTE) + " MiB"
}
</script>

<template>
  <PageContainer title="Dashboard" className="space-y-3">
    <section class="flex w-full gap-3">
      <Card class="leading-none">
        <span class="text-sm font-semibold text-gray-500">Total Images</span>
        <h1 class="text-2xl font-bold">{{ data ? data?.total_images : "Loading..." }}</h1>
      </Card>

      <Card class="leading-none">
        <span class="text-xs font-semibold text-gray-500">Monthly Uploads</span>
        <h1 class="text-2xl font-bold">{{ data ? data?.monthly_uploads : "Loading..." }}</h1>
      </Card>

      <Card class="leading-none">
        <span class="text-xs font-semibold text-gray-500">Storage Used</span>
        <h1 class="text-2xl font-bold">
          {{ data ? formatBytesToMebibytes(data?.storage_usage!) : "Loading..." }}
        </h1>
      </Card>

      <Card class="leading-none">
        <span class="text-xs font-semibold text-gray-500">Bandwidth</span>
        <h1 class="text-2xl font-bold">{{ data ? formatBytesToMebibytes(data?.user_bandwidth!) : "Loading..." }}</h1>
      </Card>
    </section>

    <section>
      <Card class="py-24"></Card>
    </section>

    <section>
      <Card class="pt-3">
        <span class="text-xs font-semibold text-gray-500">Recent Activity</span>
      </Card>
    </section>
  </PageContainer>
</template>
