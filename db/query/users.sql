-- name: CreateUser :one
INSERT INTO users (name, email, picture, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateOrUpdateSession :one
INSERT INTO sessions (user_id, token)
VALUES ($1, $2)
ON CONFLICT (token) 
DO UPDATE SET 
    user_id = EXCLUDED.user_id,
    created_at = now()
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions WHERE token = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE user_id = $1 AND token = $2;

-- name: GetAllUsers :many
SELECT id, name, email, picture, role FROM users WHERE role IN ('applicant', 'recruiter');

-- name: GetPendingRecruiters :many
SELECT id, name, email, picture FROM users WHERE role = 'pending';

-- name: ApproveRecruiter :exec
UPDATE users SET role = 'recruiter' WHERE id = $1;

-- name: RejectRecruiter :exec
DELETE FROM users WHERE id = $1;

