// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package db

import (
	"context"
)

const createReport = `-- name: CreateReport :one
INSERT INTO reports (created_at, xdr)
  VALUES (now(), $1) RETURNING id
`

func (q *Queries) CreateReport(ctx context.Context, xdr string) (int64, error) {
	row := q.db.QueryRow(ctx, createReport, xdr)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const createReportDistribute = `-- name: CreateReportDistribute :exec
INSERT INTO report_distributes (report_id, recommender, asset, amount)
  VALUES ($1, $2, $3, $4)
`

type CreateReportDistributeParams struct {
	ReportID    int64
	Recommender string
	Asset       string
	Amount      float64
}

func (q *Queries) CreateReportDistribute(ctx context.Context, arg CreateReportDistributeParams) error {
	_, err := q.db.Exec(ctx, createReportDistribute,
		arg.ReportID,
		arg.Recommender,
		arg.Asset,
		arg.Amount,
	)
	return err
}

const createReportRecommend = `-- name: CreateReportRecommend :exec
INSERT INTO report_recommends (report_id, recommender, recommended, recommended_mtlap)
  VALUES ($1, $2, $3, $4)
`

type CreateReportRecommendParams struct {
	ReportID         int64
	Recommender      string
	Recommended      string
	RecommendedMtlap int64
}

func (q *Queries) CreateReportRecommend(ctx context.Context, arg CreateReportRecommendParams) error {
	_, err := q.db.Exec(ctx, createReportRecommend,
		arg.ReportID,
		arg.Recommender,
		arg.Recommended,
		arg.RecommendedMtlap,
	)
	return err
}

const createState = `-- name: CreateState :exec
INSERT INTO states (user_id, state, data, meta, created_at)
  VALUES ($1, $2, $3, $4, now())
`

type CreateStateParams struct {
	UserID int64
	State  string
	Data   map[string]interface{}
	Meta   map[string]interface{}
}

func (q *Queries) CreateState(ctx context.Context, arg CreateStateParams) error {
	_, err := q.db.Exec(ctx, createState,
		arg.UserID,
		arg.State,
		arg.Data,
		arg.Meta,
	)
	return err
}

const deleteReport = `-- name: DeleteReport :exec
UPDATE reports
SET deleted_at = now()
WHERE id = $1
`

func (q *Queries) DeleteReport(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteReport, id)
	return err
}

const getReport = `-- name: GetReport :one
SELECT id, created_at, deleted_at, xdr FROM reports
WHERE deleted_at IS NULL AND
  id = $1
`

func (q *Queries) GetReport(ctx context.Context, id int64) (Report, error) {
	row := q.db.QueryRow(ctx, getReport, id)
	var i Report
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.DeletedAt,
		&i.Xdr,
	)
	return i, err
}

const getReportDistributes = `-- name: GetReportDistributes :many
SELECT report_id, recommender, asset, amount FROM report_distributes
WHERE report_id = $1
`

func (q *Queries) GetReportDistributes(ctx context.Context, reportID int64) ([]ReportDistribute, error) {
	rows, err := q.db.Query(ctx, getReportDistributes, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReportDistribute
	for rows.Next() {
		var i ReportDistribute
		if err := rows.Scan(
			&i.ReportID,
			&i.Recommender,
			&i.Asset,
			&i.Amount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getReportRecommends = `-- name: GetReportRecommends :many
SELECT report_id, recommender, recommended, recommended_mtlap FROM report_recommends
WHERE report_id = $1
`

func (q *Queries) GetReportRecommends(ctx context.Context, reportID int64) ([]ReportRecommend, error) {
	rows, err := q.db.Query(ctx, getReportRecommends, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReportRecommend
	for rows.Next() {
		var i ReportRecommend
		if err := rows.Scan(
			&i.ReportID,
			&i.Recommender,
			&i.Recommended,
			&i.RecommendedMtlap,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getReports = `-- name: GetReports :many
SELECT id, created_at, deleted_at, xdr FROM reports
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT nullif($1::int, 0)
`

func (q *Queries) GetReports(ctx context.Context, queryLimit int32) ([]Report, error) {
	rows, err := q.db.Query(ctx, getReports, queryLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Report
	for rows.Next() {
		var i Report
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.DeletedAt,
			&i.Xdr,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getState = `-- name: GetState :one
SELECT user_id, state, data, meta, created_at FROM states
WHERE ($1::bigint = 0::bigint OR user_id = $1)
ORDER BY created_at DESC
LIMIT 1
`

func (q *Queries) GetState(ctx context.Context, userID int64) (State, error) {
	row := q.db.QueryRow(ctx, getState, userID)
	var i State
	err := row.Scan(
		&i.UserID,
		&i.State,
		&i.Data,
		&i.Meta,
		&i.CreatedAt,
	)
	return i, err
}

const lockReport = `-- name: LockReport :exec
SELECT pg_advisory_lock(1)
`

func (q *Queries) LockReport(ctx context.Context) error {
	_, err := q.db.Exec(ctx, lockReport)
	return err
}

const unlockReport = `-- name: UnlockReport :exec
SELECT pg_advisory_unlock(1)
`

func (q *Queries) UnlockReport(ctx context.Context) error {
	_, err := q.db.Exec(ctx, unlockReport)
	return err
}
