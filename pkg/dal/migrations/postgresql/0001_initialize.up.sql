CREATE TYPE user_3rdparty_provider AS ENUM (
  'github',
  'gitlab',
  'gitea'
);

CREATE TYPE daemon_status AS ENUM (
  'Pending',
  'Doing',
  'Success',
  'Failed'
);

CREATE TABLE IF NOT EXISTS "users" (
  "id" bigserial PRIMARY KEY,
  "username" varchar(64) NOT NULL,
  "password" varchar(256) NOT NULL,
  "email" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  CONSTRAINT "users_unique_with_username" UNIQUE ("username", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "user_3rdparty" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "provider" user_3rdparty_provider NOT NULL DEFAULT 'github',
  "account_id" varchar(256),
  "token" varchar(256),
  "refresh_token" varchar(256),
  "cr_last_update_timestamp" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "cr_last_update_status" daemon_status NOT NULL DEFAULT 'Doing',
  "cr_last_update_message" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("user_id") REFERENCES "users" ("id"),
  CONSTRAINT "user_3rdparty_unique_with_account_id" UNIQUE ("provider", "account_id", "deleted_at")
);

CREATE TYPE code_repository_clone_credentials_type AS enum (
  'none',
  'ssh',
  'username',
  'token'
);

CREATE TABLE IF NOT EXISTS "code_repository_clone_credentials" (
  "id" bigserial PRIMARY KEY,
  "user_3rdparty_id" bigint NOT NULL,
  "type" code_repository_clone_credentials_type NOT NULL,
  "ssh_key" bytea,
  "username" varchar(256),
  "password" varchar(256),
  "token" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("user_3rdparty_id") REFERENCES "user_3rdparty" ("id")
);

CREATE TABLE IF NOT EXISTS "code_repository_owners" (
  "id" bigserial PRIMARY KEY,
  "user_3rdparty_id" bigint NOT NULL,
  "is_org" smallint NOT NULL DEFAULT 0,
  "owner_id" varchar(256) NOT NULL,
  "owner" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("user_3rdparty_id") REFERENCES "user_3rdparty" ("id"),
  CONSTRAINT "code_repository_owners_unique_with_name" UNIQUE ("user_3rdparty_id", "owner_id", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "code_repositories" (
  "id" bigserial PRIMARY KEY,
  "user_3rdparty_id" bigint NOT NULL,
  "repository_id" varchar(256) NOT NULL,
  "is_org" smallint NOT NULL DEFAULT 0,
  "owner_id" varchar(256) NOT NULL,
  "owner" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "ssh_url" varchar(256) NOT NULL,
  "clone_url" varchar(256) NOT NULL,
  "oci_repo_count" bigint NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("user_3rdparty_id") REFERENCES "user_3rdparty" ("id"),
  CONSTRAINT "code_repositories_unique_with_name" UNIQUE ("user_3rdparty_id", "owner_id", "repository_id", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "code_repository_branches" (
  "id" bigserial PRIMARY KEY,
  "code_repository_id" integer NOT NULL,
  "name" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" integer NOT NULL DEFAULT 0,
  FOREIGN KEY ("code_repository_id") REFERENCES "code_repositories" ("id"),
  CONSTRAINT "code_repository_branches_unique_with_name" UNIQUE ("code_repository_id", "name", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "user_recover_codes" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "code" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
  'tag',
  'builder'
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
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("repository_id") REFERENCES "repositories" ("id"),
  CONSTRAINT "artifacts_unique_with_repo" UNIQUE ("repository_id", "digest", "deleted_at")
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
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("artifact_id") REFERENCES "artifacts" ("id"),
  CONSTRAINT "artifact_vulnerability_unique_with_artifact" UNIQUE ("artifact_id", "deleted_at")
);

CREATE TABLE IF NOT EXISTS "tags" (
  "id" bigserial PRIMARY KEY,
  "repository_id" bigint NOT NULL,
  "artifact_id" bigint NOT NULL,
  "name" varchar(128) NOT NULL,
  "pushed_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "last_pull" timestamp,
  "pull_times" bigint NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
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

CREATE TYPE builder_source AS ENUM (
  'SelfCodeRepository',
  'CodeRepository',
  'Dockerfile'
);

CREATE TABLE IF NOT EXISTS "builders" (
  "id" bigserial PRIMARY KEY,
  "repository_id" bigint NOT NULL,
  "source" builder_source NOT NULL,
  -- source SelfCodeRepository
  "scm_credential_type" varchar(16),
  "scm_repository" varchar(256),
  "scm_ssh_key" bytea,
  "scm_token" varchar(256),
  "scm_username" varchar(30),
  "scm_password" varchar(30),
  -- source CodeRepository
  "code_repository_id" bigint,
  -- source Dockerfile
  "dockerfile" bytea,
  -- common settings
  "scm_branch" varchar(256),
  "scm_depth" smallint,
  "scm_submodule" smallint,
  -- cron settings
  "cron_rules" varchar(30),
  "cron_branch" varchar(256),
  "cron_tag_template" varchar(256),
  "cron_next_trigger" timestamp,
  -- webhook settings
  "webhook_branch_name" varchar(256),
  "webhook_branch_tag_template" varchar(256),
  "webhook_tag_tag_template" varchar(256),
  -- buildkit settings
  "buildkit_insecure_registries" varchar(256),
  "buildkit_context" varchar(30),
  "buildkit_dockerfile" varchar(256),
  "buildkit_platforms" varchar(256) NOT NULL DEFAULT 'linux/amd64',
  "buildkit_build_args" varchar(256),
  -- other fields
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("repository_id") REFERENCES "repositories" ("id"),
  FOREIGN KEY ("code_repository_id") REFERENCES "code_repositories" ("id"),
  CONSTRAINT "builders_unique_with_repository" UNIQUE ("repository_id", "deleted_at")
);

CREATE TYPE builder_runner_status AS ENUM (
  'Pending',
  'Doing',
  'Success',
  'Failed',
  'Scheduling',
  'Stopping',
  'Stopped'
);

CREATE TABLE IF NOT EXISTS "builder_runners" (
  "id" bigserial PRIMARY KEY,
  "builder_id" bigint NOT NULL,
  "log" bytea,
  "status" builder_runner_status NOT NULL DEFAULT 'Pending',
  -- common settings
  "tag" varchar(128),
  "raw_tag" varchar(255) NOT NULL,
  "description" varchar(255),
  "scm_branch" varchar(255),
  "started_at" timestamp,
  "ended_at" timestamp,
  "duration" bigint,
  -- other fields
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "deleted_at" bigint NOT NULL DEFAULT 0,
  FOREIGN KEY ("builder_id") REFERENCES "builders" ("id")
);

CREATE TABLE IF NOT EXISTS "work_queues" (
  "id" bigserial PRIMARY KEY,
  "topic" varchar(30) NOT NULL,
  "payload" bytea NOT NULL,
  "times" smallint NOT NULL DEFAULT 0,
  "version" varchar(36) NOT NULL,
  "status" daemon_status NOT NULL DEFAULT 'Pending',
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "caches" (
  "id" bigserial PRIMARY KEY,
  "key" varchar(256) NOT NULL UNIQUE,
  "val" bytea NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" integer NOT NULL DEFAULT 0
);

CREATE INDEX "idx_created_at" ON "caches" ("created_at");

CREATE TABLE IF NOT EXISTS "settings" (
  "id" bigserial PRIMARY KEY,
  "key" varchar(256) NOT NULL UNIQUE,
  "val" bytea,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" bigint NOT NULL DEFAULT 0
);

