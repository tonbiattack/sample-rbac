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

## 実際の処理でのユースケース

このサンプルでは、業務処理の入口で権限を確認します。

```go
db, _ := rbac.OpenMySQL("app:app@tcp(127.0.0.1:3306)/sample_rbac?parseTime=true")
repo := rbac.NewRepository(db)
authorizer := usecase.NewAuthorizer(repo)
exporter := usecase.NewReportExporter(authorizer)

fileName, err := exporter.ExportMonthlyReport(ctx, userID)
if errors.Is(err, usecase.ErrForbidden) {
    // 403 相当の扱い
}
```

処理の流れ:

1. `ReportExporter.ExportMonthlyReport` が呼ばれる
2. `Authorizer.Require` が `report.export` をチェックする
3. 内部で `Repository.HasPermission` が DB から権限を判定する
4. 権限があれば業務処理を続行、なければ `ErrForbidden`

## テストファーストで作った内容

先に失敗テストを作成し、後から実装しています。

対象テスト:

- `Repository` の統合テスト
  - `HasPermission`: ユーザーが特定権限を持つか
  - `HasPermission_FalseWhenNotGranted`: 権限未付与時の判定
  - `ListPermissions_DistinctSorted`: 権限一覧の重複排除とソート
- `Usecase` の統合テスト
  - `ExportMonthlyReport_Success`: 権限ありで成功
  - `ExportMonthlyReport_Forbidden`: 権限なしで拒否

実行:

```bash
go test ./...
```

## 実装ファイル

- [internal/rbac/mysql.go](/c:/apps/sample-rbac/internal/rbac/mysql.go): GORMでMySQL接続を開く
- [internal/rbac/repository.go](/c:/apps/sample-rbac/internal/rbac/repository.go): RBAC用クエリ実装
- [internal/rbac/repository_test.go](/c:/apps/sample-rbac/internal/rbac/repository_test.go): Repositoryの実DB統合テスト
- [internal/usecase/authorizer.go](/c:/apps/sample-rbac/internal/usecase/authorizer.go): 権限判定ユースケース
- [internal/usecase/report_exporter.go](/c:/apps/sample-rbac/internal/usecase/report_exporter.go): 業務ユースケース例
- [internal/usecase/report_exporter_test.go](/c:/apps/sample-rbac/internal/usecase/report_exporter_test.go): Usecaseの実DB統合テスト

## 片付け

```bash
docker compose down
```

データも削除する場合:

```bash
docker compose down -v
```
