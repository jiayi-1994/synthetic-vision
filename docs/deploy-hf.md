# Deploying to Hugging Face Spaces

Runbook for the live deployment at
`https://looknicemm-synthetic-vision.hf.space`
(Space: `Looknicemm/synthetic-vision`, docker SDK, free `cpu-basic` hardware).
Adjust the namespace for your own account.

## Architecture on HF

```
┌─ HF Space (docker, ephemeral disk) ──────────────────┐
│  entrypoint.sh                                       │
│   ├─ cold boot: restore /data from /backup bucket    │
│   ├─ backup.sh loop (15 min snapshots)  ─────────────┼──► hf://buckets/<ns>/synthetic-vision-data
│   └─ exec server  (SQLite WAL + images on /data)     │       (FUSE mount at /backup)
└──────────────────────────────────────────────────────┘
```

**Why dump/restore instead of mounting the bucket as `/data`:** buckets are
FUSE mounts without the shared-memory and POSIX-lock semantics WAL-mode SQLite
needs — a live DB on FUSE corrupts. The live DB stays on the Space's ephemeral
disk; only snapshots go to the bucket. **RPO ≈ `BACKUP_INTERVAL_MIN`**
(default 15 min). Zero-RPO requires HF's paid persistent storage instead.

## One-time setup

```bash
hf auth login                      # write token

# 1. Space (docker SDK). README.md frontmatter sets sdk + app_port: 8080.
hf repos create <ns>/synthetic-vision --type space --space-sdk docker --public

# 2. Private bucket for backups (holds user DB + images — keep it private).
hf buckets create synthetic-vision-data --private

# 3. Mount it at /backup. NOTE: `volumes set` REPLACES all volumes.
hf spaces volumes set <ns>/synthetic-vision \
  -v "hf://buckets/<ns>/synthetic-vision-data:/backup"

# 4. Secrets — every key from .env EXCEPT PORT and DATA_DIR (fixed in the
#    Dockerfile; overriding them breaks the container layout).
#    The secrets file must be LF-terminated: CRLF leaves a trailing \r in
#    values and the upstream API key silently 401s.
hf spaces secrets add <ns>/synthetic-vision --secrets-file <filtered .env>
```

Required secrets: `JWT_SECRET` (≥32 chars — the app refuses to boot on the
default when `IMAGE_PROVIDER=openai`), `OPENAI_BASE_URL`, `OPENAI_API_KEY`,
`IMAGE_MODEL`, `IMAGE_PROVIDER`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`,
`JWT_TTL_HOURS`, `GEN_WORKERS`, `SEED_DEMO_USERS`, `MOCK_DELAY_MS`.

## Deploy / update code

`git push` to the Space remote is rejected — the history contains binary PNGs
(`_design/`) and HF requires xet/LFS for binaries. Use `hf upload` of a clean
`git archive` export instead (this also guarantees `.env`, stress artifacts,
and other untracked files never leak into the public Space repo):

```bash
TMP=$(mktemp -d)
git archive HEAD | tar -x -C "$TMP"
hf upload <ns>/synthetic-vision "$TMP" . --repo-type space \
  --commit-message "deploy: <what changed>"
rm -rf "$TMP"
```

The upload triggers a rebuild automatically. Watch it:

```bash
hf spaces info <ns>/synthetic-vision      # stage: BUILDING → RUNNING
hf spaces logs <ns>/synthetic-vision      # boot logs: look for "[init]" lines
```

On a healthy cold boot with existing backups you should see:

```
[init] restoring DB from /backup/latest.db.gz
[init] DB restore ok
synthetic-vision listening on :8080
[backup] wrote sv-<timestamp>.db.gz (...)
```

## Backup & restore operations

Bucket layout:

| Path | What |
|---|---|
| `latest.db.gz` | newest snapshot, auto-restored on cold boot (after `PRAGMA integrity_check`) |
| `sv-<UTC ts>.db.gz` | timestamped snapshots, newest `BACKUP_KEEP` (24) kept |
| `images/`, `references/` | generated PNGs + uploaded sources, additive-only copies |

```bash
# Force an immediate snapshot (no shell access needed) — the loop notices the
# marker file within 60s, snapshots, and deletes it:
touch backup-now && hf buckets cp backup-now hf://buckets/<ns>/synthetic-vision-data/backup-now

# Inspect backups:
hf buckets ls <ns>/synthetic-vision-data

# Disaster recovery to a specific snapshot:
hf buckets cp hf://buckets/<ns>/synthetic-vision-data/sv-<ts>.db.gz ./
hf buckets cp ./sv-<ts>.db.gz hf://buckets/<ns>/synthetic-vision-data/latest.db.gz
hf spaces restart <ns>/synthetic-vision    # cold boot restores latest.db.gz
```

Design guarantees worth not breaking:

- **WAL-safe snapshots**: `backup.sh` uses `sqlite3 .backup` (online backup
  API), never a raw `cp` of the live DB file.
- **Additive-only image copies** — deliberately no `--delete`/mirroring. A
  failed restore leaves `/data` empty; a mirroring sync would then wipe the
  bucket too. Orphaned files from user deletions are unreachable once the DB
  row is gone and cost ~nothing.
- **Restore failure is non-destructive**: a corrupt `latest.db.gz` boots an
  empty DB but timestamped snapshots are never touched by the entrypoint.
- `latest.db.gz` is written via tmp-file + rename so readers never see a
  half-written snapshot.

## Day-2 commands

```bash
hf spaces restart <ns>/synthetic-vision                 # data survives (restored from bucket)
hf spaces secrets ls|add <ns>/synthetic-vision          # rotate keys (add triggers restart)
hf spaces logs <ns>/synthetic-vision                    # on Windows GBK errors:
#   [Console]::OutputEncoding=[Text.Encoding]::UTF8 first
```

Gotchas collected from the first deployment:

- Space is **public**: anyone can register (1,250-credit bonus) and spend your
  upstream image-API quota. Change `ADMIN_PASSWORD` from the README default.
- `ADMIN_PASSWORD`/`SEED_DEMO_USERS` only apply when seeding an **empty** DB;
  once backups exist, accounts come from the restored snapshot.
- `IMAGE_MODEL` must exist in the upstream `GET /v1/models` list for your key.
- Local `docker compose up` is unaffected by all of this: without a `/backup`
  mount the entrypoint skips restore and disables the backup loop.
