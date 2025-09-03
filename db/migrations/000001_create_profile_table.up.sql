CREATE TABLE IF NOT EXISTS profile (
    profile_id SERIAL PRIMARY KEY,
    user_id TEXT,
    resume_id TEXT,
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
    social_accounts TEXT[],
    UNIQUE (user_id, resume_id)
);
