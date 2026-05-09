package admin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type DispatchHistoryFilter struct {
	StartTime  string
	EndTime    string
	ChannelID  string
	UserID     string
	ProviderID string
	ModelID    string
	KeyID      string
	Status     string
	Sort       string
	Page       int
	PageSize   int
}

type DispatchHistoryRow struct {
	ID                  string
	CreatedAt           string
	ChannelID           string
	ChannelName         string
	UserID              string
	UserDisplayName     string
	ProviderID          string
	ProviderName        string
	ModelID             string
	ModelName           string
	KeyID               string
	KeyAlias            string
	Status              string
	FailureErrorCode    *string
	FailureErrorMessage *string
	FailureReportedAt   *string
}

type DispatchHistoryStats struct {
	TotalCount      int
	FailedCount     int
	SuccessCount    int
	UniqueUserCount int
	UniqueKeyCount  int
}

type DispatchHistoryPoint struct {
	BucketStart string
	TotalCount  int
	FailedCount int
}

type DispatchHistoryPagination struct {
	Page       int
	PageSize   int
	TotalItems int
	TotalPages int
}

type DispatchHistoryUserOption struct {
	ID          string
	ChannelID   string
	DisplayName string
	ChannelName string
}

func (stats DispatchHistoryStats) FailureRatePercent() int {
	if stats.TotalCount == 0 {
		return 0
	}
	return int(float64(stats.FailedCount) / float64(stats.TotalCount) * 100)
}

func (store *Store) ListDispatchHistoryUserOptions(ctx context.Context, channelID string) ([]DispatchHistoryUserOption, error) {
	channelID = strings.TrimSpace(channelID)
	query := `
SELECT users.id, users.channel_id, users.display_name, channels.name
FROM users
JOIN channels ON channels.id = users.channel_id
WHERE (? = '' OR users.channel_id = ?)
ORDER BY channels.name ASC, users.display_name ASC;
`
	rows, err := store.db.QueryContext(ctx, query, channelID, channelID)
	if err != nil {
		return nil, fmt.Errorf("list dispatch history user options: %w", err)
	}
	defer rows.Close()

	options := make([]DispatchHistoryUserOption, 0)
	for rows.Next() {
		var option DispatchHistoryUserOption
		if err := rows.Scan(&option.ID, &option.ChannelID, &option.DisplayName, &option.ChannelName); err != nil {
			return nil, fmt.Errorf("scan dispatch history user option: %w", err)
		}
		options = append(options, option)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dispatch history user options: %w", err)
	}
	return options, nil
}

func (store *Store) ListDispatchHistory(ctx context.Context, filter DispatchHistoryFilter) ([]DispatchHistoryRow, DispatchHistoryPagination, error) {
	filter = normalizeDispatchHistoryFilter(filter)
	where, args := buildDispatchHistoryWhere(filter)

	countQuery := `
SELECT COUNT(*)
FROM dispatch_logs
JOIN channels ON channels.id = dispatch_logs.channel_id
JOIN users ON users.id = dispatch_logs.user_id
JOIN providers ON providers.id = dispatch_logs.provider_id
JOIN models ON models.id = dispatch_logs.model_id
WHERE ` + where + `;
`
	var totalItems int
	if err := store.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems); err != nil {
		return nil, DispatchHistoryPagination{}, fmt.Errorf("count dispatch history: %w", err)
	}

	sortDirection := "DESC"
	if filter.Sort == "asc" {
		sortDirection = "ASC"
	}
	offset := (filter.Page - 1) * filter.PageSize
	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, filter.PageSize, offset)
	query := `
SELECT dispatch_logs.id, dispatch_logs.created_at,
       channels.id, channels.name,
       users.id, users.display_name,
       providers.id, providers.name,
       models.id, models.name,
       dispatch_logs.key_id, dispatch_logs.key_alias_snapshot,
       dispatch_logs.status,
       failure_reports.error_code, failure_reports.error_message, failure_reports.created_at
FROM dispatch_logs
JOIN channels ON channels.id = dispatch_logs.channel_id
JOIN users ON users.id = dispatch_logs.user_id
JOIN providers ON providers.id = dispatch_logs.provider_id
JOIN models ON models.id = dispatch_logs.model_id
LEFT JOIN failure_reports ON failure_reports.dispatch_log_id = dispatch_logs.id
  AND failure_reports.id = (
    SELECT latest_failure.id
    FROM failure_reports AS latest_failure
    WHERE latest_failure.dispatch_log_id = dispatch_logs.id
    ORDER BY latest_failure.created_at DESC, latest_failure.id DESC
    LIMIT 1
  )
WHERE ` + where + `
ORDER BY dispatch_logs.created_at ` + sortDirection + `, dispatch_logs.id ` + sortDirection + `
LIMIT ? OFFSET ?;
`
	rows, err := store.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, DispatchHistoryPagination{}, fmt.Errorf("list dispatch history: %w", err)
	}
	defer rows.Close()

	items := make([]DispatchHistoryRow, 0)
	for rows.Next() {
		var item DispatchHistoryRow
		var failureCode, failureMessage, failureAt sql.NullString
		if err := rows.Scan(
			&item.ID, &item.CreatedAt,
			&item.ChannelID, &item.ChannelName,
			&item.UserID, &item.UserDisplayName,
			&item.ProviderID, &item.ProviderName,
			&item.ModelID, &item.ModelName,
			&item.KeyID, &item.KeyAlias,
			&item.Status,
			&failureCode, &failureMessage, &failureAt,
		); err != nil {
			return nil, DispatchHistoryPagination{}, fmt.Errorf("scan dispatch history: %w", err)
		}
		item.FailureErrorCode = nullStringPtr(failureCode)
		item.FailureErrorMessage = nullStringPtr(failureMessage)
		item.FailureReportedAt = nullStringPtr(failureAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, DispatchHistoryPagination{}, fmt.Errorf("iterate dispatch history: %w", err)
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + filter.PageSize - 1) / filter.PageSize
	}
	pagination := DispatchHistoryPagination{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
	return items, pagination, nil
}

func (store *Store) GetDispatchHistoryStats(ctx context.Context, filter DispatchHistoryFilter) (DispatchHistoryStats, error) {
	filter = normalizeDispatchHistoryFilter(filter)
	where, args := buildDispatchHistoryWhere(filter)
	query := `
SELECT COUNT(*),
       COALESCE(SUM(CASE WHEN dispatch_logs.status = 'FAILED' THEN 1 ELSE 0 END), 0),
       COUNT(DISTINCT dispatch_logs.user_id),
       COUNT(DISTINCT dispatch_logs.key_id)
FROM dispatch_logs
JOIN channels ON channels.id = dispatch_logs.channel_id
JOIN users ON users.id = dispatch_logs.user_id
JOIN providers ON providers.id = dispatch_logs.provider_id
JOIN models ON models.id = dispatch_logs.model_id
WHERE ` + where + `;
`
	var stats DispatchHistoryStats
	if err := store.db.QueryRowContext(ctx, query, args...).Scan(&stats.TotalCount, &stats.FailedCount, &stats.UniqueUserCount, &stats.UniqueKeyCount); err != nil {
		return DispatchHistoryStats{}, fmt.Errorf("get dispatch history stats: %w", err)
	}
	stats.SuccessCount = stats.TotalCount - stats.FailedCount
	return stats, nil
}

func (store *Store) GetDispatchHistorySeries(ctx context.Context, filter DispatchHistoryFilter, bucket string) ([]DispatchHistoryPoint, error) {
	filter = normalizeDispatchHistoryFilter(filter)
	where, args := buildDispatchHistoryWhere(filter)
	bucketExpression := "strftime('%Y-%m-%dT00:00:00Z', dispatch_logs.created_at)"
	switch bucket {
	case "hour":
		bucketExpression = "strftime('%Y-%m-%dT%H:00:00Z', dispatch_logs.created_at)"
	case "month":
		bucketExpression = "strftime('%Y-%m-01T00:00:00Z', dispatch_logs.created_at)"
	}
	query := `
SELECT ` + bucketExpression + ` AS bucket_start,
       COUNT(*) AS total_count,
       COALESCE(SUM(CASE WHEN dispatch_logs.status = 'FAILED' THEN 1 ELSE 0 END), 0) AS failed_count
FROM dispatch_logs
JOIN channels ON channels.id = dispatch_logs.channel_id
JOIN users ON users.id = dispatch_logs.user_id
JOIN providers ON providers.id = dispatch_logs.provider_id
JOIN models ON models.id = dispatch_logs.model_id
WHERE ` + where + `
GROUP BY bucket_start
ORDER BY bucket_start ASC;
`
	rows, err := store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get dispatch history series: %w", err)
	}
	defer rows.Close()

	points := make([]DispatchHistoryPoint, 0)
	for rows.Next() {
		var point DispatchHistoryPoint
		if err := rows.Scan(&point.BucketStart, &point.TotalCount, &point.FailedCount); err != nil {
			return nil, fmt.Errorf("scan dispatch history series: %w", err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dispatch history series: %w", err)
	}
	return points, nil
}

func normalizeDispatchHistoryFilter(filter DispatchHistoryFilter) DispatchHistoryFilter {
	filter.StartTime = strings.TrimSpace(filter.StartTime)
	filter.EndTime = strings.TrimSpace(filter.EndTime)
	filter.ChannelID = strings.TrimSpace(filter.ChannelID)
	filter.UserID = strings.TrimSpace(filter.UserID)
	filter.ProviderID = strings.TrimSpace(filter.ProviderID)
	filter.ModelID = strings.TrimSpace(filter.ModelID)
	filter.KeyID = strings.TrimSpace(filter.KeyID)
	filter.Status = strings.TrimSpace(filter.Status)
	filter.Sort = strings.TrimSpace(filter.Sort)
	if filter.Sort != "asc" {
		filter.Sort = "desc"
	}
	if filter.Status != "DISPATCHED" && filter.Status != "FAILED" {
		filter.Status = ""
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 50
	}
	if filter.PageSize > 200 {
		filter.PageSize = 200
	}
	return filter
}

func buildDispatchHistoryWhere(filter DispatchHistoryFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := make([]any, 0)
	if filter.StartTime != "" {
		clauses = append(clauses, "dispatch_logs.created_at >= ?")
		args = append(args, filter.StartTime)
	}
	if filter.EndTime != "" {
		clauses = append(clauses, "dispatch_logs.created_at <= ?")
		args = append(args, filter.EndTime)
	}
	if filter.ChannelID != "" {
		clauses = append(clauses, "dispatch_logs.channel_id = ?")
		args = append(args, filter.ChannelID)
	}
	if filter.UserID != "" {
		clauses = append(clauses, "dispatch_logs.user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.ProviderID != "" {
		clauses = append(clauses, "dispatch_logs.provider_id = ?")
		args = append(args, filter.ProviderID)
	}
	if filter.ModelID != "" {
		clauses = append(clauses, "dispatch_logs.model_id = ?")
		args = append(args, filter.ModelID)
	}
	if filter.KeyID != "" {
		clauses = append(clauses, "dispatch_logs.key_id = ?")
		args = append(args, filter.KeyID)
	}
	if filter.Status != "" {
		clauses = append(clauses, "dispatch_logs.status = ?")
		args = append(args, filter.Status)
	}
	return strings.Join(clauses, " AND "), args
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
