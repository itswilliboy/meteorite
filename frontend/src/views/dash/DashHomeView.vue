<script setup lang="ts">
import PageContainer from "@/components/PageContainer.vue"
import Card from "@/components/Card.vue"
</script>

<template>
  <PageContainer title="Dashboard" className="space-y-3">
    <section class="flex w-full gap-3">
      <Card class="leading-none">
        <span class="text-sm font-semibold text-gray-500">Total Images</span>
        <h1 class="text-2xl font-bold">2,847</h1>
      </Card>

      <Card class="leading-none">
        <span class="text-xs font-semibold text-gray-500">Storage Used</span>
        <h1 class="text-2xl font-bold">4.2 GB</h1>
      </Card>

      <Card class="leading-none">
        <span class="text-xs font-semibold text-gray-500">Monthly Uploads</span>
        <h1 class="text-2xl font-bold">342</h1>
      </Card>

      <Card class="leading-none">
        <span class="text-xs font-semibold text-gray-500">Bandwidth</span>
        <h1 class="text-2xl font-bold">1.2 TB</h1>
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
