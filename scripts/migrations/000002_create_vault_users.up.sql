CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE vault_users (
                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                               email TEXT NOT NULL,
                               username TEXT NOT NULL,
                               password BYTEA NOT NULL,
                               created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                               updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);