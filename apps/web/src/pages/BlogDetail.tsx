import { Container } from '@/components/ui/Container'
import { BlogContent } from '@/components/blog/BlogContent'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { Button } from '@/components/ui/Button'
import { useBlog } from '@/hooks/useBlog'
import { formatDate } from '@/lib/utils'
import { useParams, Link } from '@tanstack/react-router'
import { ArrowLeft, Calendar, Clock, Lock } from 'lucide-react'

export function BlogDetailPage() {
  const { slug } = useParams({ from: '/blog/$slug' })
  const { data: blog, isLoading, error } = useBlog(slug)

  if (isLoading) {
    return (
      <Container size="sm">
        <Skeleton className="w-32 h-4 mb-8" />
        <Skeleton className="w-full h-64 mb-6" />
        <Skeleton className="w-3/4 h-10 mb-4" />
        <Skeleton className="w-1/2 h-4 mb-8" />
        <Skeleton className="w-full h-96" />
      </Container>
    )
  }

  if (error || !blog) {
    return (
      <Container size="sm">
        <div className="text-center py-20">
          <h2 className="text-2xl font-serif text-ink mb-4">Post not found</h2>
          <p className="text-ink-muted font-mono mb-8">
            The post you're looking for doesn't exist or has been removed.
          </p>
          <Link to="/blog">
            <Button variant="secondary">
              <ArrowLeft size={16} className="mr-2" />
              Back to all posts
            </Button>
          </Link>
        </div>
      </Container>
    )
  }

  return (
    <Container size="sm">
      {/* Back link */}
      <Link
        to="/blog"
        className="inline-flex items-center gap-1.5 text-ink-muted font-mono text-sm no-underline hover:text-ink transition-colors mb-8"
      >
        <ArrowLeft size={14} />
        All posts
      </Link>

      {/* Cover image */}
      {blog.coverImage && (
        <div className="mb-8 -mx-6 md:-mx-0">
          <img
            src={blog.coverImage}
            alt={blog.title}
            className="w-full h-64 md:h-80 object-cover border border-border grayscale-[20%]"
          />
        </div>
      )}

      {/* Header */}
      <header className="mb-10">
        {/* Tags & visibility */}
        <div className="flex flex-wrap items-center gap-2 mb-4">
          {blog.visibility === 'private' && (
            <Badge variant="private">
              <Lock size={10} className="mr-1" />
              Private
            </Badge>
          )}
          {blog.tags.map((tag) => (
            <Badge key={tag}>{tag}</Badge>
          ))}
        </div>

        <h1 className="text-3xl md:text-5xl font-serif font-bold text-ink mb-4 leading-tight">
          {blog.title}
        </h1>

        {/* Meta */}
        <div className="flex items-center gap-4 text-sm text-ink-muted font-mono">
          <span className="inline-flex items-center gap-1.5">
            <Calendar size={14} />
            {formatDate(blog.createdAt)}
          </span>
          <span className="inline-flex items-center gap-1.5">
            <Clock size={14} />
            {blog.readTimeMinutes} min read
          </span>
        </div>

        <div className="w-16 border-t-2 border-accent mt-6" />
      </header>

      {/* Content */}
      <article>
        <BlogContent content={blog.content} />
      </article>

      {/* Bottom nav */}
      <div className="mt-16 pt-8 border-t border-border">
        <Link to="/blog">
          <Button variant="ghost">
            <ArrowLeft size={16} className="mr-2" />
            Back to all posts
          </Button>
        </Link>
      </div>
    </Container>
  )
}
