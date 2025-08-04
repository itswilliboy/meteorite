<script setup lang="ts">
import { RouterView } from "vue-router"
import ToastProvider from "@/components/toast/ToastProvider.vue"
</script>

<template>
  <ToastProvider>
    <RouterView />
  </ToastProvider>
</template>
