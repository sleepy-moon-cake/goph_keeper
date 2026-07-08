-- ============================================================================
-- ЗАПРОСЫ ДЛЯ ТАБЛИЦЫ USERS
-- ============================================================================

-- name: CreateUser :one
INSERT INTO users (user_name, password_hash)
VALUES ($1, $2)
RETURNING user_name, password_hash;

-- name: GetUserByUsername :one
SELECT user_name, password_hash
FROM users
WHERE user_name = $1;


-- ============================================================================
-- ЗАПРОСЫ ДЛЯ ТАБЛИЦЫ RECORDS
-- ============================================================================

-- name: CreateRecord :one
INSERT INTO records (user_name, record_name, data_type, payload, nonce)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_name, record_name, data_type, payload, nonce, created_at, updated_at;

-- name: GetRecordByUniqueKey :one
SELECT id, user_name, record_name, data_type, payload, nonce, created_at, updated_at
FROM records
WHERE user_name = $1 AND record_name = $2;

-- name: GetAllRecordsByUsername :many
SELECT id, user_name, record_name, data_type, payload, nonce, created_at, updated_at
FROM records
WHERE user_name = $1
ORDER BY created_at DESC;

-- name: UpdateRecordData :one
UPDATE records
SET data_type = $3,
    payload = $4,
    nonce = $5,
    updated_at = NOW()
WHERE user_name = $1 AND record_name = $2
RETURNING id, user_name, record_name, data_type, payload, nonce, created_at, updated_at;

-- name: DeleteRecord :exec
DELETE FROM records
WHERE user_name = $1 AND record_name = $2;
