-- name: CreateProfile :one
INSERT INTO profiles (
    user_id, name, age, skills, positions, location, experience, education, preferences
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetProfileByID :one
SELECT id, user_id, name, age, skills, positions, location, experience, education, preferences, created_at, updated_at
FROM profiles
WHERE id = $1 LIMIT 1;

-- name: GetProfileByUserID :one
SELECT id, user_id, name, age, skills, positions, location, experience, education, preferences, created_at, updated_at
FROM profiles
WHERE user_id = $1 LIMIT 1;

-- name: UpdateProfile :exec
UPDATE profiles 
SET 
    name = COALESCE(sqlc.narg('name'), name),
    updated_at = NOW()
WHERE user_id = $1;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE id = $1;