ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'applicant' CHECK (role IN ('admin', 'recruiter', 'applicant')); 