CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE vault_entries (
                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                               title TEXT NOT NULL,
                               username TEXT NOT NULL,
                               password BYTEA NOT NULL,
                               notes TEXT,
                               tags TEXT[],
                               folder TEXT,
                               created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                               updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);