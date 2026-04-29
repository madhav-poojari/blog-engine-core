import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { blogService } from '@/services'
import type { BlogCreateInput, BlogUpdateInput } from '@/types/blog'

const BLOG_KEYS = {
  all: ['blogs'] as const,
  detail: (slug: string) => ['blogs', slug] as const,
}

export function useBlogs() {
  return useQuery({
    queryKey: BLOG_KEYS.all,
    queryFn: () => blogService.getBlogs(),
  })
}

export function useBlog(slug: string) {
  return useQuery({
    queryKey: BLOG_KEYS.detail(slug),
    queryFn: () => blogService.getBlogBySlug(slug),
    enabled: !!slug,
  })
}

export function useCreateBlog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: BlogCreateInput) => blogService.createBlog(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: BLOG_KEYS.all })
    },
  })
}

export function useUpdateBlog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: BlogUpdateInput) => blogService.updateBlog(input),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: BLOG_KEYS.all })
      if (variables.id) {
        queryClient.invalidateQueries({ queryKey: BLOG_KEYS.detail(variables.id) })
      }
    },
  })
}

export function useDeleteBlog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => blogService.deleteBlog(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: BLOG_KEYS.all })
    },
  })
}
