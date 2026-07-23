import { nextTick, onBeforeUnmount, watch, type TemplateRef } from "vue"

export function useInfiniteScroll(target: TemplateRef<HTMLElement>, onIntersect: () => Promise<boolean> | boolean) {
  const observer = new IntersectionObserver(
    async entries => {
      if (!entries.some(entry => entry.isIntersecting)) return
      if (await onIntersect()) {
        await nextTick()
        refresh()
      }
    },
    { rootMargin: "200px" }
  )

  const refresh = () => {
    const el = target.value
    if (!el) return
    observer.unobserve(el)
    observer.observe(el)
  }

  watch(
    target,
    (el, _prev, onCleanup) => {
      if (!el) return
      observer.observe(el)
      onCleanup(() => observer.unobserve(el))
    },
    { immediate: true }
  )

  onBeforeUnmount(() => observer.disconnect())

  return { refresh }
}
