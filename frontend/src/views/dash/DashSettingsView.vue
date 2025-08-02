<script setup lang="ts">
import { ref } from "vue"

import PageContainer from "@/components/PageContainer.vue"
import Card from "@/components/Card.vue"

import { Clipboard, RefreshCw } from "lucide-vue-next"

import useClient from "@/composables/useClient"

const client = useClient()
const token = ref<Option<string>>(null)

const copyToClipboard = () => {
  if (token.value) {
    navigator.clipboard.writeText(token.value)
    alert("Copied!")
  }
}

const resetToken = async () => {
  const newToken = await client.resetToken()
  token.value = newToken
}
</script>

<template>
  <PageContainer title="Settings">
    <Card>
      <legend>API Key</legend>
      <div class="relative overflow-clip rounded-lg">
        <input
          type="text"
          disabled
          class="text-1xl relative w-full rounded-lg border-1 border-black bg-gray-100 px-2"
          :value="token ?? '***************'" />

        <div class="absolute inset-y-0 right-0 flex h-full">
          <button class="h-full bg-red-500 px-3 text-white hover:cursor-pointer" @click="resetToken">
            <RefreshCw width="15" height="15" />
          </button>

          <button
            class="h-full bg-gray-500 px-3 hover:cursor-pointer disabled:cursor-not-allowed"
            :disabled="!token"
            @click="copyToClipboard">
            <Clipboard class="text-white" width="15" height="15" />
          </button>
        </div>
      </div>
    </Card>
  </PageContainer>
</template>
