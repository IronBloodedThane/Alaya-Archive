# Backup and Restore

Two complementary layers protect the SQLite database that lives on the
GCS-mounted Cloud Run volume.

## Layer 1: Object versioning on the live DB bucket

Continuous protection against accidental delete or corrupted overwrites.
Every modification to `alaya-archive.db` automatically saves the previous
contents as a non-current version, kept for 30 days, then garbage-collected.

### One-time setup

Replace `<DB_BUCKET>` with the bucket name (the value you set as
`DB_BUCKET` in GitHub Actions secrets).

```bash
# Enable object versioning
gcloud storage buckets update gs://<DB_BUCKET> --versioning

# Lifecycle: delete non-current versions after 30 days
cat > /tmp/db-lifecycle.json <<'EOF'
{
  "rule": [
    {
      "action": { "type": "Delete" },
      "condition": { "daysSinceNoncurrentTime": 30, "isLive": false }
    }
  ]
}
EOF
gcloud storage buckets update gs://<DB_BUCKET> --lifecycle-file=/tmp/db-lifecycle.json
```

### Restore from a non-current version

```bash
# List all versions of the DB file
gcloud storage objects list gs://<DB_BUCKET>/alaya-archive.db --all-versions

# Each row shows a generation number. To restore one:
gcloud storage cp \
  gs://<DB_BUCKET>/alaya-archive.db#<GENERATION> \
  gs://<DB_BUCKET>/alaya-archive.db
```

The Cloud Run service will pick up the restored file on next request (the
volume is mounted live).

## Layer 2: Daily external snapshots

A scheduled GitHub Actions workflow (`.github/workflows/backup.yml`) runs
every day at 05:00 UTC and copies the live DB to a separate backups bucket
with a timestamp. This protects against the entire live bucket being wiped
(operator error, IAM mistake, etc).

### One-time setup

```bash
# Create a backups bucket
gcloud storage buckets create gs://<BACKUP_BUCKET> \
  --location=<REGION> \
  --uniform-bucket-level-access

# Lifecycle: delete snapshots older than 90 days
cat > /tmp/backup-lifecycle.json <<'EOF'
{
  "rule": [
    {
      "action": { "type": "Delete" },
      "condition": { "age": 90 }
    }
  ]
}
EOF
gcloud storage buckets update gs://<BACKUP_BUCKET> --lifecycle-file=/tmp/backup-lifecycle.json

# Grant the deploy service account write access
gcloud storage buckets add-iam-policy-binding gs://<BACKUP_BUCKET> \
  --member="serviceAccount:<DEPLOY_SA_EMAIL>" \
  --role=roles/storage.objectAdmin
```

Then add `BACKUP_BUCKET` to **GitHub repo → Settings → Secrets and variables → Actions**.

The first scheduled run will fire the next morning at 05:00 UTC, or trigger
it immediately from the **Actions → Backup database → Run workflow** button.

### Restore from a daily snapshot

```bash
# List recent snapshots
gcloud storage ls gs://<BACKUP_BUCKET>/

# Restore a specific one to the live location
gcloud storage cp \
  gs://<BACKUP_BUCKET>/alaya-archive-<STAMP>.db \
  gs://<DB_BUCKET>/alaya-archive.db
```

## Limitations to be aware of

- **Snapshots are not strictly atomic.** The copy reads the file while Cloud
  Run may still write to it. For our single-user scale and infrequent writes
  the risk is small, and the previous day's snapshot remains as a fallback
  if today's is corrupt. If we ever need stronger guarantees, the upgrade is
  a Cloud Run Job that opens the DB and uses SQLite `VACUUM INTO` before
  uploading.

- **The "we delete your data when you ask" promise on the About page extends
  to backups.** Lifecycle rules expire backup files within 90 days, so a
  deleted account is fully gone (live + every backup) within ~90 days. If
  you ever need a faster guarantee, shorten the lifecycle.
