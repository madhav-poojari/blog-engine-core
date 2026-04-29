import { Container } from '@/components/ui/Container'
import { BlogCard } from '@/components/blog/BlogCard'
import { BlogCardSkeleton } from '@/components/blog/BlogCardSkeleton'
import { useBlogs } from '@/hooks/useBlog'

export function BlogListingPage() {
  const { data: blogs, isLoading, error } = useBlogs()

  // For now, show all blogs (including private). When auth is added,
  // filter private blogs for non-admin users.
  const visibleBlogs = blogs

  return (
    <Container size="md">
      <div className="mb-10">
        <h1 className="text-4xl md:text-5xl font-serif font-bold text-ink mb-3">
          Blog
        </h1>
        <p className="text-ink-muted font-mono text-sm tracking-wider">
          All posts, newest first
        </p>
        <div className="w-16 border-t-2 border-accent mt-4" />
      </div>

      {error && (
        <div className="border border-red-300 bg-red-50 p-6 text-red-800 font-mono text-sm">
          Something went wrong loading the posts. Please try again.
        </div>
      )}

      {isLoading && (
        <div className="grid gap-8 md:grid-cols-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <BlogCardSkeleton key={i} />
          ))}
        </div>
      )}

      {visibleBlogs && visibleBlogs.length === 0 && (
        <p className="text-ink-muted font-mono text-center py-16">
          No posts yet. Start writing!
        </p>
      )}

      {visibleBlogs && visibleBlogs.length > 0 && (
        <div className="grid gap-8 md:grid-cols-2">
          {visibleBlogs.map((blog) => (
            <BlogCard key={blog.id} blog={blog} />
          ))}
        </div>
      )}
    </Container>
  )
}
