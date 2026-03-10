package rbac

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// OpenMySQL は MySQL のDSNを使って GORM 接続を作成します。
// DSN例: app:app@tcp(127.0.0.1:3306)/sample_rbac?parseTime=true
func OpenMySQL(dsn string) (*gorm.DB, error) {
	// サンプルのため設定は最小限にしています。
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
