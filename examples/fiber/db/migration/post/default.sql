--
-- Apache License 2.0
-- Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
-- Contributors: Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
--
-- ======================================================================
--                          DEFAULT DATA
-- ======================================================================

-- Insert default user
INSERT OR IGNORE INTO users (username, full_name, email, enabled, test_int) VALUES
  ('admin', 'Admin User', 'admin@example.com', 1, 42);
