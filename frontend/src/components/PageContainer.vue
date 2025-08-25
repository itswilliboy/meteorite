<script setup lang="ts">
import clsx from "clsx"
import { LayoutDashboard, Images, Link, Upload, Server, Settings, LogOut } from "lucide-vue-next"
import type { FunctionalComponent } from "vue"

import { useRouter } from "vue-router"
import useClient from "@/composables/useClient"

const client = useClient()
const router = useRouter()

const logOut = async () => {
  await client.logout()
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
  },
  {
    name: "Upload",
    to: "/dash/upload",
    icon: Upload
  }
] satisfies ButtonT[]

const lowerButtons = [
  {
    name: "Admin Dashboard",
    to: "/admin/dash",
    icon: Server
  },
  {
    name: "Settings",
    to: "/dash/settings",
    icon: Settings
  }
] satisfies ButtonT[]

defineProps<{ title: string; className?: string }>()
</script>

<template>
  <div class="flex">
    <aside>
      <nav class="sticky top-0 left-0 flex h-screen w-16 flex-col items-center justify-between bg-white py-10">
        <div class="item">
          <router-link :to="item.to" v-for="item in upperButtons">
            <component :is="item.icon" :size="24" />
          </router-link>
        </div>

        <div class="item">
          <router-link :to="item.to" v-for="item in lowerButtons">
            <component :is="item.icon" :size="24" />
          </router-link>

          <div @click="logOut">
            <LogOut :size="24" class="rotate-180 rounded-lg" />
          </div>
        </div>
      </nav>
    </aside>

    <div class="w-full px-6 pt-10">
      <h1 class="mb-5 text-2xl font-extrabold">{{ title }}</h1>

      <main :class="clsx(className)">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
@reference "../assets/main.css";

.router-link-active {
  @apply !rounded-xl !bg-black/10;
}

.item {
  @apply flex flex-col;

  a,
  div {
    @apply m-1 flex items-center justify-center rounded-xl p-2;

    &:hover {
      @apply cursor-pointer bg-black/3;
    }
  }
}
</style>
