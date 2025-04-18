-- db/schema/schema.sql

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    picture TEXT,
    role TEXT NOT NULL DEFAULT 'applicant' CHECK (role IN ('admin', 'recruiter', 'applicant', 'pending')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE companies (
    id UUID PRIMARY KEY,
    recruiter_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    logo TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE job_postings (
    id UUID PRIMARY KEY,
    recruiter_id UUID REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    company_name TEXT NOT NULL,
    position TEXT NOT NULL,
    skills TEXT[],
    description TEXT,
    salary TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE applicant_skill_sets (
    applicant_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    skills TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT now()
);