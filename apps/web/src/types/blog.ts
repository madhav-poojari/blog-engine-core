export interface Blog {
  id: string
  slug: string
  title: string
  excerpt: string
  /**
   * Post body, already rendered to HTML by the API. Markdown is converted
   * server-side at index time so rendering has one source of truth and the
   * client does not ship a markdown parser.
   */
  contentHtml: string
  coverImage?: string
  tags: string[]
  visibility: 'public' | 'private'
  readTimeMinutes: number
  createdAt: string
  updatedAt: string
  /** Path in the content repo. Only present for authenticated admins. */
  path?: string
}

export interface TagCount {
  tag: string
  count: number
}
