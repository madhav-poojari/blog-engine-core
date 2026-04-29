import { api } from '@/lib/api'

interface UploadResult {
  url: string
  key: string
}

export async function uploadImage(file: File): Promise<string> {
  const formData = new FormData()
  formData.append('image', file)
  const result = await api.upload<UploadResult>('/images/upload', formData)
  return result.url
}
