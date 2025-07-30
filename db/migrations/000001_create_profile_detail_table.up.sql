CREATE TABLE IF NOT EXISTS profile_detail (
    id SERIAL PRIMARY KEY,
    user_id TEXT,
    email TEXT,
    first_name TEXT,
    last_name TEXT,
    address_1 TEXT,
    address_2  TEXT,
    address_city TEXT,
    address_state TEXT,
    zip_code TEXT,
    country TEXT,
    social_accounts TEXT[][2]
);
