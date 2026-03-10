package rbac

import (
	"context"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
)

const (
	// 固定IDを使い、テストデータを読みやすく・再現しやすくします。
	testUserID       int64 = 1001
	testRoleAdminID  int64 = 2001
	testRoleViewerID int64 = 2002
	testPermExportID int64 = 3001
	testPermViewID   int64 = 3002
	// 既存データと衝突しにくいよう、サンプル専用プレフィックスを付けます。
	testRoleAdminName  = "rbac_admin"
	testRoleViewerName = "rbac_viewer"
	testPermExportName = "rbac.report.export"
	testPermViewName   = "rbac.report.view"
)

func TestRepository_HasPermission(t *testing.T) {
	t.Run("ユーザーが権限を持つ場合はtrueを返す", func(t *testing.T) {
		// 1) テスト用DB接続とRepositoryを準備します。
		db := setupTestDB(t)
		repo := NewRepository(db)
		ctx := context.Background()

		// 2) users / roles / permissions の基本データを投入します。
		seedBase(t, db)
		// 3) ユーザーにadminロールを付与します。
		if err := repo.AssignRoleToUser(ctx, testUserID, testRoleAdminID); err != nil {
			t.Fatalf("AssignRoleToUser failed: %v", err)
		}
		// 4) adminロールにエクスポート権限を付与します。
		if err := repo.GrantPermissionToRole(ctx, testRoleAdminID, testPermExportID); err != nil {
			t.Fatalf("GrantPermissionToRole failed: %v", err)
		}

		// 5) 権限判定を実行します。
		has, err := repo.HasPermission(ctx, testUserID, testPermExportName)
		if err != nil {
			t.Fatalf("HasPermission failed: %v", err)
		}
		// 6) 付与済みなので true を期待します。
		if !has {
			t.Fatal("expected has permission")
		}
	})
}

func TestRepository_HasPermission_FalseWhenNotGranted(t *testing.T) {
	t.Run("ユーザーが権限を持たない場合はfalseを返す", func(t *testing.T) {
		// 1) テスト用DB接続とRepositoryを準備します。
		db := setupTestDB(t)
		repo := NewRepository(db)
		ctx := context.Background()

		// 2) users / roles / permissions の基本データを投入します。
		seedBase(t, db)
		// 3) ユーザーにはviewerロールのみ付与します。
		if err := repo.AssignRoleToUser(ctx, testUserID, testRoleViewerID); err != nil {
			t.Fatalf("AssignRoleToUser failed: %v", err)
		}
		// 4) viewerロールには閲覧権限のみ付与します。
		if err := repo.GrantPermissionToRole(ctx, testRoleViewerID, testPermViewID); err != nil {
			t.Fatalf("GrantPermissionToRole failed: %v", err)
		}

		// 5) 付与していないエクスポート権限を問い合わせます。
		has, err := repo.HasPermission(ctx, testUserID, testPermExportName)
		if err != nil {
			t.Fatalf("HasPermission failed: %v", err)
		}
		// 6) 未付与なので false を期待します。
		if has {
			t.Fatal("expected no permission")
		}
	})
}

func TestRepository_ListPermissions_DistinctSorted(t *testing.T) {
	t.Run("複数ロールの権限を重複排除してソート順で返す", func(t *testing.T) {
		// 1) テスト用DB接続とRepositoryを準備します。
		db := setupTestDB(t)
		repo := NewRepository(db)
		ctx := context.Background()

		// 2) users / roles / permissions の基本データを投入します。
		seedBase(t, db)
		// 3) 同一ユーザーに2つのロールを付与します。
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

		// 5) 権限一覧を取得します。
		permissions, err := repo.ListPermissions(ctx, testUserID)
		if err != nil {
			t.Fatalf("ListPermissions failed: %v", err)
		}

		// 6) 権限が2件にマージされていることを確認します。
		if len(permissions) != 2 {
			t.Fatalf("expected 2 permissions, got %d (%v)", len(permissions), permissions)
		}
		// 7) ORDER BY p.name でソート済みのため、名前順で比較します。
		if permissions[0] != testPermExportName || permissions[1] != testPermViewName {
			t.Fatalf("unexpected permissions: %v", permissions)
		}
	})
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// CIやローカル差異に対応するため、環境変数で上書き可能にします。
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		// デフォルトは docker-compose.yml の接続情報と一致させています。
		dsn = "app:app@tcp(127.0.0.1:3306)/sample_rbac?parseTime=true"
	}

	var (
		db  *gorm.DB
		err error
	)

	// DB起動直後の接続失敗を吸収するため、短いリトライを行います。
	for range 10 {
		db, err = OpenMySQL(dsn)
		if err == nil {
			break
		}
		// MySQLコンテナ起動直後は接続失敗するため、短時間リトライします。
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		// 実DBが起動していない環境では失敗ではなくスキップにします。
		t.Skipf("mysql not ready: %v", err)
	}

	// テスト独立性のため、実行前後でテーブルをクリーンアップします。
	cleanupTables(t, db)
	t.Cleanup(func() { cleanupTables(t, db) })
	return db
}

func seedBase(t *testing.T, db *gorm.DB) {
	t.Helper()

	// 全テスト共通で使う最小のマスタデータを投入します。
	// users: 権限判定対象のユーザー
	mustExec(t, db, "INSERT INTO users (id, email) VALUES (?, ?)", testUserID, "alice@example.com")
	// roles: admin / viewer の2種類
	mustExec(t, db, "INSERT INTO roles (id, name) VALUES (?, ?), (?, ?)", testRoleAdminID, testRoleAdminName, testRoleViewerID, testRoleViewerName)
	// permissions: export / view の2種類
	mustExec(t, db, "INSERT INTO permissions (id, name) VALUES (?, ?), (?, ?)", testPermExportID, testPermExportName, testPermViewID, testPermViewName)
}

func cleanupTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	// 並列テストで衝突しないよう、テストで使ったIDの行だけ削除します。
	// 依存順: role_permissions -> user_roles -> permissions/roles -> users
	mustExec(t, db, "DELETE FROM role_permissions WHERE role_id IN (?, ?) OR permission_id IN (?, ?)", testRoleAdminID, testRoleViewerID, testPermExportID, testPermViewID)
	mustExec(t, db, "DELETE FROM user_roles WHERE user_id = ? OR role_id IN (?, ?)", testUserID, testRoleAdminID, testRoleViewerID)
	mustExec(t, db, "DELETE FROM permissions WHERE id IN (?, ?)", testPermExportID, testPermViewID)
	mustExec(t, db, "DELETE FROM roles WHERE id IN (?, ?)", testRoleAdminID, testRoleViewerID)
	mustExec(t, db, "DELETE FROM users WHERE id = ?", testUserID)
}

func mustExec(t *testing.T, db *gorm.DB, sql string, args ...any) {
	t.Helper()
	// セットアップ/後片付けのSQL実行を簡潔にするためのヘルパーです。
	if err := db.Exec(sql, args...).Error; err != nil {
		t.Fatalf("exec failed: %s err=%v", sql, err)
	}
}
