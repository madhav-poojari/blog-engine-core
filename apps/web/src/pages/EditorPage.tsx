import { useState } from 'react'
import { Container } from '@/components/ui/Container'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { BlogEditor } from '@/components/blog/BlogEditor'
import { useCreateBlog } from '@/hooks/useBlog'
import { useNavigate } from '@tanstack/react-router'
import { Eye, EyeOff, Send } from 'lucide-react'
import type { JSONContent } from '@tiptap/react'

export function EditorPage() {
  const [title, setTitle] = useState('')
  const [excerpt, setExcerpt] = useState('')
  const [tags, setTags] = useState('')
  const [visibility, setVisibility] = useState<'public' | 'private'>('public')
  const [content, setContent] = useState<JSONContent>({
    type: 'doc',
    content: [{ type: 'paragraph' }],
  })

  const createBlog = useCreateBlog()
  const navigate = useNavigate()

  const handlePublish = () => {
    if (!title.trim()) return

    createBlog.mutate(
      {
        title: title.trim(),
        excerpt: excerpt.trim(),
        content,
        tags: tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
        visibility,
      },
      {
        onSuccess: () => {
          navigate({ to: '/blog' })
        },
      }
    )
  }

  return (
    <Container size="md">
      <div className="mb-8">
        <h1 className="text-3xl font-serif font-bold text-ink mb-2">New Post</h1>
        <div className="w-12 border-t-2 border-accent" />
      </div>

      <div className="space-y-6">
        {/* Title */}
        <div>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Post title"
            className="w-full text-3xl font-serif font-bold text-ink bg-transparent border-none outline-none placeholder:text-ink-muted/40 py-2"
          />
        </div>

        {/* Excerpt */}
        <div>
          <textarea
            value={excerpt}
            onChange={(e) => setExcerpt(e.target.value)}
            placeholder="A brief excerpt or summary..."
            rows={2}
            className="w-full font-mono text-sm text-ink-light bg-transparent border border-border p-4 outline-none placeholder:text-ink-muted/40 resize-none focus:border-accent transition-colors"
          />
        </div>

        {/* Tags */}
        <div>
          <input
            type="text"
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="Tags (comma-separated): react, golang, writing"
            className="w-full font-mono text-sm text-ink-light bg-transparent border border-border p-4 outline-none placeholder:text-ink-muted/40 focus:border-accent transition-colors"
          />
        </div>

        {/* Visibility toggle */}
        <div className="flex items-center gap-4">
          <span className="font-mono text-sm text-ink-muted">Visibility:</span>
          <button
            type="button"
            onClick={() => setVisibility(visibility === 'public' ? 'private' : 'public')}
            className="inline-flex items-center gap-2 cursor-pointer bg-transparent border-none p-0"
          >
            {visibility === 'public' ? (
              <>
                <Eye size={16} className="text-green-700" />
                <Badge variant="public">Public</Badge>
              </>
            ) : (
              <>
                <EyeOff size={16} className="text-accent" />
                <Badge variant="private">Private</Badge>
              </>
            )}
          </button>
        </div>

        {/* Editor */}
        <BlogEditor content={content} onChange={setContent} />

        {/* Actions */}
        <div className="flex items-center justify-end gap-4 pt-4">
          <Button
            variant="primary"
            size="lg"
            onClick={handlePublish}
            disabled={!title.trim() || createBlog.isPending}
          >
            <Send size={16} className="mr-2" />
            {createBlog.isPending ? 'Publishing...' : 'Publish'}
          </Button>
        </div>
      </div>
    </Container>
  )
}
