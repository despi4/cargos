CREATE TABLE profiles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE,
    name TEXT NOT NULL CHECK (length(btrim(name)) > 0),
    age INTEGER CHECK (age BETWEEN 0 AND 120),
    skills TEXT[] NOT NULL DEFAULT '{}',
    positions TEXT[] NOT NULL DEFAULT '{}',
    location TEXT NOT NULL DEFAULT '',
    experience INTEGER NOT NULL DEFAULT 0 CHECK (experience >= 0),
    education TEXT NOT NULL DEFAULT '',
    preferences TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- user_id belongs to Auth Service; a cross-service foreign key is intentionally absent.
