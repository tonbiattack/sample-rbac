package rbac

import (
	"context"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
)

const (
	testUserID       int64 = 1001
	testRoleAdminID  int64 = 2001
	testRoleViewerID int64 = 2002
	testPermExportID int64 = 3001
	testPermViewID   int64 = 3002
)

func TestRepository_HasPermission(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	seedBase(t, db)
	if err := repo.AssignRoleToUser(ctx, testUserID, testRoleAdminID); err != nil {
		t.Fatalf("AssignRoleToUser failed: %v", err)
	}
	if err := repo.GrantPermissionToRole(ctx, testRoleAdminID, testPermExportID); err != nil {
		t.Fatalf("GrantPermissionToRole failed: %v", err)
	}

	has, err := repo.HasPermission(ctx, testUserID, "report.export")
	if err != nil {
		t.Fatalf("HasPermission failed: %v", err)
	}
	if !has {
		t.Fatal("expected has permission")
	}
}

func TestRepository_HasPermission_FalseWhenNotGranted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	seedBase(t, db)
	if err := repo.AssignRoleToUser(ctx, testUserID, testRoleViewerID); err != nil {
		t.Fatalf("AssignRoleToUser failed: %v", err)
	}
	if err := repo.GrantPermissionToRole(ctx, testRoleViewerID, testPermViewID); err != nil {
		t.Fatalf("GrantPermissionToRole failed: %v", err)
	}

	has, err := repo.HasPermission(ctx, testUserID, "report.export")
	if err != nil {
		t.Fatalf("HasPermission failed: %v", err)
	}
	if has {
		t.Fatal("expected no permission")
	}
}

func TestRepository_ListPermissions_DistinctSorted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	seedBase(t, db)
	if err := repo.AssignRoleToUser(ctx, testUserID, testRoleAdminID); err != nil {
		t.Fatalf("AssignRoleToUser failed: %v", err)
	}
	if err := repo.AssignRoleToUser(ctx, testUserID, testRoleViewerID); err != nil {
		t.Fatalf("AssignRoleToUser failed: %v", err)
	}
	if err := repo.GrantPermissionToRole(ctx, testRoleAdminID, testPermExportID); err != nil {
		t.Fatalf("GrantPermissionToRole failed: %v", err)
	}
	if err := repo.GrantPermissionToRole(ctx, testRoleViewerID, testPermViewID); err != nil {
		t.Fatalf("GrantPermissionToRole failed: %v", err)
	}

	permissions, err := repo.ListPermissions(ctx, testUserID)
	if err != nil {
		t.Fatalf("ListPermissions failed: %v", err)
	}

	if len(permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d (%v)", len(permissions), permissions)
	}
	if permissions[0] != "report.export" || permissions[1] != "report.view" {
		t.Fatalf("unexpected permissions: %v", permissions)
	}
}

func setupTestDB(t *testing.T) *gorm.DB {
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
		db, err = OpenMySQL(dsn)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		t.Skipf("mysql not ready: %v", err)
	}

	cleanupTables(t, db)
	t.Cleanup(func() { cleanupTables(t, db) })
	return db
}

func seedBase(t *testing.T, db *gorm.DB) {
	t.Helper()

	mustExec(t, db, "INSERT INTO users (id, email) VALUES (?, ?)", testUserID, "alice@example.com")
	mustExec(t, db, "INSERT INTO roles (id, name) VALUES (?, ?), (?, ?)", testRoleAdminID, "admin", testRoleViewerID, "viewer")
	mustExec(t, db, "INSERT INTO permissions (id, name) VALUES (?, ?), (?, ?)", testPermExportID, "report.export", testPermViewID, "report.view")
}

func cleanupTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	mustExec(t, db, "DELETE FROM role_permissions")
	mustExec(t, db, "DELETE FROM user_roles")
	mustExec(t, db, "DELETE FROM permissions")
	mustExec(t, db, "DELETE FROM roles")
	mustExec(t, db, "DELETE FROM users")
}

func mustExec(t *testing.T, db *gorm.DB, sql string, args ...any) {
	t.Helper()
	if err := db.Exec(sql, args...).Error; err != nil {
		t.Fatalf("exec failed: %s err=%v", sql, err)
	}
}
