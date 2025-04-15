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