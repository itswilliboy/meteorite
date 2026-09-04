<script setup lang="ts">
import { ref, watch } from "vue"
import { RouterView } from "vue-router"
import ToastProvider from "@/components/toast/ToastProvider.vue"
import useAuth from "@/composables/useAuth"

const cachedViews = ["DashHomeView", "DashImageView", "DashLinkView", "DashSettingsView", "DashAdminView"]
const { user } = useAuth()

// only bump on user switch, not logout, or we remount the cached view early and refetch with a cleared session
const viewKeyId = ref(user.value?.id ?? "anon")
watch(user, u => {
  if (u) viewKeyId.value = u.id
})
</script>

<template>
  <ToastProvider>
    <RouterView v-slot="{ Component, route }">
      <KeepAlive :include="cachedViews">
        <component :is="Component" :key="`${String(route.name)}-${viewKeyId}`" />
      </KeepAlive>
    </RouterView>
  </ToastProvider>
</template>
