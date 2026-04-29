-- blogs table
CREATE TABLE IF NOT EXISTS blogs (
    id               TEXT PRIMARY KEY,
    slug             TEXT NOT NULL UNIQUE,
    title            TEXT NOT NULL,
    excerpt          TEXT NOT NULL DEFAULT '',
    content          JSONB NOT NULL DEFAULT '{}',
    cover_image      TEXT,
    tags             TEXT[] NOT NULL DEFAULT '{}',
    visibility       TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    read_time_minutes INTEGER NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blogs_slug ON blogs (slug);
CREATE INDEX IF NOT EXISTS idx_blogs_visibility ON blogs (visibility);
CREATE INDEX IF NOT EXISTS idx_blogs_created_at ON blogs (created_at DESC);

-- admin table
CREATE TABLE IF NOT EXISTS admins (
    id            TEXT PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
