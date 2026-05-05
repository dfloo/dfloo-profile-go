CREATE TABLE IF NOT EXISTS meeting_request (
    meeting_request_id SERIAL PRIMARY KEY,
    name               TEXT NOT NULL,
    email              TEXT NOT NULL,
    message            TEXT NOT NULL,
    created            TIMESTAMP NOT NULL DEFAULT NOW()
);
