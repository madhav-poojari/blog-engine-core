interface BlogContentProps {
  /**
   * HTML rendered by the API from the post's markdown. It is trusted because
   * it comes from our own content repo and the server renders with raw HTML
   * passthrough disabled, so markdown cannot inject script tags.
   */
  contentHtml: string
}

const proseClasses =
  'prose prose-lg max-w-none font-mono ' +
  'prose-headings:font-serif prose-headings:text-ink ' +
  'prose-p:text-ink-light prose-p:leading-relaxed ' +
  'prose-a:text-accent prose-a:underline ' +
  'prose-blockquote:border-l-accent prose-blockquote:border-l-2 prose-blockquote:pl-6 prose-blockquote:italic prose-blockquote:text-ink-muted ' +
  'prose-code:bg-cream-dark prose-code:px-1.5 prose-code:py-0.5 prose-code:text-sm prose-code:font-mono ' +
  'prose-pre:bg-ink prose-pre:text-cream prose-pre:font-mono prose-pre:text-sm prose-pre:p-6 prose-pre:overflow-x-auto ' +
  'prose-img:border prose-img:border-border ' +
  'prose-li:text-ink-light ' +
  'prose-strong:text-ink prose-strong:font-bold'

export function BlogContent({ contentHtml }: BlogContentProps) {
  return <div className={proseClasses} dangerouslySetInnerHTML={{ __html: contentHtml }} />
}
