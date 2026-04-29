import type { Blog } from '@/types/blog'

export const mockBlogs: Blog[] = [
  {
    id: '1',
    slug: 'building-a-typewriter-blog',
    title: 'Building a Typewriter-Themed Blog from Scratch',
    excerpt:
      'A deep dive into creating a minimalist blogging platform with a vintage typewriter aesthetic, using modern web technologies.',
    content: {
      type: 'doc',
      content: [
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'Why a Typewriter Theme?' }],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'There is something deeply satisfying about the mechanical rhythm of a typewriter. Each keystroke is deliberate, each word carries weight. In our age of instant deletion and infinite revision, the typewriter reminds us that writing is a craft — something that demands thought before action.',
            },
          ],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'This blog is built with that philosophy in mind. The design is intentionally sparse. No flashy animations, no gradient backgrounds, no dark mode toggle. Just ',
            },
            {
              type: 'text',
              marks: [{ type: 'bold' }],
              text: 'words on a page',
            },
            {
              type: 'text',
              text: ', the way they were meant to be read.',
            },
          ],
        },
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'The Tech Stack' }],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'Under the hood, this site uses a modern stack that prioritizes developer experience and performance:',
            },
          ],
        },
        {
          type: 'bulletList',
          content: [
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [
                    {
                      type: 'text',
                      marks: [{ type: 'bold' }],
                      text: 'React + TypeScript',
                    },
                    { type: 'text', text: ' — Type-safe component architecture' },
                  ],
                },
              ],
            },
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [
                    {
                      type: 'text',
                      marks: [{ type: 'bold' }],
                      text: 'Tailwind CSS',
                    },
                    { type: 'text', text: ' — Utility-first styling with a custom typewriter palette' },
                  ],
                },
              ],
            },
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [
                    {
                      type: 'text',
                      marks: [{ type: 'bold' }],
                      text: 'TipTap',
                    },
                    { type: 'text', text: ' — Headless rich text editor for writing posts' },
                  ],
                },
              ],
            },
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [
                    {
                      type: 'text',
                      marks: [{ type: 'bold' }],
                      text: 'Go',
                    },
                    { type: 'text', text: ' — Backend API, fast and reliable' },
                  ],
                },
              ],
            },
          ],
        },
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'Code Example' }],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'Here is a simple example of how the blog service abstraction works:',
            },
          ],
        },
        {
          type: 'codeBlock',
          attrs: { language: 'typescript' },
          content: [
            {
              type: 'text',
              text: 'const blogs = await blogService.getBlogs();\nconst published = blogs.filter(b => b.visibility === "public");',
            },
          ],
        },
        {
          type: 'blockquote',
          content: [
            {
              type: 'paragraph',
              content: [
                {
                  type: 'text',
                  marks: [{ type: 'italic' }],
                  text: '"The typewriter is holy, the typewriter is the thing." — Jack Kerouac',
                },
              ],
            },
          ],
        },
      ],
    },
    coverImage: 'https://images.unsplash.com/photo-1504691342899-4d92b50853e1?w=800&q=80',
    tags: ['web-dev', 'design', 'react'],
    visibility: 'public',
    readTimeMinutes: 5,
    createdAt: '2026-03-20T10:00:00Z',
    updatedAt: '2026-03-20T10:00:00Z',
  },
  {
    id: '2',
    slug: 'on-writing-well',
    title: 'On Writing Well: Lessons from Strunk & White',
    excerpt:
      'Timeless principles of clear writing that every blogger should internalize. Brevity, clarity, and the courage to cut.',
    content: {
      type: 'doc',
      content: [
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'Omit Needless Words' }],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'Vigorous writing is concise. A sentence should contain no unnecessary words, a paragraph no unnecessary sentences, for the same reason that a drawing should have no unnecessary lines and a machine no unnecessary parts.',
            },
          ],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'This requires not that the writer make all sentences short, or avoid all detail, but that ',
            },
            {
              type: 'text',
              marks: [{ type: 'bold' }],
              text: 'every word tell',
            },
            { type: 'text', text: '.' },
          ],
        },
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'Write with Nouns and Verbs' }],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'Write with nouns and verbs, not with adjectives and adverbs. The adjective hasn\'t been built that can pull a weak or inaccurate noun out of a tight place.',
            },
          ],
        },
        {
          type: 'blockquote',
          content: [
            {
              type: 'paragraph',
              content: [
                {
                  type: 'text',
                  marks: [{ type: 'italic' }],
                  text: '"A sentence should read as if its author, had he held a plough instead of a pen, could have drawn a furrow deep and straight."',
                },
              ],
            },
          ],
        },
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'The Discipline of Rewriting' }],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'The best writing is rewriting. Few writers are so expert that they can produce what they are after on the first try. Most need to rewrite — repeatedly. And the act of rewriting is itself an act of discovery.',
            },
          ],
        },
        {
          type: 'orderedList',
          content: [
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [{ type: 'text', text: 'Write your first draft freely' }],
                },
              ],
            },
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [{ type: 'text', text: 'Cut it by a third' }],
                },
              ],
            },
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [{ type: 'text', text: 'Read it aloud — if it stumbles, rewrite' }],
                },
              ],
            },
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [{ type: 'text', text: 'Let it rest, then cut again' }],
                },
              ],
            },
          ],
        },
      ],
    },
    coverImage: 'https://images.unsplash.com/photo-1455390582262-044cdead277a?w=800&q=80',
    tags: ['writing', 'craft'],
    visibility: 'public',
    readTimeMinutes: 4,
    createdAt: '2026-03-15T08:30:00Z',
    updatedAt: '2026-03-16T14:00:00Z',
  },
  {
    id: '3',
    slug: 'golang-concurrency-patterns',
    title: 'Go Concurrency Patterns I Use Every Day',
    excerpt:
      'Practical goroutine patterns for real-world applications — fan-out, fan-in, worker pools, and graceful shutdown.',
    content: {
      type: 'doc',
      content: [
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'Goroutines Are Not Free' }],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: "While goroutines are cheap to create (a few KB of stack space), they are not free. Spawning thousands of goroutines without control leads to resource exhaustion and unpredictable behavior. Every goroutine you spawn should have a clear lifecycle — a way to start, a way to stop.",
            },
          ],
        },
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'The Worker Pool Pattern' }],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'One of my most-used patterns. Instead of spawning a goroutine per task, you create a fixed pool of workers that pull jobs from a channel:',
            },
          ],
        },
        {
          type: 'codeBlock',
          attrs: { language: 'go' },
          content: [
            {
              type: 'text',
              text: 'func worker(id int, jobs <-chan Job, results chan<- Result) {\n    for job := range jobs {\n        result := process(job)\n        results <- result\n    }\n}',
            },
          ],
        },
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'Graceful Shutdown' }],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              text: 'Use ',
            },
            {
              type: 'text',
              marks: [{ type: 'code' }],
              text: 'context.Context',
            },
            {
              type: 'text',
              text: ' to propagate cancellation. Every long-running goroutine should accept a context and respect its cancellation signal.',
            },
          ],
        },
      ],
    },
    coverImage: 'https://images.unsplash.com/photo-1515879218367-8466d910auj7?w=800&q=80',
    tags: ['golang', 'concurrency', 'backend'],
    visibility: 'public',
    readTimeMinutes: 7,
    createdAt: '2026-03-10T16:00:00Z',
    updatedAt: '2026-03-10T16:00:00Z',
  },
  {
    id: '4',
    slug: 'draft-ideas-2026',
    title: 'Blog Ideas & Drafts — 2026',
    excerpt:
      'A private collection of half-formed thoughts, ideas for future posts, and rough drafts.',
    content: {
      type: 'doc',
      content: [
        {
          type: 'heading',
          attrs: { level: 2 },
          content: [{ type: 'text', text: 'Ideas Backlog' }],
        },
        {
          type: 'bulletList',
          content: [
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [{ type: 'text', text: 'PostgreSQL full-text search vs. Elasticsearch for a blog' }],
                },
              ],
            },
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [{ type: 'text', text: 'Deploying Go services with Docker Compose' }],
                },
              ],
            },
            {
              type: 'listItem',
              content: [
                {
                  type: 'paragraph',
                  content: [{ type: 'text', text: 'Why I switched from Next.js to Vite + React' }],
                },
              ],
            },
          ],
        },
        {
          type: 'paragraph',
          content: [
            {
              type: 'text',
              marks: [{ type: 'italic' }],
              text: 'This is a private post — only visible to the admin.',
            },
          ],
        },
      ],
    },
    tags: ['meta', 'drafts'],
    visibility: 'private',
    readTimeMinutes: 2,
    createdAt: '2026-03-05T12:00:00Z',
    updatedAt: '2026-03-18T09:00:00Z',
  },
]
