-- +migrate Up
-- pgsafemigrate:nolint:high-availability-alter-column-not-null-exclusive-lock,high-availability-avoid-required-column

ALTER TABLE "recipes" ADD COLUMN "public" boolean NOT NULL, ADD COLUMN "private" boolean;

-- alter table movies
--     add constraint public_not_null
--         check ("public" is not null) not valid;
ALTER TABLE movies ALTER COLUMN "public" SET NOT NULL;

CREATE INDEX ON films (created_at);

CREATE UNIQUE INDEX title_idx ON films (title) INCLUDE (director, rating);
CREATE INDEX CONCURRENTLY "email_idx" ON "companies" ("email");

COMMENT on column recipes.public is 'Whether the recipe is public or not';

BEGIN;
UPDATE films SET updated_at = CURRENT_TIMESTAMP;
COMMIT;

ALTER TABLE "movies" RENAME TO "movies_old";

-- +migrate Down

DROP INDEX IF EXISTS title_idx;