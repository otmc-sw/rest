-- Initial data for roles

-- Clear existing data first
DELETE FROM role;


-- Insert initial roles
INSERT OR IGNORE INTO role (id, name, description, status, created_at, updated_at) VALUES
(1, 'admin', 'Administrator with full access', 'active', datetime('now'), datetime('now')),
(2, 'user', 'Regular user with limited access', 'active', datetime('now'), datetime('now')),
(3, 'guest', 'Guest user with read-only access', 'active', datetime('now'), datetime('now'));
