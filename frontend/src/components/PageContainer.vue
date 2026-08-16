<script setup lang="ts">
import clsx from "clsx"
import { LayoutDashboard, Images, Link, Server, Settings, LogOut, Menu, X } from "lucide-vue-next"
import type { FunctionalComponent } from "vue"
import { computed, ref } from "vue"

import { useRouter } from "vue-router"
import useClient from "@/composables/useClient"
import useAuth from "@/composables/useAuth"

const client = useClient()
const router = useRouter()

const { user: currentUser, setUser } = useAuth()

const logOut = async () => {
  await client.logout()
  setUser(null)
  router.push("/login")
}

type ButtonT = { name: string; to: string; icon: FunctionalComponent }

const upperButtons = [
  {
    name: "Home",
    to: "/dash/home",
    icon: LayoutDashboard
  },
  {
    name: "Images",
    to: "/dash/images",
    icon: Images
  },
  {
    name: "Links",
    to: "/dash/links",
    icon: Link
  }
] satisfies ButtonT[]

const allLowerButtons = [
  {
    name: "Admin Dashboard",
    to: "/dash/admin",
    icon: Server,
    adminOnly: true
  },
  {
    name: "Settings",
    to: "/dash/settings",
    icon: Settings,
    adminOnly: false
  }
]

const lowerButtons = computed(() => allLowerButtons.filter(b => !b.adminOnly || currentUser.value?.admin))

const mobileNavOpen = ref(false)

defineProps<{ title: string; className?: string }>()
</script>

<template>
  <div class="flex">
    <div
      v-if="mobileNavOpen"
      @click="mobileNavOpen = false"
      class="fixed inset-0 z-40 bg-black/50 sm:hidden" />

    <aside>
      <nav
        :class="[
          'border-border bg-surface fixed top-0 left-0 z-50 flex h-screen w-16 flex-col items-center justify-between border-r py-10 transition-transform duration-200 sm:sticky sm:translate-x-0',
          mobileNavOpen ? 'translate-x-0' : '-translate-x-full'
        ]">
        <div class="item">
          <router-link :to="item.to" v-for="item in upperButtons" :key="item.to" @click="mobileNavOpen = false">
            <component :is="item.icon" :size="24" />
          </router-link>
        </div>

        <div class="item">
          <router-link :to="item.to" v-for="item in lowerButtons" :key="item.to" @click="mobileNavOpen = false">
            <component :is="item.icon" :size="24" />
          </router-link>

          <div @click="logOut">
            <LogOut :size="24" class="rotate-180 rounded-lg" />
          </div>
        </div>
      </nav>
    </aside>

    <div class="w-full min-w-0 px-4 pt-6 sm:px-6 sm:pt-10">
      <div class="mb-4 flex items-center gap-3 sm:mb-5">
        <button
          @click="mobileNavOpen = !mobileNavOpen"
          class="border-border bg-surface text-muted hover:bg-surface-2 hover:text-foreground flex items-center justify-center rounded-lg border p-2 hover:cursor-pointer sm:hidden">
          <X v-if="mobileNavOpen" :size="20" />
          <Menu v-else :size="20" />
        </button>
        <h1 class="text-xl font-extrabold sm:text-2xl">{{ title }}</h1>
      </div>

      <main :class="clsx(className)">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
@reference "../assets/main.css";

.router-link-active {
  @apply rounded-xl! bg-black/10! dark:bg-white/10!;
}

.item {
  @apply flex flex-col;

  a,
  div {
    @apply m-1 flex items-center justify-center rounded-xl p-2;

    &:hover {
      @apply cursor-pointer bg-black/3 dark:bg-white/6;
    }
  }
}
</style>
