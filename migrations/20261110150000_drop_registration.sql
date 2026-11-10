-- Drop "contest_registration_requests" table
DROP TABLE "maxit"."contest_registration_requests";
-- Modify "contests" table
ALTER TABLE "maxit"."contests" DROP COLUMN "is_registration_open";
-- Drop enum type "registration_request_status"
DROP TYPE "maxit"."registration_request_status";
