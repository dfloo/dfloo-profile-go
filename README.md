# dfloo-profile-go

## F1 Dataset Setup

This repository includes CSV source files in [db/f1-data](db/f1-data) and schema migrations for matching `f1_` tables.

### Local database migration

```bash
docker compose run --rm migrate
```

### One-time F1 data load (manual)

```bash
./scripts/load-f1-data.sh
```

The loader is idempotent by table population state:

1. If all `f1_` tables are empty, it imports all CSV files.
2. If all `f1_` tables already contain data, it exits without changes.
3. If only some tables contain data, it fails to prevent partial duplicate loads.

### Docker Compose seed profile

```bash
docker compose --profile seed up f1-loader
```

The `f1-loader` service runs after successful `migrate` completion and reads from `/db/f1-data` packaged into the image.
