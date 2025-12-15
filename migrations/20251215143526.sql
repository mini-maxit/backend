-- Modify "test_results" table
ALTER TABLE "maxit"."test_results" ADD COLUMN "peak_memory" bigint NOT NULL;
-- Rename a column from "time_limit" to "time_limit_ms"
ALTER TABLE "maxit"."test_cases" RENAME COLUMN "time_limit" TO "time_limit_ms";
-- Rename a column from "memory_limit" to "memory_limit_kb"
ALTER TABLE "maxit"."test_cases" RENAME COLUMN "memory_limit" TO "memory_limit_kb";
-- Rename a column from "execution_time" to "execution_time_s"
ALTER TABLE "maxit"."test_results" RENAME COLUMN "execution_time" TO "execution_time_s";
-- Rename a column from "peak_memory" to "peak_memory_kb"
ALTER TABLE "maxit"."test_results" RENAME COLUMN "peak_memory" TO "peak_memory_kb";
