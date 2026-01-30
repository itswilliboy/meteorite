import { onMounted, onBeforeUnmount, unref, type MaybeRef } from "vue"

export function useKeydown(handler: (e: KeyboardEvent) => void, enabled: MaybeRef<boolean> = true) {
  const onKeydown = (e: KeyboardEvent) => {
    if (!unref(enabled)) return
    handler(e)
  }

  onMounted(() => window.addEventListener("keydown", onKeydown))
  onBeforeUnmount(() => window.removeEventListener("keydown", onKeydown))
}
