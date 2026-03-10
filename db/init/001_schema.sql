-- 文字化け対策: 接続文字コードを UTF-8 (utf8mb4) に固定
SET NAMES utf8mb4;

-- 文字化け対策: DB既定の文字コード/照合順序を明示
ALTER DATABASE sample_rbac CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

-- users: システム利用者の基本情報を保持するテーブル
CREATE TABLE users (
  -- ユーザー識別子（アプリケーション側で採番する想定）
  id BIGINT PRIMARY KEY COMMENT 'ユーザーID（主キー）',
  -- ログインや通知に使うメールアドレス（重複不可）
  email VARCHAR(255) NOT NULL UNIQUE COMMENT 'ユーザーメールアドレス（一意）',
  -- レコード作成日時
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時'
) COMMENT = 'ユーザーのマスタ'
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci;

-- roles: permission の集合を表すロール定義テーブル
CREATE TABLE roles (
  -- ロール識別子
  id BIGINT PRIMARY KEY COMMENT 'ロールID（主キー）',
  -- ロール名（例: admin, viewer）
  name VARCHAR(64) NOT NULL UNIQUE COMMENT 'ロール名（一意）',
  -- レコード作成日時
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時'
) COMMENT = 'ロールのマスタ'
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci;

-- permissions: 個別操作権限の定義テーブル
CREATE TABLE permissions (
  -- 権限識別子
  id BIGINT PRIMARY KEY COMMENT '権限ID（主キー）',
  -- 権限名（例: report.export, user.create）
  name VARCHAR(128) NOT NULL UNIQUE COMMENT '権限名（一意）',
  -- レコード作成日時
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時'
) COMMENT = '権限のマスタ'
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci;

-- user_roles: ユーザーとロールの多対多を表す中間テーブル
CREATE TABLE user_roles (
  -- 付与先ユーザーID
  user_id BIGINT NOT NULL COMMENT 'ユーザーID（users.id 参照）',
  -- 付与するロールID
  role_id BIGINT NOT NULL COMMENT 'ロールID（roles.id 参照）',
  -- ロールを付与した日時
  assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'ロール付与日時',
  -- 同一ユーザーに同一ロールを二重付与しないための複合主キー
  PRIMARY KEY (user_id, role_id),
  -- 参照整合性: 存在しないユーザーへの付与を防止
  CONSTRAINT fk_user_roles_user_id FOREIGN KEY (user_id) REFERENCES users(id),
  -- 参照整合性: 存在しないロールの付与を防止
  CONSTRAINT fk_user_roles_role_id FOREIGN KEY (role_id) REFERENCES roles(id)
) COMMENT = 'ユーザーとロールの関連（多対多）'
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci;

-- role_permissions: ロールと権限の多対多を表す中間テーブル
CREATE TABLE role_permissions (
  -- 権限を持つロールID
  role_id BIGINT NOT NULL COMMENT 'ロールID（roles.id 参照）',
  -- ロールに紐づける権限ID
  permission_id BIGINT NOT NULL COMMENT '権限ID（permissions.id 参照）',
  -- 同一ロールに同一権限を二重付与しないための複合主キー
  PRIMARY KEY (role_id, permission_id),
  -- 参照整合性: 存在しないロールへの紐付けを防止
  CONSTRAINT fk_role_permissions_role_id FOREIGN KEY (role_id) REFERENCES roles(id),
  -- 参照整合性: 存在しない権限への紐付けを防止
  CONSTRAINT fk_role_permissions_permission_id FOREIGN KEY (permission_id) REFERENCES permissions(id)
) COMMENT = 'ロールと権限の関連（多対多）'
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci;

-- user_id で絞った後に role_id 側から参照するクエリを補助
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
-- role_id で絞った後に permission_id 側から参照するクエリを補助
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

