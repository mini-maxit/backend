-- Create "access_controls" table
CREATE TABLE "maxit"."access_controls" (
  "resource_id" bigint NOT NULL,
  "resource_type" character varying(20) NOT NULL,
  "user_id" bigint NOT NULL,
  "permission" character varying(20) NOT NULL,
  "created_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("resource_id", "resource_type", "user_id"),
  CONSTRAINT "fk_maxit_access_controls_user" FOREIGN KEY ("user_id") REFERENCES "maxit"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_maxit_access_controls_deleted_at" to table: "access_controls"
CREATE INDEX "idx_maxit_access_controls_deleted_at" ON "maxit"."access_controls" ("deleted_at");
