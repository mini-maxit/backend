-- Modify "submissions" table
ALTER TABLE "maxit"."submissions" ADD COLUMN "contest_id" bigint NULL, ADD CONSTRAINT "fk_maxit_submissions_contest" FOREIGN KEY ("contest_id") REFERENCES "maxit"."contests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
