import type { JSONContent } from '@tiptap/react'

export interface Blog {
  id: string
  slug: string
  title: string
  excerpt: string
  content: JSONContent
  coverImage?: string
  tags: string[]
  visibility: 'public' | 'private'
  readTimeMinutes: number
  createdAt: string
  updatedAt: string
}

export interface BlogCreateInput {
  title: string
  content: JSONContent
  excerpt: string
  coverImage?: string
  tags: string[]
  visibility: 'public' | 'private'
}

export interface BlogUpdateInput extends Partial<BlogCreateInput> {
  id: string
}
