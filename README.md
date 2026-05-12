# Typewriter Blog — Personal Blogging Platform

A minimalist, typewriter-aesthetic blogging platform built with React/TypeScript (frontend) and Go (backend) in a Turborepo monorepo.

## Tech Stack

| Layer | Tools |
|-------|-------|
| **Frontend** | React 18 + TypeScript (Vite), Tailwind CSS v4, TanStack Query, TanStack Router, TipTap editor |
| **Backend** | Go + Chi router, sqlx, PostgreSQL, JWT auth, S3 for images |
| **Monorepo** | Turborepo + pnpm workspaces |
| **Fonts** | Courier Prime (body/editor) + Playfair Display (headings) |
| **Palette** | Cream bg `#FAF8F1`, dark text `#2C2C2C`, muted accent `#8B7355` |
| **Deploy** | Frontend → Cloudflare Pages · Backend → Docker on VPS · DB → DigitalOcean PostgreSQL |


## Architecture

```
portfolio-website/
├── apps/
│   ├── web/                    # React frontend (Vite)
│   │   └── src/
│   │       ├── components/
│   │       │   ├── ui/         # Button, Card, Badge, Container
│   │       │   ├── layout/     # Navbar, Footer, Layout
│   │       │   └── blog/       # BlogCard, BlogEditor, BlogContent
│   │       ├── pages/          # Home, BlogListing, BlogDetail, EditorPage, NotFound
│   │       ├── services/       # API abstraction (mock → real swap point)
│   │       ├── hooks/          # TanStack Query wrappers
│   │       ├── types/          # TypeScript interfaces
│   │       ├── data/           # Static/mock data
│   │       └── lib/            # Utilities
│   └── api/                    # Go backend
│       ├── cmd/
│       │   └── server/         # main.go entry point
│       ├── internal/
│       │   ├── config/         # Environment config loader
│       │   ├── handler/        # HTTP handlers (blog, auth, health, image)
│       │   ├── middleware/     # Auth, CORS, logging, recovery
│       │   ├── model/          # Data models / domain types
│       │   ├── repository/     # Database access (sqlx)
│       │   ├── service/        # Business logic layer
│       │   └── storage/        # S3/object storage client
│       ├── migrations/         # SQL migration files
│       ├── Dockerfile
│       ├── .env.example
│       └── go.mod
├── turbo.json
├── pnpm-workspace.yaml
└── README.md
```

## Key Design Decisions

- **TipTap** — headless editor = full control over typewriter styling, low bundle size
- **Service abstraction layer** — swap mock → real API by changing only `services/`
- **TanStack Router** — type-safe routing with excellent TypeScript DX
- **Blog content stored as JSON** — TipTap's native format, maps to PostgreSQL `jsonb`
- **Iteration 1 excludes**: backend, auth UI, image upload, deployment config
