CREATE TABLE IF NOT EXISTS users (
    id INTEGER,
    user_name TEXT PRIMARY KEY NOT NULL, 
    password_hash TEXT NOT NULL 
);

CREATE TABLE IF NOT EXISTS records (
    id INTEGER PRIMARY KEY AUTOINCREMENT, 
    user_name TEXT NOT NULL,
    record_name TEXT NOT NULL,
    data_type TEXT NOT NULL,
    payload BLOB NOT NULL,
    nonce BLOB NOT NULL,
    sync_status TEXT NOT NULL DEFAULT 'synced',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_name, record_name),
    FOREIGN KEY(user_name) REFERENCES users(user_name) ON DELETE CASCADE 
);
