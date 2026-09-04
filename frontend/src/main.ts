import "@/assets/main.css"
import "@/composables/useTheme"

import { createApp } from "vue"
import App from "./App.vue"
import router from "./router"
import { Client } from "./utils/client"

const app = createApp(App)
const client = new Client(router)

app.use(router)
app.provide("client", client)

app.mount("#app")
