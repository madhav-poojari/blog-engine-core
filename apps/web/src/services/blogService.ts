import type { Blog, TagCount } from '@/types/blog'

/**
 * Read-only: posts are authored by committing markdown to the content repo,
 * so there is no create/update/delete here. The API answers 405 on those
 * routes until editing returns as a commit through the GitHub API.
 */
export interface BlogService {
  getBlogs(tag?: string): Promise<Blog[]>
  getBlogBySlug(slug: string): Promise<Blog | null>
  getTags(): Promise<TagCount[]>
}

export type { Blog, TagCount }
