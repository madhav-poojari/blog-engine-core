import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Typography from '@tiptap/extension-typography'
import Underline from '@tiptap/extension-underline'
import type { JSONContent } from '@tiptap/react'

interface BlogContentProps {
  content: JSONContent
}

export function BlogContent({ content }: BlogContentProps) {
  const editor = useEditor(
    {
      extensions: [StarterKit, Image, Typography, Underline],
      content,
      editable: false,
      editorProps: {
        attributes: {
          class:
            'prose prose-lg max-w-none font-mono ' +
            'prose-headings:font-serif prose-headings:text-ink ' +
            'prose-p:text-ink-light prose-p:leading-relaxed ' +
            'prose-a:text-accent prose-a:underline ' +
            'prose-blockquote:border-l-accent prose-blockquote:border-l-2 prose-blockquote:pl-6 prose-blockquote:italic prose-blockquote:text-ink-muted ' +
            'prose-code:bg-cream-dark prose-code:px-1.5 prose-code:py-0.5 prose-code:text-sm prose-code:font-mono ' +
            'prose-pre:bg-ink prose-pre:text-cream prose-pre:font-mono prose-pre:text-sm prose-pre:p-6 prose-pre:overflow-x-auto ' +
            'prose-img:border prose-img:border-border ' +
            'prose-li:text-ink-light ' +
            'prose-strong:text-ink prose-strong:font-bold',
        },
      },
    },
    [content]
  )

  if (!editor) return null

  return <EditorContent editor={editor} />
}
