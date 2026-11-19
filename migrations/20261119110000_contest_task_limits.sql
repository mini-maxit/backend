-- Modify "contest_tasks" table
ALTER TABLE "maxit"."contest_tasks" ADD COLUMN IF NOT EXISTS "time_limit_ms" bigint NULL, ADD COLUMN IF NOT EXISTS "memory_limit_kb" bigint NULL;
