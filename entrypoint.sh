#!/bin/sh
# HF Space entrypoint: restore /data from the /backup bucket on cold boot,
# start the periodic backup loop, then exec the app server.
#
# Why dump/restore instead of putting SQLite directly on the bucket: HF buckets
# are FUSE mounts without the shared-memory + POSIX-lock semantics WAL-mode
# SQLite requires — a live DB on FUSE corrupts. So the live DB stays on the
# Space's ephemeral disk and only snapshots go to the bucket
# (RPO ~= BACKUP_INTERVAL_MIN). Same pattern as MyAPI Hub's pg_dump loop.
set -eu

: "${DATA_DIR:=/data}"
: "${BACKUP_DIR:=/backup}"
DB="$DATA_DIR/synthetic-vision.db"

mkdir -p "$DATA_DIR"

if [ ! -s "$DB" ] && [ -s "$BACKUP_DIR/latest.db.gz" ]; then
    echo "[init] restoring DB from $BACKUP_DIR/latest.db.gz"
    if gunzip -c "$BACKUP_DIR/latest.db.gz" > "$DB.restore" \
        && [ -s "$DB.restore" ] \
        && [ "$(sqlite3 "$DB.restore" 'PRAGMA integrity_check;')" = "ok" ]; then
        mv "$DB.restore" "$DB"
        echo "[init] DB restore ok"
    else
        # Do NOT delete anything in the bucket here: timestamped sv-*.db.gz
        # snapshots remain available for manual recovery.
        echo "[init] DB restore FAILED — starting empty (snapshots kept in $BACKUP_DIR)"
        rm -f "$DB.restore"
    fi

    # Generated images + uploaded references: additive copy back.
    for d in images references; do
        if [ -d "$BACKUP_DIR/$d" ]; then
            mkdir -p "$DATA_DIR/$d"
            cp -r "$BACKUP_DIR/$d/." "$DATA_DIR/$d/" \
                || echo "[init] $d restore had errors (continuing)"
        fi
    done
fi

# Only run the backup loop when the bucket is actually mounted, so plain
# `docker compose up` (no /backup) keeps working unchanged.
if [ -d "$BACKUP_DIR" ]; then
    /usr/local/bin/backup.sh loop &
else
    echo "[init] $BACKUP_DIR not mounted — backups disabled"
fi

exec /usr/local/bin/server
