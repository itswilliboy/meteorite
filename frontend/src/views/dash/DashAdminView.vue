<script setup lang="ts">
import { computed, onMounted, ref } from "vue"

import PageContainer from "@/components/PageContainer.vue"
import Card from "@/components/Card.vue"
import InputBar from "@/components/InputBar.vue"
import { StatCard, PaginationControls, Button } from "@/components/common"
import ConfirmDialogue from "@/components/ConfirmDialogue.vue"
import { HTTPException } from "@/utils/client"
import useClient from "@/composables/useClient"
import useToaster from "@/composables/useToaster"
import { formatBytes } from "@/utils/format"
import type { AdminStats, AdminUser, PaginatedResponse, User } from "@/utils/type"

import { Users, UserCheck, Images, HardDrive, ShieldCheck, ShieldOff, Ban, CheckCircle2, UserPlus } from "lucide-vue-next"

defineOptions({ name: "DashAdminView" })

const client = useClient()
const { push } = useToaster()

const stats = ref<Option<AdminStats>>(null)
const response = ref<PaginatedResponse<AdminUser[]> | null>(null)
const currentUser = ref<Option<User>>(null)

onMounted(async () => {
  const stored = localStorage.getItem("user")
  if (stored) currentUser.value = JSON.parse(stored)

  const [s, users] = await Promise.all([client.adminStats(), client.adminListUsers(0)])
  stats.value = s
  response.value = users
})

const setPage = async (page: number) => {
  response.value = await client.adminListUsers(page)
}

const statCards = [
  { key: "total_users", label: "Total Users", icon: Users, format: (v: number) => v.toLocaleString("en-GB") },
  { key: "active_users", label: "Active Users", icon: UserCheck, format: (v: number) => v.toLocaleString("en-GB") },
  { key: "total_media", label: "Total Media", icon: Images, format: (v: number) => v.toLocaleString("en-GB") },
  { key: "total_storage", label: "Total Storage", icon: HardDrive, format: formatBytes }
] as const

const formatDate = (date: Date | string): string =>
  new Intl.DateTimeFormat("en-GB", { day: "numeric", month: "short", year: "numeric" }).format(new Date(date))

type PendingAction = { user: AdminUser; field: "enabled" | "admin"; next: boolean }
const pendingAction = ref<Option<PendingAction>>(null)

const requestToggle = (user: AdminUser, field: "enabled" | "admin") => {
  if (user.id === currentUser.value?.id) return
  pendingAction.value = { user, field, next: !user[field] }
}

const confirmCopy = computed(() => {
  const action = pendingAction.value
  if (!action) return null

  if (action.field === "enabled") {
    return action.next
      ? { title: "Enable this account?", description: `${action.user.name} will regain access to the dashboard and API.`, confirmText: "Enable", colour: "success" as const, icon: CheckCircle2 }
      : { title: "Disable this account?", description: `${action.user.name} will be logged out and unable to sign in or upload.`, confirmText: "Disable", colour: "danger" as const, icon: Ban }
  }

  return action.next
    ? { title: "Grant admin access?", description: `${action.user.name} will be able to manage users and view server-wide stats.`, confirmText: "Grant admin", colour: "warning" as const, icon: ShieldCheck }
    : { title: "Revoke admin access?", description: `${action.user.name} will lose admin privileges.`, confirmText: "Revoke", colour: "warning" as const, icon: ShieldOff }
})

const applyPendingAction = async () => {
  const action = pendingAction.value
  if (!action) return

  const updated =
    action.field === "enabled"
      ? await client.adminSetUserEnabled(action.user.id, action.next)
      : await client.adminSetUserAdmin(action.user.id, action.next)

  const row = response.value?.data.find(u => u.id === action.user.id)
  if (row) {
    row.enabled = updated.enabled
    row.admin = updated.admin
  }

  push({ title: `Updated ${action.user.name}`, colour: "success", delay: 4000 })
}

const showCreateUser = ref(false)
const createLoading = ref(false)
const createError = ref<Option<string>>(null)

const openCreateUser = () => {
  createError.value = null
  showCreateUser.value = true
}

const submitCreateUser = async (e: Event) => {
  const formData = new FormData(e.target as HTMLFormElement)
  const username = formData.get("username") as string
  const password = formData.get("password") as string
  const admin = formData.get("admin") === "on"

  if (!username || !password) return

  createError.value = null
  createLoading.value = true

  try {
    const user = await client.adminCreateUser(username, password, admin)
    response.value?.data.unshift({ ...user, total_images: 0, storage_usage: 0 })
    if (stats.value) stats.value.total_users += 1
    if (stats.value && user.enabled) stats.value.active_users += 1

    push({ title: `Created ${user.name}`, colour: "success", delay: 4000 })
    showCreateUser.value = false
  } catch (e) {
    createError.value = e instanceof HTTPException && e.status === 409 ? "Username already exists" : "Could not create user"
  } finally {
    createLoading.value = false
  }
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

    <Teleport to="body" v-if="showCreateUser">
      <div class="fixed inset-0 z-[999] flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="showCreateUser = false" />

        <form
          class="bg-surface border-border/60 relative z-50 w-105 max-w-full space-y-4 rounded-2xl border p-6 shadow-2xl"
          @submit.prevent="submitCreateUser">
          <h1 class="text-2xl font-semibold">Create user</h1>

          <InputBar id="new-username" label="Username" name="username" type="text" autocomplete="off" required placeholder="gopher" />

          <InputBar
            id="new-password"
            label="Password"
            name="password"
            type="password"
            autocomplete="new-password"
            required
            placeholder="••••••••••••" />

          <label class="flex items-center gap-2 text-sm font-medium">
            <input type="checkbox" name="admin" class="accent-primary size-4 rounded" />
            Grant admin access
          </label>

          <p v-if="createError" class="bg-danger/10 text-danger rounded-lg px-3 py-2 text-sm font-medium">{{ createError }}</p>

          <div class="grid grid-cols-2 gap-2">
            <Button type="button" variant="secondary" @click="showCreateUser = false">Cancel</Button>
            <Button type="submit" :loading="createLoading">{{ createLoading ? "Creating..." : "Create" }}</Button>
          </div>
        </form>
      </div>
    </Teleport>

    <!-- Stats -->
    <section class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard
        v-for="stat in statCards"
        :key="stat.key"
        :label="stat.label"
        :icon="stat.icon"
        :value="stats ? stat.format(stats[stat.key]) : null" />
    </section>

    <!-- Users -->
    <section>
      <div class="mb-3 flex items-center justify-between gap-3">
        <h2 class="text-lg font-bold">Users</h2>

        <div class="flex items-center gap-3">
          <PaginationControls
            v-if="response"
            :page="response.page"
            :has-prev="response.hasPrev"
            :has-next="response.hasNext"
            @update:page="setPage" />

          <Button :icon="UserPlus" @click="openCreateUser">Create user</Button>
        </div>
      </div>

      <Card class="overflow-x-auto p-0">
        <!-- Loading -->
        <div v-if="!response" class="space-y-2 p-5">
          <div v-for="n in 5" :key="n" class="bg-surface-2 h-10 w-full animate-pulse rounded"></div>
        </div>

        <table v-else class="w-full text-left text-sm">
          <thead>
            <tr class="border-border text-muted border-b text-xs font-semibold tracking-wide uppercase">
              <th class="px-5 py-3 font-semibold">Name</th>
              <th class="px-5 py-3 font-semibold">Joined</th>
              <th class="px-5 py-3 font-semibold">Media</th>
              <th class="px-5 py-3 font-semibold">Storage</th>
              <th class="px-5 py-3 font-semibold">Enabled</th>
              <th class="px-5 py-3 font-semibold">Admin</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="user in response.data"
              :key="user.id"
              class="border-border/60 hover:bg-surface-2/60 border-b last:border-0 hover:cursor-pointer"
              @click="$router.push({ name: 'dashAdminUser', params: { id: user.id } })">
              <td class="px-5 py-3 font-semibold">
                {{ user.name }}
                <span v-if="user.id === currentUser?.id" class="text-muted ml-1 text-xs font-normal">(you)</span>
              </td>
              <td class="text-muted px-5 py-3">{{ formatDate(user.created_at) }}</td>
              <td class="px-5 py-3 tabular-nums">{{ user.total_images.toLocaleString("en-GB") }}</td>
              <td class="px-5 py-3 tabular-nums">{{ formatBytes(user.storage_usage) }}</td>
              <td class="px-5 py-3">
                <button
                  :disabled="user.id === currentUser?.id"
                  @click.stop="requestToggle(user, 'enabled')"
                  :class="[
                    'rounded-full px-2.5 py-1 text-xs font-semibold transition disabled:cursor-not-allowed disabled:opacity-50',
                    user.enabled ? 'bg-success/10 text-success hover:opacity-80' : 'bg-danger/10 text-danger hover:opacity-80',
                    user.id !== currentUser?.id && 'hover:cursor-pointer',
                  ]">
                  {{ user.enabled ? "Enabled" : "Disabled" }}
                </button>
              </td>
              <td class="px-5 py-3">
                <button
                  :disabled="user.id === currentUser?.id"
                  @click.stop="requestToggle(user, 'admin')"
                  :class="[
                    'rounded-full px-2.5 py-1 text-xs font-semibold transition disabled:cursor-not-allowed disabled:opacity-50',
                    user.admin ? 'bg-primary/10 text-primary hover:opacity-80' : 'bg-surface-2 text-muted hover:opacity-80',
                    user.id !== currentUser?.id && 'hover:cursor-pointer',
                  ]">
                  {{ user.admin ? "Admin" : "User" }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </Card>
    </section>
  </PageContainer>
</template>
