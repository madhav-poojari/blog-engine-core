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

---

## Progress Tracker

### Iteration 1 — Frontend with Static Data

#### Phase 1: Project Scaffolding
- [x] Initialize Turborepo monorepo with pnpm workspaces
- [x] Scaffold React + Vite + TypeScript app in `apps/web`
- [x] Install & configure Tailwind CSS v4
- [x] Install TanStack Router, TanStack Query, TipTap, Lucide icons
- [x] Set up Google Fonts (Courier Prime + Playfair Display)
- [x] Create base Tailwind theme with typewriter color palette
- [x] Create this README with project overview and task tracking

#### Phase 2: Layout & Navigation
- [x] Build `Layout` component (wraps pages with Navbar + Footer)
- [x] Build `Navbar` — site title + nav links (Home, Blog), sticky, minimal
- [x] Build `Footer` — social links (GitHub, LinkedIn, Twitter/X, Email) with icons
- [x] Configure TanStack Router: `/`, `/blog`, `/blog/:slug`, `/admin/editor`

#### Phase 3: Core UI Components
- [x] Build primitives: `Button`, `Card`, `Badge`, `Container`, `Skeleton`
- [x] Build `BlogCard` — title, excerpt, date, tags, read time, visibility badge
- [x] Build `BlogContent` — TipTap read-only renderer for full blog posts

#### Phase 4: Service Layer & Mock Data
- [x] Define TypeScript types: `Blog`, `BlogCreateInput`, `BlogUpdateInput`
- [x] Create static mock data: 4 sample posts with varied content
- [x] Build service abstraction (`blogService.ts`) with mock implementation
- [x] Build TanStack Query hooks: `useBlogs()`, `useBlog(slug)`, mutations

#### Phase 5: Pages
- [x] **Home** page — hero with name/title, short bio, "Read my blog" CTA
- [x] **Blog Listing** page — grid of BlogCards sorted newest-first
- [x] **Blog Detail** page — full rendered post with metadata
- [x] **Editor** page — TipTap editor + toolbar, title input, visibility toggle, publish

#### Phase 6: Polish & Responsiveness
- [x] Responsive design (mobile hamburger nav, card layout breakpoints)
- [x] Loading states / skeleton loaders
- [x] 404 page
- [ ] Final visual polish — hover effects, focus states, consistent spacing

---

### Iteration 2 — Go Backend

#### Phase 7: Go Project Scaffolding
- [x] Initialize Go module in `apps/api`
- [x] Set up project folder structure (`cmd/`, `internal/`, `migrations/`)
- [x] Install dependencies: Chi router, sqlx, pgx, jwt-go, godotenv, cors
- [x] Create main entry point (`cmd/server/main.go`) with graceful shutdown
- [x] Create config loader (reads from environment variables)

#### Phase 8: Database Layer
- [x] Define PostgreSQL schema: `blogs` table, `admin` table
- [x] Create SQL migration files (up/down)
- [x] Build database connection helper with connection pooling
- [x] Build `BlogRepository` — CRUD operations with sqlx
- [x] Build `AdminRepository` — lookup admin by username

#### Phase 9: Business Logic (Service Layer)
- [x] Build `BlogService` — create, read, update, delete with slug generation, read-time calculation
- [x] Build `AuthService` — password hashing (bcrypt), JWT token generation/validation

#### Phase 10: HTTP Handlers & Routing
- [x] Build `BlogHandler` — REST endpoints for blog CRUD
  - `GET /api/blogs` — list blogs (public only for anonymous, all for admin)
  - `GET /api/blogs/:slug` — get single blog by slug
  - `POST /api/blogs` — create blog (admin only)
  - `PUT /api/blogs/:id` — update blog (admin only)
  - `DELETE /api/blogs/:id` — delete blog (admin only)
- [x] Build `AuthHandler` — `POST /api/auth/login`
- [x] Build `HealthHandler` — `GET /api/health`

#### Phase 11: Middleware
- [x] JWT auth middleware (extracts & validates token, sets user context)
- [x] CORS middleware (allow frontend origin)
- [x] Request logging middleware
- [x] Recovery middleware (panic → 500)

#### Phase 12: Image Upload (S3)
- [x] Build S3-compatible storage client (works with DigitalOcean Spaces / AWS S3)
- [x] Build `ImageHandler` — `POST /api/images/upload` (admin only)
- [x] Validate file type & size limits

#### Phase 13: Docker & Configuration
- [x] Create `Dockerfile` (multi-stage build)
- [x] Create `.env.example` with all required variables
- [x] Wire up all routes in main.go

#### Phase 14: Frontend Integration
- [x] Build `apiBlogService.ts` — real API calls replacing mock service
- [x] Build API client (`lib/api.ts`) — fetch wrapper with JWT token injection, error handling
- [x] Build `AuthProvider` context + `useAuth` hook — login/logout state management
- [x] Build `useLogin` hook — login mutation via TanStack Query
- [x] Build image upload service (`imageService.ts`)
- [x] Add Login page (`/admin/login`)
- [x] Add auth-aware Navbar — shows New Post + Logout when admin, Login when anonymous
- [x] Protect editor route — redirects to login if no token
- [x] Swap service in `services/index.ts` — API by default, `VITE_USE_MOCK=true` for mock fallback
- [x] TypeScript + Vite build pass with zero errors

### Iteration 3 — Deployment & Extras (Future)
- [ ] Cloudflare Pages deployment for frontend
- [ ] Docker deployment for backend on VPS
- [ ] SEO / meta tags / Open Graph
- [ ] RSS feed
- [ ] Search functionality
- [ ] Blog tags/categories filtering

---

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
