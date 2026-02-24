ALTER TABLE job_application
DROP COLUMN IF EXISTS snapshot,
DROP COLUMN IF EXISTS source_url;
