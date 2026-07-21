import { ref } from "vue"

// Bumped whenever media is uploaded/deleted so other cached (KeepAlive'd)
// views know their fetched data is stale and worth reloading on next activation.
const version = ref(0)

export const bumpMediaVersion = () => {
  version.value++
}

export default function useMediaVersion() {
  return version
}
