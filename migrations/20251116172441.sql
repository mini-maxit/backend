-- Backfill NULL start_at values before enforcing NOT NULL constraint
UPDATE "maxit"."contests" SET "start_at" = "created_at" WHERE "start_at" IS NULL;
-- Modify "contests" table
ALTER TABLE "maxit"."contests" ALTER COLUMN "start_at" SET NOT NULL;
