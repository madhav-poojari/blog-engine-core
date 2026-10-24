import type { BlogService } from './blogService'
import { apiBlogService } from './apiBlogService'

// Posts come from the content repo via the API. The old mock service is gone:
// pointing VITE_API_URL at a locally running API is the way to work offline.
export const blogService: BlogService = apiBlogService
