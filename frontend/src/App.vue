<script setup lang="ts">
import { RouterView } from "vue-router"
import ToastProvider from "@/components/toast/ToastProvider.vue"
import useAuth from "@/composables/useAuth"

const cachedViews = ["DashHomeView", "DashImageView", "DashLinkView", "DashSettingsView", "DashAdminView"]
const { user } = useAuth()
</script>

<template>
  <ToastProvider>
    <RouterView v-slot="{ Component, route }">
      <KeepAlive :include="cachedViews">
        <component :is="Component" :key="`${String(route.name)}-${user?.id ?? 'anon'}`" />
      </KeepAlive>
    </RouterView>
  </ToastProvider>
</template>
