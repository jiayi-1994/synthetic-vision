#!/bin/sh
# Periodic SQLite snapshot + image copy → /backup (HF bucket FUSE).
# Keeps latest.db.gz plus timestamped sv-*.db.gz snapshots pruned to
# BACKUP_KEEP. WAL-safe: uses sqlite3 .backup (online backup API), never a
# raw cp of the live DB file.
set -u

: "${DATA_DIR:=/data}"
: "${BACKUP_DIR:=/backup}"
: "${BACKUP_INTERVAL_MIN:=15}"
: "${BACKUP_KEEP:=24}"
: "${BACKUP_WARMUP_SEC:=60}"

DB="$DATA_DIR/synthetic-vision.db"

backup_once() {
    [ -s "$DB" ] || { echo "[backup] no DB yet, skipping"; return 0; }
    mkdir -p "$BACKUP_DIR" 2>/dev/null || true

    ts=$(date -u +%Y%m%dT%H%M%SZ)
    tmp="/tmp/sv-$ts.db"

    if ! sqlite3 "$DB" ".backup '$tmp'"; then
        echo "[backup] sqlite3 .backup failed"
        rm -f "$tmp"
        return 1
    fi
    if ! gzip -9 "$tmp"; then
        rm -f "$tmp" "$tmp.gz"
        return 1
    fi

    sz=$(stat -c%s "$tmp.gz" 2>/dev/null || echo 0)
    if [ "$sz" -lt 200 ]; then
        echo "[backup] snapshot suspiciously small ($sz bytes), not uploading"
        rm -f "$tmp.gz"
        return 0
    fi

    # Timestamped snapshot first, then latest via tmp+rename so a reader of
    # latest.db.gz never sees a half-written file.
    if cp "$tmp.gz" "$BACKUP_DIR/sv-$ts.db.gz" \
        && cp "$tmp.gz" "$BACKUP_DIR/latest.db.gz.tmp" \
        && mv "$BACKUP_DIR/latest.db.gz.tmp" "$BACKUP_DIR/latest.db.gz"; then
        echo "[backup] wrote sv-$ts.db.gz ($sz bytes)"
    else
        echo "[backup] copy to $BACKUP_DIR failed (FUSE issue?)"
    fi
    rm -f "$tmp.gz"

    # Images/references: ADDITIVE copy only — deliberately no mirroring or
    # --delete. A failed restore leaves /data empty; a mirroring sync would
    # then wipe the bucket copy too. Orphans (user-deleted images) are cheap
    # and unreachable once their DB row is gone.
    for d in images references; do
        [ -d "$DATA_DIR/$d" ] || continue
        mkdir -p "$BACKUP_DIR/$d"
        for f in "$DATA_DIR/$d"/*; do
            [ -f "$f" ] || continue
            base=${f##*/}
            [ -e "$BACKUP_DIR/$d/$base" ] || cp "$f" "$BACKUP_DIR/$d/$base" \
                || echo "[backup] copy of $d/$base failed (continuing)"
        done
    done

    # Prune timestamped snapshots, keep newest BACKUP_KEEP.
    ls -1t "$BACKUP_DIR"/sv-*.db.gz 2>/dev/null \
        | tail -n +$((BACKUP_KEEP + 1)) \
        | while read -r old; do rm -f "$old"; done
    return 0
}

case "${1:-once}" in
    loop)
        sleep "$BACKUP_WARMUP_SEC"
        while true; do
            backup_once
            # Sleep in 60s slices; dropping a `backup-now` file into the bucket
            # (hf buckets cp) triggers an immediate snapshot — no shell needed.
            slept=0
            while [ "$slept" -lt $((BACKUP_INTERVAL_MIN * 60)) ]; do
                sleep 60
                slept=$((slept + 60))
                if [ -e "$BACKUP_DIR/backup-now" ]; then
                    echo "[backup] manual trigger detected"
                    rm -f "$BACKUP_DIR/backup-now"
                    break
                fi
            done
        done
        ;;
    *)
        backup_once
        ;;
esac
