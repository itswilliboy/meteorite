<script setup lang="ts">
import PageContainer from "@/components/PageContainer.vue"
import { LucideUpload } from "lucide-vue-next"
</script>

<template>
  <PageContainer title="Upload">
    <div class="flex w-full items-center justify-center">
      <label
        for="dropzone-file"
        class="border-border bg-surface hover:border-primary/40 hover:bg-surface-2 flex h-64 w-full cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed transition-colors">
        <div class="flex flex-col items-center justify-center pt-5 pb-6">
          <LucideUpload class="text-muted mb-4 h-10 w-10" />

          <p class="text-muted mb-2 text-sm">
            <span class="text-primary font-semibold">Click to upload</span>
            or drag and drop
          </p>

          <p class="text-muted text-xs opacity-70">SVG, PNG, JPG or GIF</p>
        </div>

        <input id="dropzone-file" type="file" class="hidden" />
      </label>
    </div>
  </PageContainer>
</template>
