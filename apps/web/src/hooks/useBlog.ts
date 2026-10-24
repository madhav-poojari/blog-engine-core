import { useQuery } from '@tanstack/react-query'
import { blogService } from '@/services'

const BLOG_KEYS = {
  all: ['blogs'] as const,
  list: (tag?: string) => ['blogs', 'list', tag ?? ''] as const,
  detail: (slug: string) => ['blogs', slug] as const,
  tags: ['tags'] as const,
}

export function useBlogs(tag?: string) {
  return useQuery({
    queryKey: BLOG_KEYS.list(tag),
    queryFn: () => blogService.getBlogs(tag),
  })
}

export function useBlog(slug: string) {
  return useQuery({
    queryKey: BLOG_KEYS.detail(slug),
    queryFn: () => blogService.getBlogBySlug(slug),
    enabled: !!slug,
  })
}

export function useTags() {
  return useQuery({
    queryKey: BLOG_KEYS.tags,
    queryFn: () => blogService.getTags(),
  })
}
