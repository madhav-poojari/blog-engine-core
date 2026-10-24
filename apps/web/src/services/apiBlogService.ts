import type { BlogService, Blog, TagCount } from './blogService'
import { api } from '@/lib/api'

export const apiBlogService: BlogService = {
  async getBlogs(tag?: string): Promise<Blog[]> {
    const query = tag ? `?tag=${encodeURIComponent(tag)}` : ''
    return api.get<Blog[]>(`/blogs${query}`)
  },

  async getBlogBySlug(slug: string): Promise<Blog | null> {
    try {
      return await api.get<Blog>(`/blogs/${encodeURIComponent(slug)}`)
    } catch (err) {
      if (err instanceof Error && 'status' in err && (err as { status: number }).status === 404) {
        return null
      }
      throw err
    }
  },

  async getTags(): Promise<TagCount[]> {
    return api.get<TagCount[]>('/tags')
  },
}
