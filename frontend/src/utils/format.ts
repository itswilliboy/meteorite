export const formatBytes = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`

  const units = ["KB", "MB", "GB", "TB"]
  let value = bytes
  let unit = -1

  do {
    value /= 1024
    unit++
  } while (value >= 1024 && unit < units.length - 1)

  return `${value.toFixed(value >= 10 || unit === 0 ? 0 : 1)} ${units[unit]}`
}

export const formatDuration = (ms: number): string => {
  const totalSeconds = Math.round(ms / 1000)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (hours > 0) return `${hours}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`
  return `${minutes}:${String(seconds).padStart(2, "0")}`
}

export const formatBitrate = (bitsPerSecond: number): string => `${Math.round(bitsPerSecond / 1000)} kbps`
