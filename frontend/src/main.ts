import "@/assets/main.css"
import "@/composables/useTheme"

import { createApp } from "vue"
import App from "./App.vue"
import router from "./router"
import { Client, HTTPException } from "./utils/client"

const app = createApp(App)
const client = new Client(router)

app.use(router)
app.provide("client", client)

router.beforeEach(async (to) => {
  if (!to.path.startsWith("/dash")) return
  try {
    await client.ping()
  } catch (e) {
    if (e instanceof HTTPException && e.status === 401) {
      return "/login"
    }
  }
})

app.mount("#app")
