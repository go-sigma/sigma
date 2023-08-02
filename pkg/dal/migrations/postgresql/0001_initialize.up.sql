CREATE TYPE provider AS ENUM (
  'local',
  'github'
);

CREATE TABLE IF NOT EXISTS "users" (
  "id" bigserial PRIMARY KEY,
  "provider" provider NOT NULL DEFAULT 'local',
  "provider_account_id" varchar(256),
  "username" varchar(64) NOT NULL UNIQUE,
  "password" varchar(256) NOT NULL,
  "email" varchar(256),
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  CONSTRAINT "users_unique_with_username" UNIQUE ("username", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "user_recover_codes" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "code" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" integer NOT NULL DEFAULT 0,
  FOREIGN KEY ("user_id") REFERENCES "users" ("id"),
  CONSTRAINT "user_recover_codes_unique_with_user_id" UNIQUE ("user_id", "deleted_at")
);

CREATE TYPE visibility AS ENUM (
  'public',
  'private'
);

CREATE TABLE IF NOT EXISTS "namespaces" (
  "id" bigserial PRIMARY KEY,
  "name" varchar(64) NOT NULL,
  "description" varchar(256),
  "visibility" visibility NOT NULL DEFAULT 'private',
  "size_limit" bigint NOT NULL DEFAULT 0,
  "size" bigint NOT NULL DEFAULT 0,
  "repository_limit" bigint NOT NULL DEFAULT 0,
  "repository_count" bigint NOT NULL DEFAULT 0,
  "tag_limit" bigint NOT NULL DEFAULT 0,
  "tag_count" bigint NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  CONSTRAINT "namespaces_unique_with_name" UNIQUE ("name", "deleted_at")
);

CREATE TYPE audit_action AS ENUM (
  'create',
  'update',
  'delete',
  'pull',
  'push'
);

CREATE TYPE audit_resource_type AS ENUM (
  'namespace',
  'repository',
  'tag'
);

CREATE TABLE IF NOT EXISTS "audits" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "namespace_id" bigint,
  "action" audit_action NOT NULL,
  "resource_type" audit_resource_type NOT NULL,
  "resource" varchar(256) NOT NULL,
  "before_raw" bytea,
  "req_raw" bytea,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("user_id") REFERENCES "users" ("id"),
  FOREIGN KEY ("namespace_id") REFERENCES "namespaces" ("id")
);

CREATE TABLE IF NOT EXISTS "repositories" (
  "id" bigserial PRIMARY KEY,
  "name" varchar(64) NOT NULL,
  "description" varchar(255),
  "overview" bytea,
  "visibility" visibility NOT NULL DEFAULT 'private',
  "size_limit" bigint NOT NULL DEFAULT 0,
  "size" bigint NOT NULL DEFAULT 0,
  "tag_limit" bigint NOT NULL DEFAULT 0,
  "tag_count" bigint NOT NULL DEFAULT 0,
  "namespace_id" bigserial NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("namespace_id") REFERENCES "namespaces" ("id"),
  CONSTRAINT "repositories_unique_with_namespace" UNIQUE ("namespace_id", "name", "deleted_at")
);

CREATE TYPE artifact_type AS ENUM (
  'image',
  'imageIndex',
  'chart',
  'cnab',
  'wasm',
  'provenance',
  'unknown'
);

CREATE TABLE IF NOT EXISTS "artifacts" (
  "id" bigserial PRIMARY KEY,
  "repository_id" bigint NOT NULL,
  "digest" varchar(256) NOT NULL,
  "size" bigint NOT NULL DEFAULT 0,
  "blobs_size" bigint NOT NULL DEFAULT 0,
  "content_type" varchar(256) NOT NULL,
  "raw" bytea NOT NULL,
  "config_raw" bytea,
  "config_media_type" varchar(256),
  "type" artifact_type NOT NULL DEFAULT 'unknown',
  "pushed_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "last_pull" timestamp,
  "pull_times" bigint NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("repository_id") REFERENCES "repositories" ("id"),
  CONSTRAINT "artifacts_unique_with_repo" UNIQUE ("repository_id", "digest", "deleted_at")
);

CREATE TYPE daemon_status AS ENUM (
  'Pending',
  'Doing',
  'Success',
  'Failed'
);

CREATE TABLE IF NOT EXISTS "artifact_sboms" (
  "id" bigserial PRIMARY KEY,
  "artifact_id" bigint NOT NULL,
  "raw" bytea,
  "result" bytea,
  "status" daemon_status NOT NULL,
  "stdout" bytea,
  "stderr" bytea,
  "message" varchar(256),
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("artifact_id") REFERENCES "artifacts" ("id"),
  CONSTRAINT "artifact_sbom_unique_with_artifact" UNIQUE ("artifact_id", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "artifact_vulnerabilities" (
  "id" bigserial PRIMARY KEY,
  "artifact_id" bigint NOT NULL,
  "metadata" bytea,
  "raw" bytea,
  "result" bytea,
  "status" daemon_status NOT NULL,
  "stdout" bytea,
  "stderr" bytea,
  "message" varchar(256),
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("artifact_id") REFERENCES "artifacts" ("id"),
  CONSTRAINT "artifact_vulnerability_unique_with_artifact" UNIQUE ("artifact_id", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "tags" (
  "id" bigserial PRIMARY KEY,
  "repository_id" bigint NOT NULL,
  "artifact_id" bigint NOT NULL,
  "name" varchar(64) NOT NULL,
  "pushed_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "last_pull" timestamp,
  "pull_times" bigint NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("repository_id") REFERENCES "repositories" ("id"),
  FOREIGN KEY ("artifact_id") REFERENCES "artifacts" ("id"),
  CONSTRAINT "tags_unique_with_repo" UNIQUE ("repository_id", "name", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "blobs" (
  "id" bigserial PRIMARY KEY,
  "digest" varchar(256) NOT NULL UNIQUE,
  "size" bigint NOT NULL DEFAULT 0,
  "content_type" varchar(256) NOT NULL,
  "pushed_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "last_pull" timestamp,
  "pull_times" bigint NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  CONSTRAINT "blobs_unique_with_digest" UNIQUE ("digest", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "blob_uploads" (
  "id" bigserial PRIMARY KEY,
  "part_number" int NOT NULL,
  "upload_id" varchar(256) NOT NULL,
  "etag" varchar(256) NOT NULL,
  "repository" varchar(256) NOT NULL,
  "file_id" varchar(256) NOT NULL,
  "size" bigint NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  CONSTRAINT "blob_uploads_unique_with_upload_id_etag" UNIQUE ("upload_id", "etag", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "artifact_artifacts" (
  "artifact_id" bigint NOT NULL,
  "artifact_index_id" bigint NOT NULL,
  PRIMARY KEY ("artifact_id", "artifact_index_id"),
  CONSTRAINT "fk_artifact_artifacts_artifact" FOREIGN KEY ("artifact_id") REFERENCES "artifacts" ("id"),
  CONSTRAINT "fk_artifact_artifacts_artifact_index" FOREIGN KEY ("artifact_index_id") REFERENCES "artifacts" ("id")
);

CREATE TABLE IF NOT EXISTS "artifact_blobs" (
  "artifact_id" bigint NOT NULL,
  "blob_id" bigint NOT NULL,
  PRIMARY KEY ("artifact_id", "blob_id"),
  CONSTRAINT "fk_artifact_blobs_artifact" FOREIGN KEY ("artifact_id") REFERENCES "artifacts" ("id"),
  CONSTRAINT "fk_artifact_blobs_blob" FOREIGN KEY ("blob_id") REFERENCES "blobs" ("id")
);

CREATE TYPE daemon_type AS ENUM (
  'Gc',
  'Vulnerability',
  'Sbom'
);

CREATE TABLE IF NOT EXISTS "daemon_logs" (
  "id" bigserial PRIMARY KEY,
  "namespace_id" bigint,
  "type" daemon_type NOT NULL,
  "action" audit_action NOT NULL,
  "resource" varchar(256) NOT NULL,
  "status" daemon_status NOT NULL,
  "message" bytea,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0
);

CREATE TABLE "casbin_rules" (
  "id" bigserial PRIMARY KEY,
  "ptype" varchar(100),
  "v0" varchar(100),
  "v1" varchar(100),
  "v2" varchar(100),
  "v3" varchar(100),
  "v4" varchar(100),
  "v5" varchar(100),
  CONSTRAINT "idx_casbin_rules" UNIQUE ("ptype", "v0", "v1", "v2", "v3", "v4", "v5")
);

-- ptype type
-- v0 sub
-- v1 dom
-- v2 url
-- v3 attr
-- v4 method
-- v5 allow or deny
INSERT INTO "casbin_rules" ("ptype", "v0", "v1", "v2", "v3", "v4", "v5")
  VALUES ('p', 'admin', '*', '*', '*', '*', 'allow'),
  ('p', 'anonymous', '/*', '/v2/', 'public|private', 'GET', 'allow'),
  ('p', 'anonymous', '/*', 'DS$*/**$blobs$*', 'public', 'GET|HEAD', 'allow'),
  ('p', 'anonymous', '/*', 'DS$*/**$manifests$*', 'public', 'GET|HEAD', 'allow'),
  ('p', 'namespace_reader', '/*', 'DS$*/**$blobs$*', 'public|private', 'GET|HEAD', 'allow'), -- get blob
  ('p', 'namespace_reader', '/*', 'DS$*/**$manifests$*', 'public|private', 'GET|HEAD', 'allow'), -- get manifest
  ('p', 'namespace_reader', '/*', 'API$*/**$namespaces/*', 'public|private', 'GET', 'allow'), -- get namespace
  ('p', 'namespace_reader', '/*', 'API$*/**$namespaces/*/artifacts/*', 'public|private', 'GET', 'allow'), -- get artifact
  ('p', 'namespace_reader', '/*', 'API$*/**$namespaces/*/artifacts/', 'public|private', 'GET', 'allow'), -- list artifacts
  ('p', 'namespace_reader', '/*', 'API$*/**$namespaces/*/repositories/', 'public|private', 'GET', 'allow'), -- list repositories
  ('p', 'namespace_reader', '/*', 'API$*/**$namespaces/*/repositories/*', 'public|private', 'GET', 'allow'), -- get repository
  ('p', 'namespace_admin', '/*', '*', 'public', 'GET|HEAD', 'allow'),
  ('p', 'namespace_owner', '/*', '*', 'public', 'GET|HEAD', 'allow');

INSERT INTO "namespaces" ("name", "visibility")
  VALUES ('library', 'public');

CREATE TABLE IF NOT EXISTS "webhooks" (
  "id" bigserial PRIMARY KEY,
  "namespace_id" bigint NOT NULL,
  "url" varchar(128) NOT NULL,
  "secret" varchar(63),
  "ssl_verify" smallint NOT NULL DEFAULT 1,
  "retry_times" smallint NOT NULL DEFAULT 3,
  "retry_duration" smallint NOT NULL DEFAULT 5,
  "event_namespace" smallint,
  "event_repository" smallint NOT NULL DEFAULT 1,
  "event_tag" smallint NOT NULL DEFAULT 1,
  "event_pull_push" smallint NOT NULL DEFAULT 1,
  "event_member" smallint NOT NULL DEFAULT 1,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "webhook_logs" (
  "id" bigserial PRIMARY KEY,
  "webhook_id" bigint NOT NULL,
  "event" varchar(128) NOT NULL,
  "status_code" smallint NOT NULL,
  "req_header" bytea NOT NULL,
  "req_body" bytea NOT NULL,
  "resp_header" bytea NOT NULL,
  "resp_body" bytea,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("webhook_id") REFERENCES "webhooks" ("id")
);

CREATE TABLE IF NOT EXISTS "builders" (
  "id" bigserial PRIMARY KEY,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "builder_logs" (
  "id" bigserial PRIMARY KEY,
  "builder_id" bigint NOT NULL,
  "log" bytea,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("builder_id") REFERENCES "builders" ("id")
);

