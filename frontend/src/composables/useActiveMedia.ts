import { ref } from "vue"

const activeMedia = ref<HTMLMediaElement | null>(null)

// shared across every component instance that
// imports this composable, so at most one audio/video element plays at once
export function useActiveMedia() {
  const setActive = (el: HTMLMediaElement) => {
    if (activeMedia.value && activeMedia.value !== el && !activeMedia.value.paused) {
      activeMedia.value.pause()
    }
    activeMedia.value = el
  }

  const clearActive = (el: HTMLMediaElement) => {
    if (activeMedia.value === el) activeMedia.value = null
  }

  return { setActive, clearActive }
}
