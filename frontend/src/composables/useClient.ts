import type { Client } from "@/utils/client"
import { inject } from "vue"

export default function useClient(): Client {
  return inject("client") as Client
}
