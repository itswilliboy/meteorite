import { onBeforeUnmount, onMounted, ref, type Ref, type TemplateRef } from "vue"

export function useHover(template: TemplateRef<HTMLElement>): Ref<boolean> {
  const isHovered = ref<boolean>(false)

  const enter = () => (isHovered.value = true)
  const leave = () => (isHovered.value = false)

  onMounted(() => {
    if (!template.value) return
    template.value.addEventListener("mouseenter", enter)
    template.value.addEventListener("mouseleave", leave)
  })

  onBeforeUnmount(() => {
    if (!template.value) return
    template.value.addEventListener("mouseenter", enter)
    template.value.addEventListener("mouseleave", leave)
  })

  return isHovered
}
