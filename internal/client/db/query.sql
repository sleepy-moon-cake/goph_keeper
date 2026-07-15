-- name: SaveRecord :one
INSERT INTO records (user_name, record_name, data_type, payload, nonce, sync_status, updated_at)
VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(user_name, record_name) DO UPDATE SET
    data_type = excluded.data_type,
    payload = excluded.payload,
    nonce = excluded.nonce,
    sync_status = excluded.sync_status,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetRecordByUniqueKey :one
SELECT id, user_name, record_name, data_type, payload, nonce, sync_status, created_at, updated_at
FROM records
WHERE user_name = ? AND record_name = ?;

-- name: GetAllRecordsByUsername :many
SELECT id, user_name, record_name, data_type, payload, nonce, sync_status, created_at, updated_at
FROM records
WHERE user_name = ?
ORDER BY created_at DESC;

-- name: DeleteRecord :exec
DELETE FROM records
WHERE user_name = ? AND record_name = ?;

-- name: GetPendingRecords :many
SELECT id, user_name, record_name, data_type, payload, nonce, created_at, updated_at
FROM records
WHERE sync_status = 'pending';

-- name: UpdateSyncStatus :exec
UPDATE records
SET sync_status = 'synced'
WHERE id = ?;
