# sample-rbac

MySQL + Go + GORM + Docker で作る、最小のRBACサンプルです。

この記事 `docs/なぜ壊れた権限設計が生まれるのか（RBAC設計の実務）.md` のテーブル設計をそのまま採用しています。

## テーブル設計（記事と同一）

- `users`
- `roles`
- `permissions`
- `user_roles`
- `role_permissions`

DDL は [db/init/001_schema.sql](/c:/apps/sample-rbac/db/init/001_schema.sql) を参照してください。

## セットアップ

```bash
docker compose up -d mysql
```

MySQL接続情報（デフォルト）:

- host: `127.0.0.1`
- port: `3306`
- database: `sample_rbac`
- user: `app`
- password: `app`

## テストファーストで作った内容

先に失敗テストを作成し、後から実装しています。

対象テスト:

- `HasPermission`: ユーザーが特定権限を持つか
- `HasPermission_FalseWhenNotGranted`: 権限未付与時の判定
- `ListPermissions_DistinctSorted`: 権限一覧の重複排除とソート

実行:

```bash
go test ./...
```

## 実装ファイル

- [internal/rbac/mysql.go](/c:/apps/sample-rbac/internal/rbac/mysql.go): GORMでMySQL接続を開く
- [internal/rbac/repository.go](/c:/apps/sample-rbac/internal/rbac/repository.go): RBAC用クエリ実装
- [internal/rbac/repository_test.go](/c:/apps/sample-rbac/internal/rbac/repository_test.go): 実DB統合テスト

## 片付け

```bash
docker compose down
```

データも削除する場合:

```bash
docker compose down -v
```
