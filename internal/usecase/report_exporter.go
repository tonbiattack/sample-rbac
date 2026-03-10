package usecase

import "context"

// ReportExporter はレポート出力の業務ユースケースです。
// 実処理の前に report.export 権限を必ずチェックします。
type ReportExporter struct {
	authorizer *Authorizer
}

// NewReportExporter はレポート出力ユースケースを生成します。
func NewReportExporter(authorizer *Authorizer) *ReportExporter {
	return &ReportExporter{authorizer: authorizer}
}

// ExportMonthlyReport は月次レポートを出力します。
// このサンプルではファイル名を返し、権限チェックの流れを示します。
func (u *ReportExporter) ExportMonthlyReport(ctx context.Context, userID int64) (string, error) {
	if err := u.authorizer.Require(ctx, userID, "report.export"); err != nil {
		return "", err
	}

	// ここに本来の業務処理（集計、CSV生成、保存など）を実装します。
	return "monthly_report.csv", nil
}
