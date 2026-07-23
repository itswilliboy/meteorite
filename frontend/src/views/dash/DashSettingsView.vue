<script setup lang="ts">
import { onMounted, ref } from "vue"
import { startRegistration } from "@simplewebauthn/browser"

import PageContainer from "@/components/PageContainer.vue"
import Card from "@/components/Card.vue"
import ConfirmDialogue from "@/components/ConfirmDialogue.vue"
import { Button } from "@/components/common"

import {
  Clipboard,
  RefreshCw,
  Eye,
  EyeOff,
  KeyRound,
  ShieldCheck,
  User as UserIcon,
  Calendar,
  Palette,
  Sun,
  Moon,
  Monitor,
  Fingerprint,
  Plus,
  Trash2Icon
} from "lucide-vue-next"

import useClient from "@/composables/useClient"
import useToaster from "@/composables/useToaster"
import useTheme, { type Theme } from "@/composables/useTheme"
import type { Passkey, User } from "@/utils/type"

defineOptions({ name: "DashSettingsView" })

const client = useClient()
const { push } = useToaster()
const { theme, setTheme } = useTheme()

const token = ref<Option<string>>(null)
const revealed = ref(false)
const resetting = ref(false)
const user = ref<Option<User>>(null)

const themeOptions = [
  { value: "light", label: "Light", icon: Sun },
  { value: "dark", label: "Dark", icon: Moon },
  { value: "system", label: "System", icon: Monitor }
] satisfies { value: Theme; label: string; icon: unknown }[]

const passkeys = ref<Option<Passkey[]>>(null)
const newPasskeyName = ref("")
const addingPasskey = ref(false)
const deleteTarget = ref<Option<Passkey>>(null)

onMounted(() => {
  const stored = localStorage.getItem("user")
  if (stored) user.value = JSON.parse(stored)
})

const loadPasskeys = async () => {
  passkeys.value = await client.webauthnListCredentials()
}

onMounted(loadPasskeys)

const addPasskey = async () => {
  if (addingPasskey.value) return

  addingPasskey.value = true
  try {
    const optionsJSON = await client.webauthnRegisterBegin()
    const response = await startRegistration({ optionsJSON })
    await client.webauthnRegisterFinish(response, newPasskeyName.value || "Passkey")

    newPasskeyName.value = ""
    await loadPasskeys()
    push({ title: "Passkey added", colour: "success", delay: 4000 })
  } catch (e) {
    if ((e as Error)?.name !== "NotAllowedError") {
      push({ title: "Could not add passkey", colour: "danger", delay: 4000 })
    }
  } finally {
    addingPasskey.value = false
  }
}

const deletePasskey = async () => {
  if (!deleteTarget.value) return

  await client.webauthnDeleteCredential(deleteTarget.value.id)
  await loadPasskeys()
  push({ title: "Passkey removed", colour: "info", delay: 4000 })
}

const copyToClipboard = () => {
  if (token.value) {
    navigator.clipboard.writeText(token.value)
    push({ title: "Copied to clipboard!", colour: "info", delay: 4000 })
  }
}

const resetTokenCallback = async () => {
  if (resetting.value) return

  resetting.value = true
  try {
    token.value = await client.resetToken()
    revealed.value = true
    push({ title: "API key regenerated", colour: "success", delay: 4000 })
  } finally {
    resetting.value = false
  }
}

const formatDate = (date: Date | string): string =>
  new Intl.DateTimeFormat("en-GB", { day: "numeric", month: "long", year: "numeric" }).format(new Date(date))
</script>

<template>
  <PageContainer title="Settings" className="max-w-2xl space-y-6">
    <!-- Account -->
    <Card class="space-y-4">
      <div class="flex items-center gap-2">
        <UserIcon :size="18" class="text-primary" />
        <h2 class="text-lg font-bold">Account</h2>
      </div>

      <div class="border-border flex items-center gap-4 border-t pt-4">
        <div class="bg-primary flex size-12 shrink-0 items-center justify-center rounded-full text-lg font-bold text-white">
          {{ user?.name?.charAt(0).toUpperCase() ?? "?" }}
        </div>

        <div class="min-w-0 leading-tight">
          <p class="truncate font-semibold">{{ user?.name ?? "Loading..." }}</p>
          <p class="text-muted flex items-center gap-1 text-sm">
            <Calendar :size="13" />
            <span v-if="user">Member since {{ formatDate(user.created_at) }}</span>
          </p>
        </div>

        <span
          v-if="user?.admin"
          class="bg-primary/10 text-primary ml-auto flex items-center gap-1 rounded-full px-3 py-1 text-xs font-semibold">
          <ShieldCheck :size="13" />
          Admin
        </span>
      </div>
    </Card>

    <!-- Passkeys -->
    <Card class="space-y-4">
      <div class="flex items-center gap-2">
        <Fingerprint :size="18" class="text-primary" />
        <h2 class="text-lg font-bold">Passkeys</h2>
      </div>

      <p class="text-muted text-sm">
        Sign in with Face ID, Touch ID, Windows Hello, or a security key instead of your password.
      </p>

      <ConfirmDialogue
        v-if="deleteTarget"
        @dismiss="deleteTarget = null"
        title="Remove this passkey?"
        :description="`You won't be able to sign in with '${deleteTarget.name}' anymore.`"
        confirm-text="Remove"
        confirm-colour="danger"
        :confirm-icon="Trash2Icon"
        :confirm-action="() => deletePasskey()" />

      <div v-if="passkeys && passkeys.length > 0" class="border-border divide-border border-t *:border-b *:last:border-0">
        <div v-for="pk in passkeys" :key="pk.id" class="flex items-center justify-between gap-3 py-3">
          <div class="min-w-0">
            <p class="truncate text-sm font-semibold">{{ pk.name }}</p>
            <p class="text-muted text-xs">Added {{ formatDate(pk.created_at) }}</p>
          </div>

          <button class="text-muted hover:text-danger shrink-0 hover:cursor-pointer" title="Remove" @click="deleteTarget = pk">
            <Trash2Icon :size="16" />
          </button>
        </div>
      </div>

      <div class="border-border flex flex-col gap-3 border-t pt-4 sm:flex-row sm:items-center">
        <input
          v-model="newPasskeyName"
          type="text"
          placeholder="e.g. MacBook Touch ID"
          class="border-border bg-surface-2 text-foreground focus:border-primary focus:ring-primary/20 w-full rounded-lg border px-3 py-2 text-sm outline-none focus:ring-2 sm:flex-1" />

        <Button :icon="Plus" :loading="addingPasskey" class="w-full sm:w-auto" @click="addPasskey">
          {{ addingPasskey ? "Waiting for passkey..." : "Add a passkey" }}
        </Button>
      </div>
    </Card>

    <!-- Appearance -->
    <Card class="space-y-4">
      <div class="flex items-center gap-2">
        <Palette :size="18" class="text-primary" />
        <h2 class="text-lg font-bold">Appearance</h2>
      </div>

      <p class="text-muted text-sm">Choose how the dashboard looks. System follows your device settings.</p>

      <div class="bg-surface-2 grid grid-cols-3 gap-1 rounded-lg p-1">
        <button
          v-for="opt in themeOptions"
          :key="opt.value"
          class="flex items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition hover:cursor-pointer"
          :class="
            theme === opt.value ? 'bg-primary text-white shadow-sm' : 'text-muted hover:text-foreground'
          "
          @click="setTheme(opt.value)">
          <component :is="opt.icon" :size="16" />
          {{ opt.label }}
        </button>
      </div>
    </Card>

    <!-- API Key -->
    <Card class="space-y-4">
      <div class="flex items-center gap-2">
        <KeyRound :size="18" class="text-primary" />
        <h2 class="text-lg leading-none font-bold">API Key</h2>
      </div>

      <p class="text-muted text-sm">
        Use this key to authenticate uploads from external clients. Keep it secret — anyone with it can upload on your
        behalf.
      </p>

      <div class="border-border bg-surface-2 flex items-center overflow-clip rounded-lg border">
        <input
          :type="revealed && token ? 'text' : 'password'"
          readonly
          class="text-foreground w-full bg-transparent px-3 py-2.5 font-mono text-sm outline-none"
          :value="token ?? '••••••••••••••••••••••••'" />

        <button
          v-if="token"
          class="text-muted hover:text-foreground flex h-10 items-center px-3 hover:cursor-pointer"
          :title="revealed ? 'Hide' : 'Reveal'"
          @click="revealed = !revealed">
          <component :is="revealed ? EyeOff : Eye" :size="16" />
        </button>

        <button
          class="bg-primary flex h-10 items-center px-3.5 text-white transition hover:cursor-pointer hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
          title="Copy to clipboard"
          :disabled="!token"
          @click="copyToClipboard">
          <Clipboard :size="16" />
        </button>
      </div>

      <p v-if="!token" class="text-muted text-xs">
        For security, existing keys can't be displayed. Regenerate to view and copy a new one.
      </p>

      <div class="border-border flex flex-col items-start gap-3 border-t pt-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p class="text-sm font-semibold">Regenerate key</p>
          <p class="text-muted text-xs">This would invalidate your current key.</p>
        </div>

        <Button variant="danger" :icon="RefreshCw" :loading="resetting" class="w-full sm:w-auto" @click="resetTokenCallback">
          {{ resetting ? "Regenerating..." : "Regenerate" }}
        </Button>
      </div>
    </Card>
  </PageContainer>
</template>
