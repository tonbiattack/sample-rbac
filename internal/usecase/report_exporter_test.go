package usecase

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"sample-rbac/internal/rbac"

	"gorm.io/gorm"
)

const (
	// usecaseテスト専用ID（rbacテストとは分離）
	ucUserID      int64 = 5001
	ucRoleAdminID int64 = 6001
	ucRoleUserID  int64 = 6002
	ucPermExpName       = "report.export"
	ucRoleAdminName     = "usecase_admin"
	ucRoleUserName      = "usecase_user"
)

func TestReportExporter_ExportMonthlyReport_Success(t *testing.T) {
	t.Run("権限があるユーザーは月次レポートを出力できる", func(t *testing.T) {
		// 1) テストDB接続と依存オブジェクトを組み立てます。
		db := setupUsecaseTestDB(t)
		repo := rbac.NewRepository(db)
		ctx := context.Background()

		// 2) users / roles / permissions の基本データを投入します。
		seedUsecaseBase(t, db)
		// 3) ユーザーにadminロール、adminロールにreport.export権限を付与します。
		mustExecUC(t, db, "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", ucUserID, ucRoleAdminID)
		if err := repo.GrantPermissionToRoleByName(ctx, ucRoleAdminID, ucPermExpName); err != nil {
			t.Fatalf("GrantPermissionToRoleByName failed: %v", err)
		}

		// 4) 業務ユースケースを作成します。
		authorizer := NewAuthorizer(repo)
		exporter := NewReportExporter(authorizer)

		// 5) 実行: 権限ありなので成功し、ファイル名が返る想定です。
		fileName, err := exporter.ExportMonthlyReport(ctx, ucUserID)
		if err != nil {
			t.Fatalf("ExportMonthlyReport failed: %v", err)
		}
		// 6) サンプル実装の戻り値を検証します。
		if fileName != "monthly_report.csv" {
			t.Fatalf("unexpected fileName: %s", fileName)
		}
	})
}

func TestReportExporter_ExportMonthlyReport_Forbidden(t *testing.T) {
	t.Run("権限がないユーザーは月次レポートを出力できない", func(t *testing.T) {
		// 1) テストDB接続と依存オブジェクトを組み立てます。
		db := setupUsecaseTestDB(t)
		repo := rbac.NewRepository(db)
		ctx := context.Background()

		// 2) users / roles / permissions の基本データを投入します。
		seedUsecaseBase(t, db)
		// 3) 一般ユーザーロールのみ付与し、report.export権限は付与しません。
		mustExecUC(t, db, "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", ucUserID, ucRoleUserID)

		// 4) 業務ユースケースを作成します。
		authorizer := NewAuthorizer(repo)
		exporter := NewReportExporter(authorizer)

		// 5) 実行: 権限不足のため ErrForbidden を期待します。
		_, err := exporter.ExportMonthlyReport(ctx, ucUserID)
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}

func setupUsecaseTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// 環境変数があればそちらを優先し、なければローカル既定値を使います。
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "app:app@tcp(127.0.0.1:3306)/sample_rbac?parseTime=true"
	}

	var (
		db  *gorm.DB
		err error
	)

	// MySQLコンテナ起動直後の接続失敗に備えてリトライします。
	for range 10 {
		db, err = rbac.OpenMySQL(dsn)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		// DB未起動環境では失敗扱いではなくスキップします。
		t.Skipf("mysql not ready: %v", err)
	}

	// テスト独立性のため、実行前後で対象データを掃除します。
	cleanupUsecaseTables(t, db)
	t.Cleanup(func() { cleanupUsecaseTables(t, db) })
	return db
}

func seedUsecaseBase(t *testing.T, db *gorm.DB) {
	t.Helper()

	// users: 業務処理の実行ユーザー
	mustExecUC(t, db, "INSERT INTO users (id, email) VALUES (?, ?)", ucUserID, "bob@example.com")
	// roles: テスト専用名で衝突を避けます。
	mustExecUC(t, db, "INSERT INTO roles (id, name) VALUES (?, ?), (?, ?)", ucRoleAdminID, ucRoleAdminName, ucRoleUserID, ucRoleUserName)
	// permissions: report.export が未登録の場合のみ投入します。
	mustExecUC(t, db, "INSERT INTO permissions (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = VALUES(name)", 9204, ucPermExpName)
}

func cleanupUsecaseTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	// 並列テストで衝突しないよう、テストで使ったIDの行だけ削除します。
	// 依存順: role_permissions -> user_roles -> permissions/roles -> users
	mustExecUC(t, db, "DELETE FROM role_permissions WHERE role_id IN (?, ?)", ucRoleAdminID, ucRoleUserID)
	mustExecUC(t, db, "DELETE FROM user_roles WHERE user_id = ? OR role_id IN (?, ?)", ucUserID, ucRoleAdminID, ucRoleUserID)
	mustExecUC(t, db, "DELETE FROM roles WHERE id IN (?, ?)", ucRoleAdminID, ucRoleUserID)
	mustExecUC(t, db, "DELETE FROM users WHERE id = ?", ucUserID)
}

func mustExecUC(t *testing.T, db *gorm.DB, sql string, args ...any) {
	t.Helper()
	// SQL実行失敗時は即テスト終了し、原因SQLを表示します。
	if err := db.Exec(sql, args...).Error; err != nil {
		t.Fatalf("exec failed: %s err=%v", sql, err)
	}
}
