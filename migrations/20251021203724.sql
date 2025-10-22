-- Add new schema named "maxit"
CREATE SCHEMA "maxit";
-- Create "task_users" table
CREATE TABLE "maxit"."task_users" (
  "task_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  PRIMARY KEY ("task_id", "user_id")
);
-- Create "users" table
CREATE TABLE "maxit"."users" (
  "id" bigserial NOT NULL,
  "name" text NOT NULL,
  "surname" text NOT NULL,
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "role" text NOT NULL DEFAULT 'student',
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_maxit_users_email" UNIQUE ("email")
);
-- Create "user_groups" table
CREATE TABLE "maxit"."user_groups" (
  "user_id" bigint NOT NULL,
  "group_id" bigint NOT NULL,
  PRIMARY KEY ("user_id", "group_id")
);
-- Create "contests" table
CREATE TABLE "maxit"."contests" (
  "id" bigserial NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "created_by" bigint NOT NULL,
  "start_at" timestamptz NULL,
  "end_at" timestamptz NULL,
  "is_registration_open" boolean NOT NULL DEFAULT true,
  "is_submission_open" boolean NOT NULL DEFAULT false,
  "is_visible" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_maxit_contests_name" UNIQUE ("name"),
  CONSTRAINT "fk_maxit_contests_creator" FOREIGN KEY ("created_by") REFERENCES "maxit"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_maxit_contests_deleted_at" to table: "contests"
CREATE INDEX "idx_maxit_contests_deleted_at" ON "maxit"."contests" ("deleted_at");
-- Create "groups" table
CREATE TABLE "maxit"."groups" (
  "id" bigserial NOT NULL,
  "name" text NOT NULL,
  "created_by" bigint NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maxit_groups_author" FOREIGN KEY ("created_by") REFERENCES "maxit"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "contest_participant_groups" table
CREATE TABLE "maxit"."contest_participant_groups" (
  "contest_id" bigint NOT NULL,
  "group_id" bigint NOT NULL,
  PRIMARY KEY ("contest_id", "group_id"),
  CONSTRAINT "fk_maxit_contest_participant_groups_contest" FOREIGN KEY ("contest_id") REFERENCES "maxit"."contests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_contest_participant_groups_group" FOREIGN KEY ("group_id") REFERENCES "maxit"."groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "contest_participants" table
CREATE TABLE "maxit"."contest_participants" (
  "contest_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  PRIMARY KEY ("contest_id", "user_id"),
  CONSTRAINT "fk_maxit_contest_participants_contest" FOREIGN KEY ("contest_id") REFERENCES "maxit"."contests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_contest_participants_user" FOREIGN KEY ("user_id") REFERENCES "maxit"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "contest_pending_registrations" table
CREATE TABLE "maxit"."contest_pending_registrations" (
  "id" bigserial NOT NULL,
  "contest_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maxit_contest_pending_registrations_contest" FOREIGN KEY ("contest_id") REFERENCES "maxit"."contests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_contest_pending_registrations_user" FOREIGN KEY ("user_id") REFERENCES "maxit"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "files" table
CREATE TABLE "maxit"."files" (
  "id" bigserial NOT NULL,
  "filename" text NOT NULL,
  "path" text NOT NULL,
  "bucket" text NOT NULL,
  "server_type" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "tasks" table
CREATE TABLE "maxit"."tasks" (
  "id" bigserial NOT NULL,
  "title" character varying(255) NULL,
  "description_file_id" bigint NULL,
  "created_by" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maxit_tasks_author" FOREIGN KEY ("created_by") REFERENCES "maxit"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_tasks_description_file" FOREIGN KEY ("description_file_id") REFERENCES "maxit"."files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_maxit_tasks_deleted_at" to table: "tasks"
CREATE INDEX "idx_maxit_tasks_deleted_at" ON "maxit"."tasks" ("deleted_at");
-- Create "contest_tasks" table
CREATE TABLE "maxit"."contest_tasks" (
  "contest_id" bigint NOT NULL,
  "task_id" bigint NOT NULL,
  PRIMARY KEY ("contest_id", "task_id"),
  CONSTRAINT "fk_maxit_contest_tasks_contest" FOREIGN KEY ("contest_id") REFERENCES "maxit"."contests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_contest_tasks_task" FOREIGN KEY ("task_id") REFERENCES "maxit"."tasks" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "language_configs" table
CREATE TABLE "maxit"."language_configs" (
  "id" bigserial NOT NULL,
  "type" text NOT NULL,
  "version" text NOT NULL,
  "file_extension" text NOT NULL,
  "is_disabled" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create "submissions" table
CREATE TABLE "maxit"."submissions" (
  "id" bigserial NOT NULL,
  "task_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "order" bigint NOT NULL,
  "language_id" bigint NOT NULL,
  "file_id" bigint NOT NULL,
  "status" character varying(50) NOT NULL,
  "status_message" character varying(255) NULL DEFAULT NULL::character varying,
  "submitted_at" timestamp NULL,
  "checked_at" timestamp NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maxit_submissions_file" FOREIGN KEY ("file_id") REFERENCES "maxit"."files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_submissions_language" FOREIGN KEY ("language_id") REFERENCES "maxit"."language_configs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_submissions_task" FOREIGN KEY ("task_id") REFERENCES "maxit"."tasks" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_submissions_user" FOREIGN KEY ("user_id") REFERENCES "maxit"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "queue_messages" table
CREATE TABLE "maxit"."queue_messages" (
  "id" text NOT NULL,
  "submission_id" bigint NOT NULL,
  "queued_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maxit_queue_messages_submission" FOREIGN KEY ("submission_id") REFERENCES "maxit"."submissions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "submission_results" table
CREATE TABLE "maxit"."submission_results" (
  "id" bigserial NOT NULL,
  "submission_id" bigint NOT NULL,
  "code" bigint NOT NULL,
  "message" character varying(255) NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maxit_submissions_result" FOREIGN KEY ("submission_id") REFERENCES "maxit"."submissions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "task_groups" table
CREATE TABLE "maxit"."task_groups" (
  "task_id" bigint NOT NULL,
  "group_id" bigint NOT NULL,
  PRIMARY KEY ("task_id", "group_id"),
  CONSTRAINT "fk_maxit_task_groups_group" FOREIGN KEY ("group_id") REFERENCES "maxit"."groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_task_groups_task" FOREIGN KEY ("task_id") REFERENCES "maxit"."tasks" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "test_cases" table
CREATE TABLE "maxit"."test_cases" (
  "id" bigserial NOT NULL,
  "task_id" bigint NOT NULL,
  "input_file_id" bigint NOT NULL,
  "output_file_id" bigint NOT NULL,
  "order" bigint NOT NULL,
  "time_limit" bigint NOT NULL,
  "memory_limit" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maxit_test_cases_input_file" FOREIGN KEY ("input_file_id") REFERENCES "maxit"."files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_test_cases_output_file" FOREIGN KEY ("output_file_id") REFERENCES "maxit"."files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_test_cases_task" FOREIGN KEY ("task_id") REFERENCES "maxit"."tasks" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "test_results" table
CREATE TABLE "maxit"."test_results" (
  "id" bigserial NOT NULL,
  "submission_result_id" bigint NOT NULL,
  "test_case_id" bigint NOT NULL,
  "passed" boolean NOT NULL,
  "execution_time" numeric NOT NULL,
  "status_code" bigint NOT NULL,
  "error_message" character varying NULL,
  "stdout_file_id" bigint NOT NULL,
  "stderr_file_id" bigint NOT NULL,
  "diff_file_id" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maxit_submission_results_test_results" FOREIGN KEY ("submission_result_id") REFERENCES "maxit"."submission_results" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_test_results_diff_file" FOREIGN KEY ("diff_file_id") REFERENCES "maxit"."files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_test_results_stderr_file" FOREIGN KEY ("stderr_file_id") REFERENCES "maxit"."files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_test_results_stdout_file" FOREIGN KEY ("stdout_file_id") REFERENCES "maxit"."files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_test_results_test_case" FOREIGN KEY ("test_case_id") REFERENCES "maxit"."test_cases" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
