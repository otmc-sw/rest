-- Initial data for users
INSERT OR IGNORE INTO users (id, name, email, role_id) VALUES
(1, 'Admin User', 'admin@otmc.com', 1),
(2, 'John Doe', 'john@example.com', 2),
(3, 'Jane Smith', 'jane@example.com', 2),
(4, 'Guest User', 'guest@example.com', 3);
