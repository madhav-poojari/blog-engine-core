import type { BlogService, Blog, BlogCreateInput, BlogUpdateInput } from './blogService'
import { api } from '@/lib/api'

export const apiBlogService: BlogService = {
  async getBlogs(): Promise<Blog[]> {
    return api.get<Blog[]>('/blogs')
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

  async createBlog(input: BlogCreateInput): Promise<Blog> {
    return api.post<Blog>('/blogs', input)
  },

  async updateBlog(input: BlogUpdateInput): Promise<Blog> {
    const { id, ...data } = input
    return api.put<Blog>(`/blogs/${encodeURIComponent(id)}`, data)
  },

  async deleteBlog(id: string): Promise<void> {
    return api.delete(`/blogs/${encodeURIComponent(id)}`)
  },
}
