-- Backfill an owner row (the creator) for every task, contest and group that has no live owner,
-- e.g. rows created before access control or whose grant failed silently. Must run before the index.
INSERT INTO "maxit"."access_controls" ("resource_id", "resource_type", "user_id", "permission")
SELECT r."id", 'task', r."created_by", 'owner' FROM "maxit"."tasks" r
WHERE r."created_by" IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM "maxit"."access_controls" ac WHERE ac."resource_type" = 'task' AND ac."resource_id" = r."id" AND ac."permission" = 'owner' AND ac."deleted_at" IS NULL)
ON CONFLICT ("resource_id", "resource_type", "user_id") DO UPDATE SET "permission" = 'owner', "deleted_at" = NULL, "updated_at" = CURRENT_TIMESTAMP;
INSERT INTO "maxit"."access_controls" ("resource_id", "resource_type", "user_id", "permission")
SELECT r."id", 'contest', r."created_by", 'owner' FROM "maxit"."contests" r
WHERE r."created_by" IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM "maxit"."access_controls" ac WHERE ac."resource_type" = 'contest' AND ac."resource_id" = r."id" AND ac."permission" = 'owner' AND ac."deleted_at" IS NULL)
ON CONFLICT ("resource_id", "resource_type", "user_id") DO UPDATE SET "permission" = 'owner', "deleted_at" = NULL, "updated_at" = CURRENT_TIMESTAMP;
INSERT INTO "maxit"."access_controls" ("resource_id", "resource_type", "user_id", "permission")
SELECT r."id", 'group', r."created_by", 'owner' FROM "maxit"."groups" r
WHERE r."created_by" IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM "maxit"."access_controls" ac WHERE ac."resource_type" = 'group' AND ac."resource_id" = r."id" AND ac."permission" = 'owner' AND ac."deleted_at" IS NULL)
ON CONFLICT ("resource_id", "resource_type", "user_id") DO UPDATE SET "permission" = 'owner', "deleted_at" = NULL, "updated_at" = CURRENT_TIMESTAMP;
-- Create index "idx_one_owner" to table: "access_controls"
CREATE UNIQUE INDEX "idx_one_owner" ON "maxit"."access_controls" ("resource_type", "resource_id") WHERE (((permission)::text = 'owner'::text) AND (deleted_at IS NULL));
