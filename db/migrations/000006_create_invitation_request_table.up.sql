CREATE TABLE IF NOT EXISTS invitation_request (
    invitation_request_id SERIAL PRIMARY KEY,
    name                  TEXT NOT NULL,
    email                 TEXT NOT NULL,
    reason                TEXT NOT NULL,
    created               TIMESTAMP NOT NULL DEFAULT NOW()
);
