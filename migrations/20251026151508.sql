-- Modify "contest_participant_groups" table
ALTER TABLE "maxit"."contest_participant_groups" DROP CONSTRAINT "fk_maxit_contest_participant_groups_contest", DROP CONSTRAINT "fk_maxit_contest_participant_groups_group";
-- Modify "contest_participants" table
ALTER TABLE "maxit"."contest_participants" DROP CONSTRAINT "fk_maxit_contest_participants_contest", DROP CONSTRAINT "fk_maxit_contest_participants_user";
-- Modify "contest_tasks" table
ALTER TABLE "maxit"."contest_tasks" DROP CONSTRAINT "fk_maxit_contest_tasks_contest", DROP CONSTRAINT "fk_maxit_contest_tasks_task";
-- Modify "task_groups" table
ALTER TABLE "maxit"."task_groups" DROP CONSTRAINT "fk_maxit_task_groups_group", DROP CONSTRAINT "fk_maxit_task_groups_task";
-- Modify "users" table
ALTER TABLE "maxit"."users" ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "updated_at" timestamptz NULL, ADD COLUMN "deleted_at" timestamptz NULL;
-- Create index "idx_maxit_users_deleted_at" to table: "users"
CREATE INDEX "idx_maxit_users_deleted_at" ON "maxit"."users" ("deleted_at");
