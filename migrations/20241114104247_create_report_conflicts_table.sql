-- +goose Up
-- +goose StatementBegin
CREATE TABLE report_conflicts (
  report_id bigint NOT NULL,
  recommender text NOT NULL,
  recommended text NOT NULL
);

CREATE INDEX idx_report_conflicts_report_id
ON report_conflicts (report_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
