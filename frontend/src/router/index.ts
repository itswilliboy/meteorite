import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router"
import HomeView from "../views/HomeView.vue"

const dashRoutes = [
  {
    path: "home",
    name: "dashHome",
    component: () => import("@/views/dash/DashHomeView.vue")
  }
] satisfies RouteRecordRaw[]

export default createRouter({
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
      children: dashRoutes
    }
  ]
})
