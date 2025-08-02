import "@/assets/main.css"

import { createApp } from "vue"
import App from "./App.vue"
import router from "./router"
import { Client } from "./utils/client"

const app = createApp(App)
app.use(router)
app.provide("client", new Client(router))

app.mount("#app")
