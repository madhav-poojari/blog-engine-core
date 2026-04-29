import type { BlogService } from './blogService'
import { mockBlogService } from './mockBlogService'
import { apiBlogService } from './apiBlogService'

// Use the real API service when VITE_API_URL is set, otherwise fall back to mock data.
const useMock = import.meta.env.VITE_USE_MOCK === 'true'
export const blogService: BlogService = useMock ? mockBlogService : apiBlogService
