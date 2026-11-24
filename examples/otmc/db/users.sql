-- Initial data for users

-- Clear existing data first
DELETE FROM users;

-- Insert initial users
INSERT OR IGNORE INTO users (id, username, email, full_name, status, role_id) VALUES
(1, 'admin', 'admin@otmc.com', 'Admin User', 'active', 1),
(2, 'john_doe', 'john@example.com', 'John Doe', 'active', 2),
(3, 'jane_smith', 'jane@example.com', 'Jane Smith', 'active', 2),
(4, 'guest', 'guest@example.com', 'Guest User', 'active', 3);
