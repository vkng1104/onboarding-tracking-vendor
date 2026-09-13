CREATE TABLE coordinators (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(btrim(name)) > 0),
    email TEXT NOT NULL UNIQUE CHECK (email = lower(email) AND length(btrim(email)) > 0),
    password_hash TEXT NOT NULL CHECK (length(password_hash) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
