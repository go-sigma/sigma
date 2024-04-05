CREATE TABLE IF NOT EXISTS "lockers" (
  "id" bigserial PRIMARY KEY,
  "key" varchar(256) NOT NULL,
  "value" varchar(256) NOT NULL,
  "expire" bigint NOT NULL DEFAULT 0,
  "created_at" bigint NOT NULL DEFAULT ((EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000)::bigint),
  "updated_at" bigint NOT NULL DEFAULT ((EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000)::bigint),
  "deleted_at" bigint NOT NULL DEFAULT 0,
  CONSTRAINT "idx_lockers_key" UNIQUE ("key", "deleted_at")
);

