CREATE TABLE IF NOT EXISTS job_application (
    job_application_id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    resume_id INTEGER NULL,
    FOREIGN KEY (resume_id) REFERENCES resume(resume_id),
    status TEXT NOT NULL,
    sort_index INTEGER NOT NULL,
    company TEXT,
    role TEXT,
    description TEXT,
    notes TEXT,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    updated TIMESTAMP NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_job_application_updated'
    ) THEN
        EXECUTE 'DROP TRIGGER set_job_application_updated ON job_application;';
    END IF;
END;
$$;

CREATE TRIGGER set_job_application_updated
BEFORE UPDATE ON job_application
FOR EACH ROW
EXECUTE FUNCTION update_updated_column();
