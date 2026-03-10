package rbac

import (
	"context"

	"gorm.io/gorm"
)

// Repository は MySQL に対する RBAC クエリを担当します。
// テーブル関係は次の通りです。
// users -> user_roles -> roles -> role_permissions -> permissions
type Repository struct {
	db *gorm.DB
}

// NewRepository は既存の GORM DB を使って Repository を生成します。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// AssignRoleToUser はユーザーにロールを付与します。
// INSERT IGNORE により、既に同じ組み合わせがある場合でも冪等に動作します。
func (r *Repository) AssignRoleToUser(ctx context.Context, userID, roleID int64) error {
	return r.db.WithContext(ctx).
		Exec("INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).
		Error
}

// GrantPermissionToRole はロールに権限を付与します。
// INSERT IGNORE により、既に同じ組み合わせがある場合でも冪等に動作します。
func (r *Repository) GrantPermissionToRole(ctx context.Context, roleID, permissionID int64) error {
	return r.db.WithContext(ctx).
		Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).
		Error
}

// HasPermission はユーザーが指定権限名を持っているかを判定します。
// 記事で示した EXISTS クエリをそのまま実装しています。
func (r *Repository) HasPermission(ctx context.Context, userID int64, permissionName string) (bool, error) {
	var has bool
	// 同じ権限が複数ロールから付与される可能性があるため、
	// 件数ではなく true/false を返す EXISTS を使います。
	err := r.db.WithContext(ctx).Raw(`
SELECT EXISTS (
  SELECT 1
  FROM user_roles ur
  JOIN role_permissions rp ON rp.role_id = ur.role_id
  JOIN permissions p ON p.id = rp.permission_id
  WHERE ur.user_id = ?
    AND p.name = ?
) AS has_permission
`, userID, permissionName).Scan(&has).Error

	return has, err
}

// ListPermissions はユーザーが持つ権限名一覧を返します。
// ソート済みで返すことで、呼び出し側とテストで扱いやすくします。
func (r *Repository) ListPermissions(ctx context.Context, userID int64) ([]string, error) {
	permissions := make([]string, 0)
	// 複数ロールに同じ権限が含まれる場合に備え、DISTINCT で重複を除去します。
	err := r.db.WithContext(ctx).Raw(`
SELECT DISTINCT p.name
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE ur.user_id = ?
ORDER BY p.name
`, userID).Scan(&permissions).Error

	return permissions, err
}
