package usecase

import (
	"context"
	"errors"
)

// ErrForbidden は権限不足時に返すエラーです。
var ErrForbidden = errors.New("forbidden")

// PermissionChecker は権限判定に必要な最小インターフェースです。
// rbac.Repository がこのインターフェースを満たします。
type PermissionChecker interface {
	HasPermission(ctx context.Context, userID int64, permissionName string) (bool, error)
}

// Authorizer は「このユーザーがこの権限を持つか」の判定を集約します。
type Authorizer struct {
	checker PermissionChecker
}

// NewAuthorizer は権限判定コンポーネントを生成します。
func NewAuthorizer(checker PermissionChecker) *Authorizer {
	return &Authorizer{checker: checker}
}

// Require は指定権限を必須条件としてチェックします。
// 権限がない場合は ErrForbidden を返します。
func (a *Authorizer) Require(ctx context.Context, userID int64, permissionName string) error {
	has, err := a.checker.HasPermission(ctx, userID, permissionName)
	if err != nil {
		return err
	}
	if !has {
		return ErrForbidden
	}
	return nil
}
