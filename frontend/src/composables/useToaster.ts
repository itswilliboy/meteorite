import type { Toast } from "@/utils/type"
import { reactive, readonly } from "vue"

const toasts = reactive<Toast[]>([])
let id = 0

function push(toast: Omit<Toast, "id" | "popped">) {
  const next = id++
  toasts.push({ ...toast, id: next, popped: false })
  if (toast.delay) setTimeout(() => pop(next), toast.delay)
}

function pop(id: number) {
  const toast = toasts.find(t => t.id === id)
  if (!toast || toast.popped) return
  toasts.splice(toasts.indexOf(toast), 1)
}

export default function useToaster() {
  return {
    toasts: readonly(toasts),
    push,
    pop
  }
}
