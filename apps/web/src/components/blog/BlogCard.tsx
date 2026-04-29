import { Link } from '@tanstack/react-router'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { formatDate } from '@/lib/utils'
import { Calendar, Clock, Lock } from 'lucide-react'
import type { Blog } from '@/types/blog'

interface BlogCardProps {
  blog: Blog
}

export function BlogCard({ blog }: BlogCardProps) {
  return (
    <Link to="/blog/$slug" params={{ slug: blog.slug }} className="no-underline block">
      <Card hover>
        {/* Cover image */}
        {blog.coverImage && (
          <div className="mb-4 -mx-6 -mt-6 overflow-hidden">
            <img
              src={blog.coverImage}
              alt={blog.title}
              className="w-full h-48 object-cover grayscale-[30%] hover:grayscale-0 transition-all duration-500"
            />
          </div>
        )}

        {/* Tags & visibility */}
        <div className="flex flex-wrap items-center gap-2 mb-3">
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

        {/* Title */}
        <h2 className="text-xl font-serif font-bold text-ink mb-2 leading-snug">
          {blog.title}
        </h2>

        {/* Excerpt */}
        <p className="text-ink-light text-sm leading-relaxed mb-4 line-clamp-2">
          {blog.excerpt}
        </p>

        {/* Meta */}
        <div className="flex items-center gap-4 text-xs text-ink-muted font-mono">
          <span className="inline-flex items-center gap-1">
            <Calendar size={12} />
            {formatDate(blog.createdAt)}
          </span>
          <span className="inline-flex items-center gap-1">
            <Clock size={12} />
            {blog.readTimeMinutes} min read
          </span>
        </div>
      </Card>
    </Link>
  )
}
