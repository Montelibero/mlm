-- +goose Up
-- +goose StatementBegin
CREATE TABLE report_distributes (
  report_id bigint NOT NULL,
  recommender text NOT NULL,
  asset text NOT NULL,
  amount double precision NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
