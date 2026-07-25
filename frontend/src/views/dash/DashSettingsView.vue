<script setup lang="ts">
import { onMounted, ref } from "vue"
import { startRegistration } from "@simplewebauthn/browser"

import PageContainer from "@/components/PageContainer.vue"
import ConfirmDialogue from "@/components/ConfirmDialogue.vue"
import InputBar from "@/components/InputBar.vue"
import SettingsCard from "@/components/settings/SettingsCard.vue"
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
  Trash2Icon,
  Asterisk
} from "lucide-vue-next"

import useClient from "@/composables/useClient"
import { HTTPException } from "@/utils/client"
import useToaster from "@/composables/useToaster"
import useTheme, { type Theme } from "@/composables/useTheme"
import useAuth from "@/composables/useAuth"
import type { Passkey } from "@/utils/type"

defineOptions({ name: "DashSettingsView" })

const client = useClient()
const { push } = useToaster()
const { theme, setTheme } = useTheme()
const { user } = useAuth()

const token = ref<Option<string>>(null)
const revealed = ref(false)
const resetting = ref(false)

const themeOptions = [
  { value: "light", label: "Light", icon: Sun },
  { value: "dark", label: "Dark", icon: Moon },
  { value: "system", label: "System", icon: Monitor }
] satisfies { value: Theme; label: string; icon: unknown }[]

const passkeys = ref<Option<Passkey[]>>(null)
const newPasskeyName = ref("")
const addingPasskey = ref(false)
const deleteTarget = ref<Option<Passkey>>(null)

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

const changingPassword = ref(false)
const showOldPassword = ref(false)
const showNewPassword = ref(false)
const passwordError = ref<Option<string>>(null)

const submitPasswordChange = async (e: Event) => {
  const form = e.target as HTMLFormElement
  const formData = new FormData(form)
  const oldPassword = formData.get("old-password") as string
  const newPassword = formData.get("new-password") as string
  const confirmPassword = formData.get("confirm-password") as string

  passwordError.value = null

  if (newPassword !== confirmPassword) {
    passwordError.value = "New passwords do not match."
    return
  }

  changingPassword.value = true
  try {
    await client.changePassword(oldPassword, newPassword)
    form.reset()

    // the old key was invalidated server-side, drop it from view
    token.value = null
    revealed.value = false

    push({
      title: "Password changed",
      desc: "Your other sessions and API key were invalidated.",
      colour: "success",
      delay: 6000
    })
  } catch (e) {
    if (e instanceof HTTPException && e.status === 403) {
      passwordError.value = "Old password does not match."
    } else {
      push({ title: "Could not change password", desc: "Something went wrong, please try again.", colour: "danger", delay: 6000 })
    }
  } finally {
    changingPassword.value = false
  }
}

const formatDate = (date: Date | string): string =>
  new Intl.DateTimeFormat("en-GB", { day: "numeric", month: "long", year: "numeric" }).format(new Date(date))
</script>

<template>
  <PageContainer title="Settings" className="max-w-2xl space-y-6">
    <!-- Account -->
    <SettingsCard :icon="UserIcon" title="Account">
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
    </SettingsCard>

    <SettingsCard
      :icon="Asterisk"
      title="Change Password"
      description="Change your password. Warning: this will invalidate all other sessions and reset your API key.">
      <form class="border-border space-y-3 border-t pt-4" @submit.prevent="submitPasswordChange">
        <InputBar
          id="old-password"
          label="Current password"
          name="old-password"
          :type="showOldPassword ? 'text' : 'password'"
          autocomplete="current-password"
          required
          placeholder="••••••••••••">
          <template #suffix>
            <button
              type="button"
              class="grid size-8 place-items-center rounded-md text-gray-400 transition hover:cursor-pointer hover:text-gray-600"
              :title="showOldPassword ? 'Hide password' : 'Show password'"
              @click="showOldPassword = !showOldPassword">
              <component :is="showOldPassword ? EyeOff : Eye" :size="18" />
            </button>
          </template>
        </InputBar>

        <InputBar
          id="new-password"
          label="New password"
          name="new-password"
          :type="showNewPassword ? 'text' : 'password'"
          autocomplete="new-password"
          required
          minlength="8"
          placeholder="••••••••••••">
          <template #suffix>
            <button
              type="button"
              class="grid size-8 place-items-center rounded-md text-gray-400 transition hover:cursor-pointer hover:text-gray-600"
              :title="showNewPassword ? 'Hide password' : 'Show password'"
              @click="showNewPassword = !showNewPassword">
              <component :is="showNewPassword ? EyeOff : Eye" :size="18" />
            </button>
          </template>
        </InputBar>

        <InputBar
          id="confirm-password"
          label="Confirm new password"
          name="confirm-password"
          :type="showNewPassword ? 'text' : 'password'"
          autocomplete="new-password"
          required
          minlength="8"
          placeholder="••••••••••••" />

        <p v-if="passwordError" class="bg-danger/10 text-danger rounded-lg px-3 py-2 text-sm font-medium">
          {{ passwordError }}
        </p>

        <Button type="submit" variant="danger" :icon="Asterisk" :loading="changingPassword" class="w-full sm:w-auto">
          {{ changingPassword ? "Changing..." : "Change password" }}
        </Button>
      </form>
    </SettingsCard>

    <!-- Passkeys -->
    <SettingsCard :icon="Fingerprint" title="Passkeys" description="Sign in using a security key instead of your password.">
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
            <p class="text-muted text-xs">{{ pk.id }}</p>
            <p class="text-muted text-xs">Added {{ formatDate(pk.created_at) }}</p>
          </div>

          <button
            class="text-muted hover:text-danger shrink-0 hover:cursor-pointer"
            title="Remove"
            @click="deleteTarget = pk">
            <Trash2Icon :size="16" />
          </button>
        </div>
      </div>

      <div class="border-border flex flex-col gap-3 border-t pt-4 sm:flex-row sm:items-center">
        <input
          v-model="newPasskeyName"
          type="text"
          placeholder="e.g. MacBook"
          class="border-border bg-surface-2 text-foreground focus:border-primary focus:ring-primary/20 w-full rounded-lg border px-3 py-2 text-sm outline-none focus:ring-2 sm:flex-1" />

        <Button :icon="Plus" :loading="addingPasskey" class="w-full sm:w-auto" @click="addPasskey">
          {{ addingPasskey ? "Waiting for passkey..." : "Add a passkey" }}
        </Button>
      </div>
    </SettingsCard>

    <!-- Appearance -->
    <SettingsCard
      :icon="Palette"
      title="Appearance"
      description="Choose how the dashboard looks. System follows your device settings.">
      <div class="bg-surface-2 grid grid-cols-3 gap-1 rounded-lg p-1">
        <button
          v-for="opt in themeOptions"
          :key="opt.value"
          class="flex items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition hover:cursor-pointer"
          :class="theme === opt.value ? 'bg-primary text-white shadow-sm' : 'text-muted hover:text-foreground'"
          @click="setTheme(opt.value)">
          <component :is="opt.icon" :size="16" />
          {{ opt.label }}
        </button>
      </div>
    </SettingsCard>

    <!-- API Key -->
    <SettingsCard
      :icon="KeyRound"
      title="API Key"
      description="Use this key to authenticate uploads from external clients. Keep it secret — anyone with it can upload on your behalf.">
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

      <div
        class="border-border flex flex-col items-start gap-3 border-t pt-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p class="text-sm font-semibold">Regenerate key</p>
          <p class="text-muted text-xs">This would invalidate your current key.</p>
        </div>

        <Button variant="danger" :icon="RefreshCw" :loading="resetting" class="w-full sm:w-auto" @click="resetTokenCallback">
          {{ resetting ? "Regenerating..." : "Regenerate" }}
        </Button>
      </div>
    </SettingsCard>
  </PageContainer>
</template>
