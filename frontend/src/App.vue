<script setup lang="ts">
import { RouterView } from "vue-router"
import ToastProvider from "@/components/toast/ToastProvider.vue"

const cachedViews = ["DashHomeView", "DashImageView", "DashLinkView", "DashSettingsView", "DashAdminView"]
</script>

<template>
  <ToastProvider>
    <RouterView v-slot="{ Component, route }">
      <KeepAlive :include="cachedViews">
        <component :is="Component" :key="route.name" />
      </KeepAlive>
    </RouterView>
  </ToastProvider>
</template>
