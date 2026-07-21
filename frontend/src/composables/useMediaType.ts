export type MediaCategory = "image" | "video" | "audio" | "text" | "other"

export type MediaType = {
  category: MediaCategory
  isImage: boolean
  isVideo: boolean
  isAudio: boolean
  isText: boolean
  isOther: boolean
}

export function useMediaType(mimetype: string): MediaType {
  const category = (mimetype.split("/")[0] as MediaCategory) ?? "other"
  const isImage = category === "image"
  const isVideo = category === "video"
  const isAudio = category === "audio"
  const isText = category === "text"

  return {
    category,
    isImage,
    isVideo,
    isAudio,
    isText,
    isOther: !isImage && !isVideo && !isAudio && !isText,
  }
}
