<script setup lang="ts">
import PageContainer from "@/components/PageContainer.vue"
const token = localStorage.getItem("token")
</script>

<template>
  <PageContainer title="Settings">
    Your token:
    <br />
    <textarea readonly class="h-20 w-full">{{ token }}</textarea>
  </PageContainer>
</template>
