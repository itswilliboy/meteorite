import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router"
import HomeView from "../views/HomeView.vue"

const dashRoutes = [
  {
    path: "home",
    name: "dashHome",
    component: () => import("@/views/dash/DashHomeView.vue")
  },
  {
    path: "images",
    name: "dashImages",
    component: () => import("@/views/dash/DashImageView.vue")
  },
  {
    path: "links",
    name: "dashLinks",
    component: () => import("@/views/dash/DashLinkView.vue")
  },
  {
    path: "settings",
    name: "dashSettings",
    component: () => import("@/views/dash/DashSettingsView.vue")
  }
] satisfies RouteRecordRaw[]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: HomeView
    },
    {
      path: "/login",
      name: "login",
      component: () => import("@/views/LoginView.vue")
    },
    { path: "/:pathMatch(.*)", name: "NotFound", component: () => import("@/views/NotFoundView.vue") },
    {
      path: "/dash",
      redirect: "/dash/home",
      name: "dash",
      children: dashRoutes
    }
  ]
})

router.beforeEach(async (to, _) => {
  if (to.name == "home") {
    return
  }

  const token = localStorage.getItem("token")
  if (!token) {
    if (to.name != "login") {
      return { name: "login" }
    }

    return
  }

  // TDOD: check if token is valid
  if (to.name == "login") {
    return { name: "dash" }
  }

  return
})

export default router
