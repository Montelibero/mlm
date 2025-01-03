-- +goose Up
-- +goose StatementBegin
ALTER TABLE reports
  ADD COLUMN hash text,
  ADD COLUMN updated_at timestamp with time zone;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
