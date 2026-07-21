<script setup lang="ts">
import { computed, onActivated, onMounted, ref } from "vue"

import PageContainer from "@/components/PageContainer.vue"
import Card from "@/components/Card.vue"
import { StatCard } from "@/components/common"
import ImageOrVideo from "@/components/ImageOrVideo.vue"
import TrendChart from "@/components/TrendChart.vue"
import useClient from "@/composables/useClient"
import useMediaVersion from "@/composables/useMediaVersion"
import { formatBytes } from "@/utils/format"

import { Images, Upload, HardDrive, Gauge, ArrowUpRight, ImageOff } from "lucide-vue-next"
import { RouterLink } from "vue-router"
import type { DashboardStats, DashboardTimeseries, Image } from "@/utils/type"

defineOptions({ name: "DashHomeView" })

const client = useClient()
const mediaVersion = useMediaVersion()

const data = ref<Option<DashboardStats>>(null)
const recent = ref<Option<Image[]>>(null)
const timeseries = ref<Option<DashboardTimeseries>>(null)

let loadedVersion = -1
const loadDashboard = async () => {
  const [stats, images, ts] = await Promise.all([client.dashboardStats(), client.getImages(0), client.dashboardTimeseries()])
  data.value = stats
  recent.value = images.data.slice(0, 6)
  timeseries.value = ts
  loadedVersion = mediaVersion.value
}

onMounted(loadDashboard)

// Only refetch on revisit if media was actually uploaded/deleted elsewhere
// since we last loaded — otherwise this stays a no-op, so switching back and
// forth between pages doesn't keep re-requesting the same data.
onActivated(() => {
  if (mediaVersion.value !== loadedVersion) loadDashboard()
})

const formatBytesToMebibytes = (bytes: number): string => {
  const ONE_MEBIBYTE = 1_048_576
  return new Intl.NumberFormat("en-GB").format(Math.round(bytes / ONE_MEBIBYTE)) + " MiB"
}

const uploadLabels = computed(() => timeseries.value?.days.map(d => d.date) ?? [])
const uploadValues = computed(() => timeseries.value?.days.map(d => d.uploads) ?? [])

const storageValues = computed(() => {
  const ts = timeseries.value
  if (!ts) return []
  let running = ts.baseline_bytes
  return ts.days.map(d => {
    running += d.bytes
    return running
  })
})

const stats = [
  { key: "total_images", label: "Total Images", icon: Images, format: (v: number) => v.toLocaleString("en-GB") },
  { key: "monthly_uploads", label: "Monthly Uploads", icon: Upload, format: (v: number) => v.toLocaleString("en-GB") },
  { key: "storage_usage", label: "Storage Used", icon: HardDrive, format: formatBytesToMebibytes },
  { key: "user_bandwidth", label: "Bandwidth", icon: Gauge, format: formatBytesToMebibytes }
] as const
</script>

<template>
  <PageContainer title="Dashboard" className="space-y-6">
    <!-- Stats -->
    <section class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard
        v-for="stat in stats"
        :key="stat.key"
        :label="stat.label"
        :icon="stat.icon"
        :value="data ? stat.format(data[stat.key]) : null" />
    </section>

    <!-- Trends -->
    <section class="grid gap-4 lg:grid-cols-2">
      <Card>
        <h2 class="mb-4 text-sm font-semibold">Uploads (last 30 days)</h2>
        <TrendChart v-if="timeseries" :labels="uploadLabels" :values="uploadValues" variant="bar" />
        <div v-else class="bg-surface-2 h-40 w-full animate-pulse rounded"></div>
      </Card>

      <Card>
        <h2 class="mb-4 text-sm font-semibold">Storage growth (last 30 days)</h2>
        <TrendChart v-if="timeseries" :labels="uploadLabels" :values="storageValues" variant="area" :format-value="formatBytes" />
        <div v-else class="bg-surface-2 h-40 w-full animate-pulse rounded"></div>
      </Card>
    </section>

    <!-- Recent uploads -->
    <section>
      <div class="mb-3 flex items-center justify-between">
        <h2 class="text-lg font-bold">Recent Uploads</h2>
        <RouterLink
          to="/dash/images"
          class="text-primary flex items-center gap-1 text-sm font-semibold hover:underline">
          View all
          <ArrowUpRight :size="15" />
        </RouterLink>
      </div>

      <Card>
        <!-- Loading -->
        <div v-if="!recent" class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
          <div v-for="n in 6" :key="n" class="bg-surface-2 aspect-square animate-pulse rounded-lg"></div>
        </div>

        <!-- Empty -->
        <div v-else-if="recent.length === 0" class="text-muted flex flex-col items-center gap-2 py-12">
          <ImageOff :size="32" />
          <p class="text-sm">No uploads yet</p>
          <RouterLink to="/dash/images" class="text-primary text-sm font-semibold hover:underline">
            Upload your first file
          </RouterLink>
        </div>

        <!-- Grid -->
        <div v-else class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
          <a
            v-for="image in recent"
            :key="image.id"
            :href="image.url"
            target="_blank"
            class="group hover:ring-primary/40 border-border bg-surface-2 relative block aspect-square overflow-hidden rounded-lg ring-1 ring-black/5 transition dark:ring-white/5">
            <ImageOrVideo :image="image" :preview="true" :thumbnail="true" class="[&>*]:!size-full [&>*]:!rounded-none" />

            <!-- Hover overlay -->
            <div
              class="pointer-events-none absolute inset-x-0 bottom-0 flex items-center justify-between gap-1 bg-gradient-to-t from-black/60 to-transparent p-2 pt-6 opacity-0 transition group-hover:opacity-100">
              <span class="truncate text-xs font-medium text-white">{{ image.id }}</span>
              <span class="shrink-0 text-xs text-white/80">{{ image.views }} views</span>
            </div>
          </a>
        </div>
      </Card>
    </section>
  </PageContainer>
</template>
