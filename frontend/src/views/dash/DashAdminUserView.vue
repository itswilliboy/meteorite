<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"

import PageContainer from "@/components/PageContainer.vue"
import Card from "@/components/Card.vue"
import { StatCard } from "@/components/common"
import ConfirmDialogue from "@/components/ConfirmDialogue.vue"
import useClient from "@/composables/useClient"
import useToaster from "@/composables/useToaster"
import { formatBytes } from "@/utils/format"
import type { AdminUserDetail, User } from "@/utils/type"

import {
  ArrowLeft,
  Images,
  HardDrive,
  Gauge,
  CalendarDays,
  ShieldCheck,
  ShieldOff,
  Ban,
  CheckCircle2,
  ImageIcon,
  Video,
  Music,
  File
} from "lucide-vue-next"

defineOptions({ name: "DashAdminUserView" })

const client = useClient()
const router = useRouter()
const route = useRoute()
const { push } = useToaster()

const userId = computed(() => Number(route.params.id))

const currentUser = ref<Option<User>>(null)
const detail = ref<Option<AdminUserDetail>>(null)

const load = async () => {
  detail.value = null
  detail.value = await client.adminGetUser(userId.value)
}

onMounted(async () => {
  const stored = localStorage.getItem("user")
  if (stored) currentUser.value = JSON.parse(stored)
  await load()
})

watch(userId, load)

const formatDate = (date: Date | string): string =>
  new Intl.DateTimeFormat("en-GB", { day: "numeric", month: "short", year: "numeric" }).format(new Date(date))

const statCards = computed(() => [
  { label: "Media", icon: Images, value: detail.value ? detail.value.total_images.toLocaleString("en-GB") : null },
  { label: "Storage", icon: HardDrive, value: detail.value ? formatBytes(detail.value.storage_usage) : null },
  { label: "Bandwidth", icon: Gauge, value: detail.value ? formatBytes(detail.value.bandwidth) : null },
  { label: "Joined", icon: CalendarDays, value: detail.value ? formatDate(detail.value.created_at) : null }
])

const mediaTypeCards = computed(() => [
  { label: "Images", icon: ImageIcon, value: detail.value?.media_types.images ?? 0 },
  { label: "Videos", icon: Video, value: detail.value?.media_types.videos ?? 0 },
  { label: "Audio", icon: Music, value: detail.value?.media_types.audio ?? 0 },
  { label: "Other", icon: File, value: detail.value?.media_types.other ?? 0 }
])

type PendingAction = { field: "enabled" | "admin"; next: boolean }
const pendingAction = ref<Option<PendingAction>>(null)

const isSelf = computed(() => detail.value?.id === currentUser.value?.id)

const requestToggle = (field: "enabled" | "admin") => {
  if (!detail.value || isSelf.value) return
  pendingAction.value = { field, next: !detail.value[field] }
}

const confirmCopy = computed(() => {
  const action = pendingAction.value
  if (!action || !detail.value) return null

  if (action.field === "enabled") {
    return action.next
      ? { title: "Enable this account?", description: `${detail.value.name} will regain access to the dashboard and API.`, confirmText: "Enable", colour: "success" as const, icon: CheckCircle2 }
      : { title: "Disable this account?", description: `${detail.value.name} will be logged out and unable to sign in or upload.`, confirmText: "Disable", colour: "danger" as const, icon: Ban }
  }

  return action.next
    ? { title: "Grant admin access?", description: `${detail.value.name} will be able to manage users and view server-wide stats.`, confirmText: "Grant admin", colour: "warning" as const, icon: ShieldCheck }
    : { title: "Revoke admin access?", description: `${detail.value.name} will lose admin privileges.`, confirmText: "Revoke", colour: "warning" as const, icon: ShieldOff }
})

const applyPendingAction = async () => {
  const action = pendingAction.value
  if (!action || !detail.value) return

  const updated =
    action.field === "enabled"
      ? await client.adminSetUserEnabled(detail.value.id, action.next)
      : await client.adminSetUserAdmin(detail.value.id, action.next)

  detail.value.enabled = updated.enabled
  detail.value.admin = updated.admin

  push({ title: `Updated ${detail.value.name}`, colour: "success", delay: 4000 })
}
</script>

<template>
  <PageContainer title="Server Management" className="space-y-6">
    <ConfirmDialogue
      v-if="pendingAction && confirmCopy"
      @dismiss="pendingAction = null"
      :title="confirmCopy.title"
      :description="confirmCopy.description"
      :confirm-text="confirmCopy.confirmText"
      :confirm-colour="confirmCopy.colour"
      :confirm-icon="confirmCopy.icon"
      :confirm-action="() => applyPendingAction()" />

    <button
      class="text-muted hover:text-foreground flex items-center gap-1.5 text-sm font-medium hover:cursor-pointer"
      @click="router.push({ name: 'dashAdmin' })">
      <ArrowLeft :size="16" />
      Back to users
    </button>

    <section v-if="detail" class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-xl font-bold">
          {{ detail.name }}
          <span v-if="isSelf" class="text-muted ml-1 text-sm font-normal">(you)</span>
        </h2>
      </div>

      <div class="flex items-center gap-2">
        <button
          :disabled="isSelf"
          @click="requestToggle('enabled')"
          :class="[
            'rounded-full px-3 py-1.5 text-xs font-semibold transition disabled:cursor-not-allowed disabled:opacity-50',
            detail.enabled ? 'bg-success/10 text-success hover:opacity-80' : 'bg-danger/10 text-danger hover:opacity-80',
            !isSelf && 'hover:cursor-pointer',
          ]">
          {{ detail.enabled ? "Enabled" : "Disabled" }}
        </button>

        <button
          :disabled="isSelf"
          @click="requestToggle('admin')"
          :class="[
            'rounded-full px-3 py-1.5 text-xs font-semibold transition disabled:cursor-not-allowed disabled:opacity-50',
            detail.admin ? 'bg-primary/10 text-primary hover:opacity-80' : 'bg-surface-2 text-muted hover:opacity-80',
            !isSelf && 'hover:cursor-pointer',
          ]">
          {{ detail.admin ? "Admin" : "User" }}
        </button>
      </div>
    </section>

    <!-- Stats -->
    <section class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard v-for="stat in statCards" :key="stat.label" :label="stat.label" :icon="stat.icon" :value="stat.value" />
    </section>

    <!-- Media breakdown -->
    <section>
      <h2 class="mb-3 text-lg font-bold">Media breakdown</h2>

      <Card>
        <div v-if="!detail" class="space-y-2">
          <div v-for="n in 4" :key="n" class="bg-surface-2 h-6 w-full animate-pulse rounded"></div>
        </div>

        <div v-else class="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div v-for="type in mediaTypeCards" :key="type.label" class="flex items-center gap-3">
            <div class="bg-surface-2 grid size-9 shrink-0 place-items-center rounded-lg">
              <component :is="type.icon" :size="16" class="text-primary" />
            </div>
            <div>
              <p class="text-lg font-bold leading-none tabular-nums">{{ type.value.toLocaleString("en-GB") }}</p>
              <p class="text-muted text-xs">{{ type.label }}</p>
            </div>
          </div>
        </div>
      </Card>
    </section>
  </PageContainer>
</template>
