-- +migrate Up notransaction
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS title_idx ON films (title) INCLUDE (director, rating);

-- +migrate Down notransaction
DROP INDEX CONCURRENTLY IF EXISTS title_idx;