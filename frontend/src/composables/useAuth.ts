import { ref } from "vue"
import type { User } from "@/utils/type"

const stored = localStorage.getItem("user")
export const currentUser = ref<User | null>(stored ? JSON.parse(stored) : null)

export function setCurrentUser(user: User | null) {
  currentUser.value = user
  if (user) {
    localStorage.setItem("user", JSON.stringify(user))
  } else {
    localStorage.removeItem("user")
  }
}

export default function useAuth() {
  return { user: currentUser, setUser: setCurrentUser }
}
