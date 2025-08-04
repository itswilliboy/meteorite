<script setup lang="ts">
import { ref } from "vue"

import PageContainer from "@/components/PageContainer.vue"
import Card from "@/components/Card.vue"

import { Clipboard, RefreshCw } from "lucide-vue-next"

import useClient from "@/composables/useClient"
import useToaster from "@/composables/useToaster"

const client = useClient()
const { push } = useToaster()
const token = ref<Option<string>>(null)

const copyToClipboard = () => {
  if (token.value) {
    navigator.clipboard.writeText(token.value)
    push({ title: "Copied to clipboard!", colour: "info", delay: 4000 })
  }
}

const resetToken = async () => {
  const newToken = await client.resetToken()
  token.value = newToken
}
</script>

<template>
  <PageContainer title="Settings">
    <div class="w-2/5">
      <Card>
        <legend>API Key</legend>
        <div class="relative overflow-clip rounded-lg">
          <input
            type="text"
            disabled
            class="relative w-full rounded-lg border-1 border-black/50 bg-gray-100 px-2 text-xl"
            :value="token ?? '***************'" />

          <div class="absolute inset-y-0 right-0 flex h-full">
            <button class="bg-danger h-full px-3 text-white hover:cursor-pointer" @click="resetToken">
              <RefreshCw width="15" height="15" />
            </button>

            <button
              class="bg-info h-full px-3 hover:cursor-pointer disabled:cursor-not-allowed"
              :disabled="!token"
              @click="copyToClipboard">
              <Clipboard class="text-white" width="15" height="15" />
            </button>
          </div>
        </div>
      </Card>
    </div>
  </PageContainer>
</template>
