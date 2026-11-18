-- Modify "contests" table
ALTER TABLE "maxit"."contests" ADD COLUMN "is_public" boolean NOT NULL DEFAULT false;
-- Create "contest_registration_requests" table
CREATE TABLE "maxit"."contest_registration_requests" (
  "contest_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "status" character varying(16) NOT NULL,
  "created_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("contest_id", "user_id"),
  CONSTRAINT "fk_maxit_contest_registration_requests_contest" FOREIGN KEY ("contest_id") REFERENCES "maxit"."contests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maxit_contest_registration_requests_user" FOREIGN KEY ("user_id") REFERENCES "maxit"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
