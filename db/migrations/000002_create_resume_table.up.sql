CREATE TABLE IF NOT EXISTS resume (
    resume_id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    profile_id INTEGER NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profile(profile_id),
    sections TEXT[],
    description TEXT,
    defaultResume BOOLEAN NOT NULL DEFAULT FALSE,
    summary TEXT,
    skills TEXT[],
    experience JSONB,
    education JSONB,
    file_name TEXT,
    template_settings JSONB,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    updated TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX only_one_default
ON resume((TRUE))
WHERE defaultResume = TRUE;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_resume_updated'
    ) THEN
        EXECUTE 'DROP TRIGGER set_resume_updated ON resume;';
    END IF;
END;
$$;

CREATE TRIGGER set_resume_updated
BEFORE UPDATE ON resume
FOR EACH ROW
EXECUTE FUNCTION update_updated_column();
