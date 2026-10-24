# Going live, layer by layer

The blog is now served from a private GitHub repo of markdown files. The API
keeps a working clone on the VM, indexes it in memory, and follows pushes.
There is no database.

Each layer below can be checked on its own, and each one only depends on the
ones above it. If something is wrong, the lowest failing layer tells you where
it is — don't debug layer 5 while layer 2 is broken.

---

## Layer 0 — The content logic (offline, no git, no network)

This is where all the correctness lives: frontmatter parsing, defaults, tags,
slugs, quarantine.

```bash
cd apps/api && go test ./internal/content/ -v
```

27 tests. The one that matters most is `TestNormalizeIsAFixpoint`: it proves a
second normalize pass changes nothing, which is what stops the write-back →
webhook → write-back cycle from committing forever. If you ever change the
normalizer and that test fails, stop and fix it before deploying.

---

## Layer 1 — Your real content, interpreted

Point the indexer at a clone of your actual content repo. No config, no server,
no network:

```bash
cd apps/api && go build -o server ./cmd/server && ./server check /path/to/your/blog-content
```

You get a table of every post with its visibility, slug, tags and path, plus
two lists: files that are quarantined, and files still missing frontmatter.

**This is the single most useful command for checking the blog.** Read the
output carefully before going further:

- Every post you have not explicitly marked `visibility: public` shows as
  `private`. That is correct and intended — nothing is public until you say so.
- `uncategorised` should not appear as a tag. Other folder names should.
- Anything quarantined is **excluded from the site entirely** and is never
  auto-rewritten, because its YAML intent is ambiguous. Fix those by hand.

---

## Layer 2 — Git access from the VM

Create a **deploy key with write access** on the content repo
(Settings → Deploy keys → Add, tick "Allow write access"). Write access is
needed because the server pushes generated frontmatter back.

```bash
ssh-keygen -t ed25519 -f /etc/blog-engine/deploy_key -N ''
chmod 600 /etc/blog-engine/deploy_key
# paste /etc/blog-engine/deploy_key.pub into GitHub

# verify, as the user the service runs as:
GIT_SSH_COMMAND='ssh -i /etc/blog-engine/deploy_key -o IdentitiesOnly=yes' \
  git ls-remote git@github.com:<you>/blog-content.git refs/heads/main
```

That must print a SHA. If it doesn't, nothing above this layer will work.

The clone at `CONTENT_REPO_PATH` is **server-owned and disposable**. Never edit
files there by hand — the server runs `git reset --hard` on every sync and will
discard anything local. Deleting the whole directory is always a safe recovery;
it gets re-cloned.

---

## Layer 3 — The API

```bash
cd apps/api
cp .env.example .env     # then fill it in
./server hashpw 'your-password'   # paste the output into ADMIN_PASSWORD_HASH
./server
```

On boot it clones (first run takes a moment), indexes, and logs a line like:

```
sync (startup): ok, 43 posts (12 public) at a1b2c3d4 in 180ms
```

Check it from outside, in this order:

```bash
# 1. Is the index warm?
curl -s localhost:8081/api/health | jq

# 2. Public reads — must contain ONLY posts you marked public
curl -s localhost:8081/api/blogs | jq -r '.[] | "\(.visibility)  \(.slug)"'

# 3. A private post must 404 anonymously (not 403 — we don't confirm it exists)
curl -s -o /dev/null -w '%{http_code}\n' localhost:8081/api/blogs/<a-private-slug>

# 4. Tags
curl -s localhost:8081/api/tags | jq

# 5. Log in
TOKEN=$(curl -s -X POST localhost:8081/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"<user>","password":"<pass>"}' | jq -r .token)

# 6. Same list as admin — now includes private posts
curl -s localhost:8081/api/blogs -H "Authorization: Bearer $TOKEN" | jq length

# 7. Sync status: quarantined files, last error, breaker state
curl -s localhost:8081/api/admin/sync/status -H "Authorization: Bearer $TOKEN" | jq
```

**Step 2 and 3 are the checks that matter.** Default-private is the one
property where a bug is expensive, so confirm it with your real content before
the site is reachable.

Manual triggers, for when you don't want to wait:

```bash
curl -s -X POST localhost:8081/api/admin/sync      -H "Authorization: Bearer $TOKEN" | jq
curl -s -X POST localhost:8081/api/admin/normalize -H "Authorization: Bearer $TOKEN" | jq
```

---

## Layer 4 — Push-to-publish

Add the webhook on the content repo (Settings → Webhooks → Add):

- **Payload URL**: `https://blogs-api.<your-domain>/api/sync`
- **Content type**: `application/json`
- **Secret**: the same value as `GITHUB_WEBHOOK_SECRET`
- **Events**: just the push event

GitHub sends a ping immediately; it should go green. Then:

```bash
# in the content repo
echo 'test' >> uncategorised/ping.md && git commit -am 'ping' && git push
# within a few seconds:
curl -s localhost:8081/api/health | jq .headSha   # matches the new commit
```

A tampered signature must be rejected:

```bash
curl -s -o /dev/null -w '%{http_code}\n' -X POST localhost:8081/api/sync \
  -H 'X-GitHub-Event: push' -H 'X-Hub-Signature-256: sha256=deadbeef' \
  -H 'Content-Type: application/json' -d '{"ref":"refs/heads/main"}'
# expect 401
```

If the webhook can't reach you, nothing breaks: the poll
(`SYNC_POLL_INTERVAL`) picks the push up within a few minutes.

### What writes frontmatter, and when

Reading and writing are deliberately separate:

- **Every sync** fills in missing metadata **in memory**, so a file you push
  with no frontmatter at all is live within seconds with a correct title, tags
  and dates.
- **Once a day** (`SYNC_DAILY_SWEEP_AT`), and whenever you call
  `/api/admin/normalize`, that metadata is written into the files, committed as
  `blog-engine-bot` and pushed.

So a new post is *served* immediately and becomes *self-describing* by the next
morning. `created` is written into the file once and then never touched again.
`updated` is deliberately never written — it is computed from the last non-bot
commit, because writing it would change the file's commit time and loop.

To confirm write-back is a true no-op once settled, run `./server check` on the
clone right after a normalize: the "would add" list must be empty.

---

## Layer 5 — The frontend

```bash
cd apps/web
echo 'VITE_API_URL=http://localhost:8081/api' > .env.local
pnpm dev
```

Check: a post renders with real typography, code blocks are highlighted,
Cloudflare images load, and `/blog/<slug>` survives a hard refresh.

Then log in at `/admin/login` and confirm private posts appear in the listing
(the server decides this from your token — the frontend does no filtering).

**Cloudflare Pages:**

- Build command `pnpm build`, output directory `apps/web/dist`
- Environment variable `VITE_API_URL=https://blogs-api.<your-domain>/api`
- `public/_redirects` is already committed with `/*  /index.html  200`. Without
  it, deep links like `/blog/<slug>` 404 on Pages.

---

## Layer 6 — Cutover checklist

1. Content repo is **private**.
2. Deploy key has **write** access (Layer 2 verified).
3. `ADMIN_PASSWORD_HASH` is set; the old `ADMIN_PASSWORD` is gone.
4. `ALLOWED_ORIGINS` contains the exact Pages domain. The old config allowed
   only `madhavapoojari.fyi` while the frontend called `brschess.com` — make
   these agree, or every browser request fails CORS.
5. `git` is installed on the VM (`git --version`). The server refuses to boot
   without it, by design.
6. Postgres is no longer used — stop and disable it once you're happy.
7. `./server check` output reviewed against your real content.

### Dockerfile note

If you deploy with Docker rather than running the binary directly, the runtime
stage now needs git:

```dockerfile
RUN apk add --no-cache ca-certificates tzdata git openssh-client
```

and no longer needs to copy `migrations/`. Mount a volume at
`CONTENT_REPO_PATH` so the clone survives a deploy, or accept a re-clone on
each start (fine at this size, but the first health check will be slow).

---

## If something looks wrong

| Symptom | Where to look |
|---|---|
| A post is missing from the site | `./server check` — is it quarantined, or just private? |
| Every post's date is today | Dates seed from git history; check the clone isn't shallow |
| Repeated `chore(content): normalize frontmatter` commits | The circuit breaker caps this at 5/hour and reports `breakerOpen` in sync status. Run the Layer 0 fixpoint test |
| Site stale after a push | `/api/health` `headSha` vs GitHub; check the webhook's Recent Deliveries |
| 500s on boot | `git` missing, or the deploy key can't read the repo |
| Empty blog list but healthy | Everything is private — that is the default until you mark posts public |

---

## Not built yet (deliberately)

- **Writing posts from the frontend.** The editor page was removed; `POST/PUT/
  DELETE /api/blogs` answer 405 with an explanatory message. The API shape is
  designed so this returns later as a commit through the GitHub Contents API,
  picked up by the normal sync path.
- **Tag filtering UI.** `GET /api/tags` and `GET /api/blogs?tag=x` work and
  `useTags()` is wired; the listing page doesn't render filter controls yet.
- **Admin sync panel.** Use the `curl` calls in Layer 3 for now.
