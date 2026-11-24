-- Initial data for roles

-- Clear existing data first
DELETE FROM roles;


-- Insert initial roles
INSERT OR IGNORE INTO roles (id, name, description, status) VALUES
(1, 'admin', 'Administrator with full access', 'active'),
(2, 'user', 'Regular user with limited access', 'active'),
(3, 'guest', 'Guest user with read-only access', 'active');
