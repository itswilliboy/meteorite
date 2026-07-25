<script setup lang="ts">
import clsx from "clsx"
import { LayoutDashboard, Images, Link, Server, Settings, LogOut } from "lucide-vue-next"
import type { FunctionalComponent } from "vue"
import { computed } from "vue"

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

defineProps<{ title: string; className?: string }>()
</script>

<template>
  <div class="flex">
    <aside>
      <nav
        class="border-border bg-surface sticky top-0 left-0 flex h-screen w-16 flex-col items-center justify-between border-r py-10">
        <div class="item">
          <router-link :to="item.to" v-for="item in upperButtons" :key="item.to">
            <component :is="item.icon" :size="24" />
          </router-link>
        </div>

        <div class="item">
          <router-link :to="item.to" v-for="item in lowerButtons" :key="item.to">
            <component :is="item.icon" :size="24" />
          </router-link>

          <div @click="logOut">
            <LogOut :size="24" class="rotate-180 rounded-lg" />
          </div>
        </div>
      </nav>
    </aside>

    <div class="w-full min-w-0 px-4 pt-6 sm:px-6 sm:pt-10">
      <h1 class="mb-4 text-xl font-extrabold sm:mb-5 sm:text-2xl">{{ title }}</h1>

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
