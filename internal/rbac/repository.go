package rbac

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AssignRoleToUser(ctx context.Context, userID, roleID int64) error {
	return r.db.WithContext(ctx).
		Exec("INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).
		Error
}

func (r *Repository) GrantPermissionToRole(ctx context.Context, roleID, permissionID int64) error {
	return r.db.WithContext(ctx).
		Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).
		Error
}

func (r *Repository) HasPermission(ctx context.Context, userID int64, permissionName string) (bool, error) {
	var has bool
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

func (r *Repository) ListPermissions(ctx context.Context, userID int64) ([]string, error) {
	permissions := make([]string, 0)
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
