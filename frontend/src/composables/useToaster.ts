import type { Toast } from "@/utils/type"
import { reactive, readonly } from "vue"

const toasts = reactive<Toast[]>([])
let id = 0

function push(toast: Omit<Toast, "id">) {
  const next = id++
  toasts.push({ ...toast, id: next })
  if (toast.delay) setTimeout(() => pop(next), toast.delay)
}

function pop(id: number) {
  const index = toasts.findIndex(t => t.id === id)
  toasts.splice(index, 1)
}

export default function useToaster() {
  return {
    toasts: readonly(toasts),
    push,
    pop
  }
}
