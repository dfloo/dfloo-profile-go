CREATE TABLE IF NOT EXISTS profile (
    profile_id SERIAL PRIMARY KEY,
    user_id TEXT,
    phone_number TEXT,
    email TEXT,
    first_name TEXT,
    middle_name TEXT,
    last_name TEXT,
    address_1 TEXT,
    address_2  TEXT,
    city TEXT,
    state TEXT,
    zip_code TEXT,
    country TEXT,
    social_accounts JSONB,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    updated TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id)
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_profile_updated'
    ) THEN
        EXECUTE 'DROP TRIGGER set_profile_updated ON profile;';
    END IF;
END;
$$;

CREATE TRIGGER set_profile_updated
BEFORE UPDATE ON profile
FOR EACH ROW
EXECUTE FUNCTION update_updated_column();
