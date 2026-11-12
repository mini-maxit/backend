-- Modify "contest_tasks" table
ALTER TABLE "maxit"."contest_tasks" ADD COLUMN "time_limit_ms" bigint NULL, ADD COLUMN "memory_limit_kb" bigint NULL;
