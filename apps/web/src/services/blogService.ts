import type { Blog, BlogCreateInput, BlogUpdateInput } from '@/types/blog'

export interface BlogService {
  getBlogs(): Promise<Blog[]>
  getBlogBySlug(slug: string): Promise<Blog | null>
  createBlog(input: BlogCreateInput): Promise<Blog>
  updateBlog(input: BlogUpdateInput): Promise<Blog>
  deleteBlog(id: string): Promise<void>
}

export type { Blog, BlogCreateInput, BlogUpdateInput }
