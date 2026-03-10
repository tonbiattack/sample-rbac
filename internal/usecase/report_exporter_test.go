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
	ucUserID      int64 = 5001
	ucRoleAdminID int64 = 6001
	ucRoleUserID  int64 = 6002
	ucPermExpID   int64 = 7001
)

func TestReportExporter_ExportMonthlyReport_Success(t *testing.T) {
	t.Run("権限があるユーザーは月次レポートを出力できる", func(t *testing.T) {
		db := setupUsecaseTestDB(t)
		repo := rbac.NewRepository(db)
		ctx := context.Background()

		seedUsecaseBase(t, db)
		mustExecUC(t, db, "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", ucUserID, ucRoleAdminID)
		mustExecUC(t, db, "INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", ucRoleAdminID, ucPermExpID)

		authorizer := NewAuthorizer(repo)
		exporter := NewReportExporter(authorizer)

		fileName, err := exporter.ExportMonthlyReport(ctx, ucUserID)
		if err != nil {
			t.Fatalf("ExportMonthlyReport failed: %v", err)
		}
		if fileName != "monthly_report.csv" {
			t.Fatalf("unexpected fileName: %s", fileName)
		}
	})
}

func TestReportExporter_ExportMonthlyReport_Forbidden(t *testing.T) {
	t.Run("権限がないユーザーは月次レポートを出力できない", func(t *testing.T) {
		db := setupUsecaseTestDB(t)
		repo := rbac.NewRepository(db)
		ctx := context.Background()

		seedUsecaseBase(t, db)
		mustExecUC(t, db, "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", ucUserID, ucRoleUserID)

		authorizer := NewAuthorizer(repo)
		exporter := NewReportExporter(authorizer)

		_, err := exporter.ExportMonthlyReport(ctx, ucUserID)
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}

func setupUsecaseTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "app:app@tcp(127.0.0.1:3306)/sample_rbac?parseTime=true"
	}

	var (
		db  *gorm.DB
		err error
	)

	for range 10 {
		db, err = rbac.OpenMySQL(dsn)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		t.Skipf("mysql not ready: %v", err)
	}

	cleanupUsecaseTables(t, db)
	t.Cleanup(func() { cleanupUsecaseTables(t, db) })
	return db
}

func seedUsecaseBase(t *testing.T, db *gorm.DB) {
	t.Helper()

	mustExecUC(t, db, "INSERT INTO users (id, email) VALUES (?, ?)", ucUserID, "bob@example.com")
	mustExecUC(t, db, "INSERT INTO roles (id, name) VALUES (?, ?), (?, ?)", ucRoleAdminID, "admin", ucRoleUserID, "user")
	mustExecUC(t, db, "INSERT INTO permissions (id, name) VALUES (?, ?)", ucPermExpID, "report.export")
}

func cleanupUsecaseTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	// 並列テストで衝突しないよう、テストで使ったIDの行だけ削除します。
	mustExecUC(t, db, "DELETE FROM role_permissions WHERE role_id IN (?, ?) OR permission_id = ?", ucRoleAdminID, ucRoleUserID, ucPermExpID)
	mustExecUC(t, db, "DELETE FROM user_roles WHERE user_id = ? OR role_id IN (?, ?)", ucUserID, ucRoleAdminID, ucRoleUserID)
	mustExecUC(t, db, "DELETE FROM permissions WHERE id = ?", ucPermExpID)
	mustExecUC(t, db, "DELETE FROM roles WHERE id IN (?, ?)", ucRoleAdminID, ucRoleUserID)
	mustExecUC(t, db, "DELETE FROM users WHERE id = ?", ucUserID)
}

func mustExecUC(t *testing.T, db *gorm.DB, sql string, args ...any) {
	t.Helper()
	if err := db.Exec(sql, args...).Error; err != nil {
		t.Fatalf("exec failed: %s err=%v", sql, err)
	}
}
