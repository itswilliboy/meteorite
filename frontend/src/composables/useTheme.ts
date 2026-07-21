import { ref, watch } from "vue"

export type Theme = "light" | "dark" | "system"

const STORAGE_KEY = "theme"

const media = window.matchMedia("(prefers-color-scheme: dark)")

const stored = localStorage.getItem(STORAGE_KEY) as Theme | null
const theme = ref<Theme>(stored ?? "system")

const apply = (t: Theme) => {
  const isDark = t === "dark" || (t === "system" && media.matches)
  document.documentElement.classList.toggle("dark", isDark)
}

apply(theme.value)

watch(theme, t => {
  localStorage.setItem(STORAGE_KEY, t)
  apply(t)
})

media.addEventListener("change", () => {
  if (theme.value === "system") apply("system")
})

export default function useTheme() {
  const setTheme = (t: Theme) => {
    theme.value = t
  }

  return { theme, setTheme }
}
