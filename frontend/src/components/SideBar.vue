<script setup lang="ts">
import { LayoutDashboard, Images, Link, Upload, Server, Settings } from "lucide-vue-next"
import type { FunctionalComponent } from "vue"

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
</script>

<template>
  <div class="flex">
    <aside>
      <nav class="flex h-screen w-16 flex-col items-center justify-between bg-white py-10">
        <div class="flex flex-col gap-6">
          <router-link :to="item.to" v-for="item in upperButtons">
            <component :is="item.icon" :size="24" class="rounded-lg" />
          </router-link>
        </div>

        <div class="flex flex-col gap-6">
          <router-link :to="item.to" v-for="item in lowerButtons">
            <component :is="item.icon" :size="24" class="rounded-lg" />
          </router-link>
        </div>
      </nav>
    </aside>

    <main>
      <slot />
    </main>
  </div>
</template>

<style scoped>
.router-link-active {
  background-color: red;
}
</style>
