-- db/schema/schema.sql

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    picture TEXT,
    role TEXT NOT NULL DEFAULT 'applicant' CHECK (role IN ('admin', 'recruiter', 'applicant', 'pending')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT now()
);