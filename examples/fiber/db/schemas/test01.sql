CREATE TABLE IF NOT EXISTS test01 (
    id TEXT PRIMARY KEY, -- UUID
    name TEXT DEFAULT '', -- String
    test_int INTEGER, -- Integer
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Timestamp
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Timestamp
);