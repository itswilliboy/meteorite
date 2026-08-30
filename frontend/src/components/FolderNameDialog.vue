<script setup lang="ts">
import { useKeydown } from "@/composables/useKeydown"
import { FolderPlusIcon, XIcon } from "lucide-vue-next"
import { nextTick, ref, useTemplateRef } from "vue"

const { title, confirmText = "Create", initialValue = "", confirmAction } = defineProps<{
  title: string
  confirmText?: string
  initialValue?: string
  confirmAction: (name: string) => unknown
}>()

const emit = defineEmits<{
  dismiss: []
}>()

const name = ref(initialValue)
const inputRef = useTemplateRef<HTMLInputElement>("input")

nextTick(() => {
  inputRef.value?.focus()
  inputRef.value?.select()
})

const submitting = ref(false)

const submit = async () => {
  const trimmed = name.value.trim()
  if (!trimmed || submitting.value) return

  submitting.value = true
  try {
    await confirmAction(trimmed)
    emit("dismiss")
  } finally {
    submitting.value = false
  }
}

useKeydown(e => {
  if (e.key === "Escape") emit("dismiss")
})
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-[999] flex items-center justify-center p-4">
      <Transition name="backdrop" appear>
        <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="$emit('dismiss')" />
      </Transition>

      <Transition name="panel" appear>
        <form @submit.prevent="submit" class="bg-surface border-border/60 relative z-50 w-105 max-w-full rounded-2xl border p-6 shadow-2xl">
          <h1 class="text-2xl font-semibold">{{ title }}</h1>

          <input
            ref="input"
            v-model="name"
            type="text"
            maxlength="255"
            placeholder="Folder name"
            class="border-border bg-surface-2 text-foreground mt-4 w-full rounded-lg border px-3 py-2.5 text-sm outline-none focus:border-primary" />

          <div
            class="mt-6 grid w-full grid-cols-2 gap-2 text-center *:flex *:cursor-pointer *:items-center *:justify-center *:gap-1.5 *:rounded-xl *:p-3 *:text-sm *:font-semibold *:transition *:hover:opacity-90">
            <button type="button" class="bg-surface-2 text-foreground" @click="$emit('dismiss')">
              <XIcon class="size-4" />
              Cancel
            </button>
            <button
              type="submit"
              class="bg-primary text-white disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="!name.trim() || submitting">
              <FolderPlusIcon class="size-4" />
              {{ submitting ? "Saving..." : confirmText }}
            </button>
          </div>
        </form>
      </Transition>
    </div>
  </Teleport>
</template>
