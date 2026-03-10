-- 文字化け対策: シード投入時も UTF-8 (utf8mb4) を明示
SET NAMES utf8mb4;

-- =============================
-- RBAC サンプルデータ
-- =============================
-- このファイルは「最小だけど現実的」な権限構成を試せるように、
-- 3ユーザー / 3ロール / 5権限のサンプルを投入します。
--
-- 想定する役割:
-- - admin    : ほぼ全操作可能
-- - operator : 日常運用に必要な操作を実行可能
-- - viewer   : 閲覧のみ可能
--
-- 参照関係の流れ:
-- users -> user_roles -> roles -> role_permissions -> permissions
--
-- 再実行を考慮:
-- - マスタ（users/roles/permissions）は ON DUPLICATE KEY UPDATE を使い冪等化
-- - 関連（user_roles/role_permissions）は INSERT IGNORE で重複追加を回避

-- ----------------------------------------------------------
-- users: システム利用者
-- ----------------------------------------------------------
-- 9001: 管理者
-- 9002: 運用担当者
-- 9003: 閲覧専用ユーザー
INSERT INTO users (id, email) VALUES
  (9001, 'admin@example.com'),
  (9002, 'operator@example.com'),
  (9003, 'viewer@example.com')
ON DUPLICATE KEY UPDATE email = VALUES(email);

-- ----------------------------------------------------------
-- roles: 業務上の役割（permission の集合）
-- ----------------------------------------------------------
INSERT INTO roles (id, name) VALUES
  (9101, 'admin'),
  (9102, 'operator'),
  (9103, 'viewer')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- ----------------------------------------------------------
-- permissions: 個別操作権限
-- ----------------------------------------------------------
-- 命名ルールは resource.action を採用
-- 例:
-- - user.create   : ユーザー作成
-- - user.delete   : ユーザー削除
-- - report.view   : レポート閲覧
-- - report.export : レポート出力
-- - order.cancel  : 取引取消
INSERT INTO permissions (id, name) VALUES
  (9201, 'user.create'),
  (9202, 'user.delete'),
  (9203, 'report.view'),
  (9204, 'report.export'),
  (9205, 'order.cancel')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- ----------------------------------------------------------
-- role_permissions: ロールに対する権限割り当て
-- ----------------------------------------------------------
-- ポイント:
-- - 1ロールに複数権限を持たせる（role = permission集合）
-- - 同じ権限を複数ロールで共有できる
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

-- ----------------------------------------------------------
-- user_roles: ユーザーにロールを割り当て
-- ----------------------------------------------------------
-- ポイント:
-- - ユーザーは複数ロールを持てる（このサンプルでは1ユーザー1ロール）
-- - 将来的に兼務を表現する場合は同じ user_id に複数 role_id を追加する
INSERT IGNORE INTO user_roles (user_id, role_id) VALUES
  (9001, 9101),
  (9002, 9102),
  (9003, 9103);

-- ----------------------------------------------------------
-- 動作確認用SQL（必要時に手動実行）
-- ----------------------------------------------------------
-- ユーザー 9002 (operator) の権限一覧を確認
-- SELECT DISTINCT p.name
-- FROM user_roles ur
-- JOIN role_permissions rp ON rp.role_id = ur.role_id
-- JOIN permissions p ON p.id = rp.permission_id
-- WHERE ur.user_id = 9002
-- ORDER BY p.name;

