# なぜ壊れた権限設計が生まれるのか（RBAC設計の実務）

## はじめに

業務システムの権限管理では、次のような設計をよく見かけます。

- フロントとバックエンドに `role` 定数がある
- DBにも `roles` テーブルがある
- UI制御が `if role == ADMIN` のように書かれている
- role を追加すると複数箇所を修正する必要がある

こうした設計は「悪い設計」として語られることが多いですが、実際には最初から間違っていたわけではないことがほとんどです。

多くの場合、開発の進行とともに局所的に合理的な判断が積み重なった結果として生まれます。

この記事では、なぜそのような設計が生まれるのかを整理します。

## 権限設計は最初はシンプルで良い

多くのシステムは小さく始まります。

例えば初期の権限は次の程度です。

```text
admin
operator
viewer
```

この程度であれば、次のようなコードでも問題ありません。

```go
if user.Role == ADMIN {
    showAdminMenu()
}
```

フロントでも同様です。

```ts
if (role === ADMIN) {
  showSettings()
}
```

この段階では次の条件が揃っています。

- role数が少ない
- UIも単純
- 機能も少ない

そのため、roleベースの判定は合理的です。

## UI制御の都合でroleが肥大化する

しかし機能が増えると、UIの出し分けが増えていきます。

例えば次のような機能です。

- ユーザー作成
- レポート閲覧
- 取引取消
- CSVエクスポート

このとき簡単な方法は、`if (role === ADMIN)` のような判定を増やすことです。

すると role は次をすべて表すようになります。

- 組織
- 権限
- UI制御

結果として、`role = 権限の集合` ではなく、`role = UI制御 + 権限 + 組織` という状態になります。

## role定数が複数箇所に存在するようになる

次によく起きるのが role定数の多重管理です。

フロント

```ts
export const RoleID = {
  ADMIN: 1,
  OPERATOR: 2
}
```

バックエンド

```go
const (
  RoleAdmin = 1
  RoleOperator = 2
)
```

DB

```text
roles
----
1 admin
2 operator
```

この設計の意図は理解できます。

- コード上で型のように扱いたい
- if文で分岐したい
- SQL問い合わせを減らしたい

つまり、開発を楽にするための工夫です。

しかし結果として、フロント、バックエンド、DBの三重管理になります。

## 履歴テーブルや外部キーが構造を固定する

さらに問題を固定化する要因があります。

例えば監査ログテーブルや操作履歴テーブルが `role_id -> roles.id` を外部キー参照している場合、role id がシステム内の識別子として固定されます。

すると、DB、コード、フロントのすべてで同じIDを共有する必要が生まれます。

この段階になると、設計を変更するのが難しくなります。

## 運用で整合性を保つようになる

結果として、運用ルールが生まれます。

「roleを追加するときは」

1. フロントの定数を追加
2. バックエンドの定数を追加
3. DBにレコード追加

つまり、設計ではなく運用で整合性を保つ状態になります。

これは多くの業務システムで見られる状態です。

## 本来の権限管理モデル（RBAC）

一般的な権限管理は RBAC（Role Based Access Control）です。

基本構造は次です。

```text
users
roles
permissions
user_roles
role_permissions
```

ERは次の関係になります。

```text
users
  |
user_roles
  |
roles
  |
role_permissions
  |
permissions
```

ここでは次の関係を守ります。

- permission = 個別権限
- role = permissionの集合

UI制御も role ではなく permission で行います。
permission名は `resource.action` 形式にすると管理しやすくなります。

```text
user.create
user.delete
order.cancel
report.export
```

## RBACのテーブル設計例

実装しやすい最小構成のDDL例です。

```sql
CREATE TABLE users (
  id BIGINT PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE roles (
  id BIGINT PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE permissions (
  id BIGINT PRIMARY KEY,
  name VARCHAR(128) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_roles (
  user_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL,
  assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, role_id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (role_id) REFERENCES roles(id)
);

CREATE TABLE role_permissions (
  role_id BIGINT NOT NULL,
  permission_id BIGINT NOT NULL,
  PRIMARY KEY (role_id, permission_id),
  FOREIGN KEY (role_id) REFERENCES roles(id),
  FOREIGN KEY (permission_id) REFERENCES permissions(id)
);

CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
```

## 権限取得と判定の基本SQL

### ユーザーが持つ権限一覧を取得する

```sql
SELECT DISTINCT p.name
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE ur.user_id = :user_id
ORDER BY p.name;
```

### ユーザーが特定権限を持つか判定する

```sql
SELECT EXISTS (
  SELECT 1
  FROM user_roles ur
  JOIN role_permissions rp ON rp.role_id = ur.role_id
  JOIN permissions p ON p.id = rp.permission_id
  WHERE ur.user_id = :user_id
    AND p.name = :permission_name
) AS has_permission;
```

### ユーザーにロールを付与する

```sql
INSERT INTO user_roles (user_id, role_id)
VALUES (:user_id, :role_id);
```

既に付与済みの可能性がある場合はUPSERTを使います。

```sql
INSERT INTO user_roles (user_id, role_id)
VALUES (:user_id, :role_id)
ON CONFLICT (user_id, role_id) DO NOTHING;
```

### ロールに権限を付与する

```sql
INSERT INTO role_permissions (role_id, permission_id)
VALUES (:role_id, :permission_id)
ON CONFLICT (role_id, permission_id) DO NOTHING;
```

### 権限名からpermission_idを引いて付与する

```sql
INSERT INTO role_permissions (role_id, permission_id)
SELECT :role_id, p.id
FROM permissions p
WHERE p.name = :permission_name
ON CONFLICT (role_id, permission_id) DO NOTHING;
```

## 壊れた設計はなぜ生まれるのか

ここまで整理すると、理由は明確です。

多くの場合、次の小さな合理性の積み重ねで起こります。

- 最初はシンプルだった
- UI実装の都合が優先された
- コードで扱いやすくしたい要求があった
- 外部キーや監査系テーブルでIDが固定化された
- 運用で整合性を維持する形になった

つまり、壊れた設計は間違いから生まれるのではなく、自然な進化の結果として生まれることが多いです。

## Go実装サンプル

この記事の内容をそのまま試せるサンプル実装を公開しています。

https://github.com/tonbiattack/sample-rbac

実装内容の要点は次です。

- 技術: `Go + GORM + MySQL + Docker`
- テーブル: `users / roles / permissions / user_roles / role_permissions`
- 権限判定: `HasPermission`（`EXISTS`クエリ）
- 権限一覧取得: `ListPermissions`（`DISTINCT + ORDER BY`）
- 業務ユースケース例: `report.export` を事前チェックしてから処理実行
- テスト: 実DBを使った統合テスト（テストファースト）

ローカルでの最小実行手順:

```bash
docker compose up -d mysql
go test ./...
```

## まとめ

業務システムの権限設計は次の流れで壊れやすくなります。

1. 最初はroleだけで十分
2. UI制御が増える
3. role定数をコードで共有する
4. DBとコードのIDが固定化する
5. 運用で整合性を保つ

その結果、次の問題が発生します。

- roleが肥大化する
- 定数の多重管理が起きる
- 修正コストが増加する

権限設計は後から修正が難しいため、早い段階でRBACなどの構造を意識しておくことが重要です。
