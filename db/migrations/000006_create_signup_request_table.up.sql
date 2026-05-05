CREATE TABLE IF NOT EXISTS signup_request (
    signup_request_id SERIAL PRIMARY KEY,
    name              TEXT NOT NULL,
    email             TEXT NOT NULL,
    reason            TEXT NOT NULL,
    created           TIMESTAMP NOT NULL DEFAULT NOW()
);
