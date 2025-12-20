-- Modify "test_results" table
ALTER TABLE "maxit"."test_results" ALTER COLUMN "passed" DROP NOT NULL, ALTER COLUMN "execution_time_s" DROP NOT NULL, ALTER COLUMN "peak_memory_kb" DROP NOT NULL, ADD COLUMN "exit_code" integer NULL;
