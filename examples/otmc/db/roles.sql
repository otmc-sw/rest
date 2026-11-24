-- Initial data for roles
INSERT OR IGNORE INTO roles (id, name, description) VALUES
(1, 'admin', 'Administrator with full access'),
(2, 'user', 'Regular user with limited access'),
(3, 'guest', 'Guest user with read-only access');
