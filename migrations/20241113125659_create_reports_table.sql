-- +goose Up
-- +goose StatementBegin
CREATE TABLE reports (
  id bigserial NOT NULL,
  created_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  xdr text NOT NULL
);

CREATE INDEX idx_reports_created_at_desc_deleted_at_is_null
ON reports (created_at DESC)
WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
