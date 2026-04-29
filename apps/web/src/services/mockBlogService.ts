import type { BlogService, Blog, BlogCreateInput, BlogUpdateInput } from './blogService'
import { mockBlogs } from '@/data/mockBlogs'

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

let blogs = [...mockBlogs]

export const mockBlogService: BlogService = {
  async getBlogs(): Promise<Blog[]> {
    await delay(300)
    return [...blogs].sort(
      (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
    )
  },

  async getBlogBySlug(slug: string): Promise<Blog | null> {
    await delay(200)
    return blogs.find((b) => b.slug === slug) ?? null
  },

  async createBlog(input: BlogCreateInput): Promise<Blog> {
    await delay(400)
    const now = new Date().toISOString()
    const slug = input.title
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/(^-|-$)/g, '')
    const newBlog: Blog = {
      id: crypto.randomUUID(),
      slug,
      readTimeMinutes: Math.max(1, Math.ceil(JSON.stringify(input.content).length / 1000)),
      createdAt: now,
      updatedAt: now,
      ...input,
    }
    blogs = [newBlog, ...blogs]
    return newBlog
  },

  async updateBlog(input: BlogUpdateInput): Promise<Blog> {
    await delay(300)
    const index = blogs.findIndex((b) => b.id === input.id)
    if (index === -1) throw new Error('Blog not found')
    const updated: Blog = {
      ...blogs[index],
      ...input,
      updatedAt: new Date().toISOString(),
    }
    blogs = blogs.map((b) => (b.id === input.id ? updated : b))
    return updated
  },

  async deleteBlog(id: string): Promise<void> {
    await delay(200)
    blogs = blogs.filter((b) => b.id !== id)
  },
}
