-- +goose Up
-- +goose StatementBegin
CREATE TABLE report_recommends (
  report_id bigint NOT NULL,
  recommender text NOT NULL,
  recommended text NOT NULL,
  recommended_mtlap bigint NOT NULL
);

CREATE INDEX idx_report_recommends_report_id_recommender_recommended
ON report_recommends (report_id, recommender, recommended);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
