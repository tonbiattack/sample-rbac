-- 文字化け対策: シード投入時も UTF-8 (utf8mb4) を明示
SET NAMES utf8mb4;

-- =============================
-- RBAC サンプルデータ
-- =============================

-- users
INSERT INTO users (id, email) VALUES
  (9001, 'admin@example.com'),
  (9002, 'operator@example.com'),
  (9003, 'viewer@example.com')
ON DUPLICATE KEY UPDATE email = VALUES(email);

-- roles
INSERT INTO roles (id, name) VALUES
  (9101, 'admin'),
  (9102, 'operator'),
  (9103, 'viewer')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- permissions
INSERT INTO permissions (id, name) VALUES
  (9201, 'user.create'),
  (9202, 'user.delete'),
  (9203, 'report.view'),
  (9204, 'report.export'),
  (9205, 'order.cancel')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- role_permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES
  -- admin はすべてのサンプル権限を保持
  (9101, 9201),
  (9101, 9202),
  (9101, 9203),
  (9101, 9204),
  (9101, 9205),
  -- operator は閲覧・エクスポート・取消を保持
  (9102, 9203),
  (9102, 9204),
  (9102, 9205),
  -- viewer は閲覧のみ
  (9103, 9203);

-- user_roles
INSERT IGNORE INTO user_roles (user_id, role_id) VALUES
  (9001, 9101),
  (9002, 9102),
  (9003, 9103);

