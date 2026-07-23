<script setup lang="ts">
import { ref } from "vue"
import { useRouter } from "vue-router"
import { startAuthentication } from "@simplewebauthn/browser"

import InputBar from "@/components/InputBar.vue"
import { Button } from "@/components/common"
import useClient from "@/composables/useClient"
import { Eye, EyeOff, KeyRound } from "lucide-vue-next"

const router = useRouter()
const client = useClient()

const showPassword = ref(false)
const loading = ref(false)
const passkeyLoading = ref(false)
const error = ref<Option<string>>(null)

const loginCallback = async (e: Event) => {
  const formData = new FormData(e.target as HTMLFormElement)
  const auth = Object.fromEntries((formData as any).entries())

  if (!auth.username || !auth.password) return

  error.value = null
  loading.value = true

  try {
    const user = await client.login(auth.username, auth.password)
    localStorage.setItem("user", JSON.stringify(user))
    router.push("/dash")
  } catch {
    error.value = "Invalid username or password"
  } finally {
    loading.value = false
  }
}

const passkeyLogin = async () => {
  if (passkeyLoading.value) return

  error.value = null
  passkeyLoading.value = true

  try {
    const optionsJSON = await client.webauthnLoginBegin()
    const response = await startAuthentication({ optionsJSON })
    const user = await client.webauthnLoginFinish(response)
    localStorage.setItem("user", JSON.stringify(user))
    router.push("/dash")
  } catch (e) {
    // the user simply dismissed the passkey prompt - not an error worth showing
    if ((e as Error)?.name !== "NotAllowedError") {
      error.value = "Passkey sign-in failed"
    }
  } finally {
    passkeyLoading.value = false
  }
}
</script>

<template>
  <a href="https://github.com/itswilliboy/meteorite" target="_blank" class="absolute p-3 opacity-40 cursor-pointer hover:opacity-60 transition-opacity">
<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 16 16"><!-- Icon from Charm Icons by Jay Newey - https://github.com/jaynewey/charm-icons/blob/main/LICENSE --><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"><path d="m5.75 14.25s-.5-2 .5-3c0 0-2 0-3.5-1.5s-1-4.5 0-5.5c-.5-1.5.5-2.5.5-2.5s1.5 0 2.5 1c1-.5 3.5-.5 4.5 0 1-1 2.5-1 2.5-1s1 1 .5 2.5c1 1 1.5 4 0 5.5s-3.5 1.5-3.5 1.5c1 1 .5 3 .5 3"/><path d="m5.25 13.75c-1.5.5-3-.5-3.5-1"/></g></svg>
  </a>
  <div class="bg-background flex min-h-screen items-center justify-center px-4 py-10">
    <div class="w-full max-w-sm">
      <div class="mb-6 flex flex-col items-center text-center">
        <img src="/logo.png" alt="Meteorite" class="mb-4 size-14 object-contain" />
        <h1 class="text-foreground text-2xl font-extrabold">Meteorite</h1>
        <p class="text-muted mt-1 text-sm">Sign in to your account</p>
      </div>

      <form
        class="border-border bg-surface space-y-4 rounded-2xl border p-6 shadow-lg dark:shadow-none sm:p-8"
        @submit.prevent="loginCallback">
        <InputBar
          id="username"
          label="Username"
          name="username"
          type="text"
          autocomplete="username"
          required
          placeholder="gopher" />

        <InputBar
          id="password"
          label="Password"
          name="password"
          :type="showPassword ? 'text' : 'password'"
          autocomplete="current-password"
          required
          placeholder="••••••••••••">
          <template #suffix>
            <button
              type="button"
              class="grid size-8 place-items-center rounded-md text-gray-400 transition hover:cursor-pointer hover:text-gray-600"
              :title="showPassword ? 'Hide password' : 'Show password'"
              @click="showPassword = !showPassword">
              <component :is="showPassword ? EyeOff : Eye" :size="18" />
            </button>
          </template>
        </InputBar>

        <p v-if="error" class="bg-danger/10 text-danger rounded-lg px-3 py-2 text-sm font-medium">
          {{ error }}
        </p>

        <Button type="submit" :loading="loading" class="focus:ring-primary/40 w-full focus:ring-2 focus:outline-none">
          {{ loading ? "Signing in..." : "Sign in" }}
        </Button>

        <div class="flex items-center gap-3">
          <div class="border-border h-px flex-1 border-t" />
          <span class="text-muted text-xs">or</span>
          <div class="border-border h-px flex-1 border-t" />
        </div>

        <Button
          type="button"
          variant="secondary"
          :icon="KeyRound"
          :loading="passkeyLoading"
          class="w-full"
          @click="passkeyLogin">
          {{ passkeyLoading ? "Waiting for passkey..." : "Sign in with a passkey" }}
        </Button>
      </form>
    </div>
  </div>
</template>
