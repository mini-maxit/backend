-- Modify "contest_tasks" table
ALTER TABLE "maxit"."contest_tasks" ADD COLUMN IF NOT EXISTS "position" bigint NOT NULL DEFAULT 0;
-- Initial order = previous (task_id) order; only rows not yet ordered, so a re-run keeps custom orders
UPDATE "maxit"."contest_tasks" ct SET "position" = r.rn FROM (SELECT contest_id, task_id, ROW_NUMBER() OVER (PARTITION BY contest_id ORDER BY task_id) AS rn FROM "maxit"."contest_tasks") r WHERE ct.contest_id = r.contest_id AND ct.task_id = r.task_id AND ct."position" = 0;
