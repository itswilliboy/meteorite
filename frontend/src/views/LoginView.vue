<script setup lang="ts">
import { useRouter } from "vue-router"

import InputBar from "@/components/InputBar.vue"
import useClient from "@/composables/useClient"

const router = useRouter()
const client = useClient()

const login = async (e: Event) => {
  const formData = new FormData(e.target as HTMLFormElement)
  const auth = Object.fromEntries((formData as any).entries())

  if (!auth.username || !auth.password) return

  try {
    const token = await client.login(auth.username, auth.password)
    localStorage.setItem("token", token)
    router.push("/dash")
  } catch (err) {
    alert("Invalid username or password")
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <div class="w-full max-w-md space-y-8">
      <section>
        <img src="/logo.png" alt="logo" class="mx-auto h-12 w-12 object-contain" />

        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">Meteorite</h2>
        <p class="mt-2 text-center text-sm text-gray-600">Sign in to your account</p>
      </section>

      <form class="space-y-6" @submit.prevent="login">
        <div class="rounded-md shadow-sm">
          <InputBar label="Username" name="username" type="text" required placeholder="gopher" className="rounded-t-md" />

          <InputBar
            label="Password"
            name="password"
            type="password"
            required
            placeholder="••••••••••••"
            className="rounded-b-md" />
        </div>

        <div>
          <button
            type="submit"
            class="group relative flex w-full justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none">
            Sign in
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
