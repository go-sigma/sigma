CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `username` varchar(64) NOT NULL,
  `password` varchar(256),
  `email` varchar(256),
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  CONSTRAINT `users_unique_with_username` UNIQUE (`username`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `user_3rdparty` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` bigint NOT NULL,
  `provider` ENUM ('github', 'gitlab', 'gitea') NOT NULL,
  `account_id` varchar(256),
  `token` varchar(256),
  `refresh_token` varchar(256),
  `cr_last_update_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `cr_last_update_status` ENUM ('Success', 'Failed', 'Doing') NOT NULL DEFAULT 'Doing',
  `cr_last_update_message` varchar(256),
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `user_3rdparty_unique_with_account_id` UNIQUE (`provider`, `account_id`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `code_repository_clone_credentials` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_3rdparty_id` bigint NOT NULL,
  `type` ENUM ('none', 'ssh', 'username', 'token') NOT NULL,
  `ssh_key` BLOB,
  `username` varchar(256),
  `password` varchar(256),
  `token` varchar(256),
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_3rdparty_id`) REFERENCES `user_3rdparty` (`id`)
);

CREATE TABLE IF NOT EXISTS `code_repository_owners` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_3rdparty_id` bigint NOT NULL,
  `is_org` tinyint NOT NULL DEFAULT 0,
  `owner_id` varchar(256) NOT NULL,
  `owner` varchar(256) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_3rdparty_id`) REFERENCES `user_3rdparty` (`id`),
  CONSTRAINT `code_repository_owners_unique_with_name` UNIQUE (`user_3rdparty_id`, `owner_id`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `code_repositories` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_3rdparty_id` bigint NOT NULL,
  `repository_id` varchar(256) NOT NULL,
  `is_org` tinyint NOT NULL DEFAULT 0,
  `owner_id` varchar(256) NOT NULL,
  `owner` varchar(256) NOT NULL,
  `name` varchar(256) NOT NULL,
  `ssh_url` varchar(256) NOT NULL,
  `clone_url` varchar(256) NOT NULL,
  `oci_repo_count` bigint NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_3rdparty_id`) REFERENCES `user_3rdparty` (`id`),
  CONSTRAINT `code_repositories_unique_with_name` UNIQUE (`user_3rdparty_id`, `owner_id`, `repository_id`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `code_repository_branches` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `code_repository_id` bigint NOT NULL,
  `name` varchar(256) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`code_repository_id`) REFERENCES `code_repositories` (`id`),
  CONSTRAINT `code_repository_branches_unique_with_name` UNIQUE (`code_repository_id`, `name`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `user_recover_codes` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` bigint NOT NULL,
  `code` varchar(256) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `user_recover_codes_unique_with_use_id` UNIQUE (`user_id`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `namespaces` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `name` varchar(64) NOT NULL,
  `description` varchar(256),
  `visibility` ENUM ('public', 'private') NOT NULL DEFAULT 'private',
  `size_limit` bigint NOT NULL DEFAULT 0,
  `size` bigint NOT NULL DEFAULT 0,
  `repository_limit` bigint NOT NULL DEFAULT 0,
  `repository_count` bigint NOT NULL DEFAULT 0,
  `tag_limit` bigint NOT NULL DEFAULT 0,
  `tag_count` bigint NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  CONSTRAINT `namespaces_unique_with_name` UNIQUE (`name`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `audits` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` bigint NOT NULL,
  `namespace_id` bigint,
  `action` ENUM ('create', 'update', 'delete', 'pull', 'push') NOT NULL,
  `resource_type` ENUM ('namespace', 'repository', 'tag', 'builder') NOT NULL,
  `resource` varchar(256) NOT NULL,
  `before_raw` BLOB,
  `req_raw` BLOB,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`)
);

CREATE TABLE IF NOT EXISTS `repositories` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `name` varchar(64) NOT NULL,
  `description` varchar(255),
  `overview` BLOB,
  `visibility` ENUM ('public', 'private') NOT NULL DEFAULT 'private',
  `size_limit` bigint NOT NULL DEFAULT 0,
  `size` bigint NOT NULL DEFAULT 0,
  `tag_limit` bigint NOT NULL DEFAULT 0,
  `tag_count` bigint NOT NULL DEFAULT 0,
  `namespace_id` bigint NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`),
  CONSTRAINT `repositories_unique_with_namespace` UNIQUE (`namespace_id`, `name`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `artifacts` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `repository_id` bigint NOT NULL,
  `digest` varchar(256) NOT NULL,
  `size` bigint NOT NULL DEFAULT 0,
  `blobs_size` bigint NOT NULL DEFAULT 0,
  `content_type` varchar(256) NOT NULL,
  `raw` MEDIUMBLOB NOT NULL,
  `config_raw` MEDIUMBLOB,
  `config_media_type` varchar(256),
  `type` ENUM ('image', 'imageIndex', 'chart', 'cnab', 'wasm', 'provenance', 'cosign', 'unknown') NOT NULL DEFAULT 'unknown',
  `pushed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_pull` timestamp,
  `pull_times` bigint NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`),
  CONSTRAINT `artifacts_unique_with_repo` UNIQUE (`repository_id`, `digest`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `artifact_sboms` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `artifact_id` bigint NOT NULL,
  `raw` MEDIUMBLOB,
  `result` MEDIUMBLOB,
  `status` varchar(64) NOT NULL,
  `stdout` MEDIUMBLOB,
  `stderr` MEDIUMBLOB,
  `message` varchar(256),
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `artifact_sbom_unique_with_artifact` UNIQUE (`artifact_id`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `artifact_vulnerabilities` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `artifact_id` bigint NOT NULL,
  `metadata` BLOB,
  `raw` MEDIUMBLOB,
  `result` MEDIUMBLOB,
  `status` varchar(64) NOT NULL,
  `stdout` MEDIUMBLOB,
  `stderr` MEDIUMBLOB,
  `message` varchar(256),
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `artifact_vulnerability_unique_with_artifact` UNIQUE (`artifact_id`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `tags` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `repository_id` bigint NOT NULL,
  `artifact_id` bigint NOT NULL,
  `name` varchar(128) NOT NULL,
  `pushed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_pull` timestamp,
  `pull_times` bigint NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`),
  FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `tags_unique_with_repo` UNIQUE (`repository_id`, `name`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `blobs` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `digest` varchar(256) NOT NULL UNIQUE,
  `size` bigint NOT NULL,
  `content_type` varchar(256) NOT NULL,
  `pushed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_pull` timestamp,
  `pull_times` bigint NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  CONSTRAINT `blobs_unique_with_digest` UNIQUE (`digest`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `blob_uploads` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `part_number` int NOT NULL,
  `upload_id` varchar(256) NOT NULL,
  `etag` varchar(256) NOT NULL,
  `repository` varchar(256) NOT NULL,
  `file_id` varchar(256) NOT NULL,
  `size` bigint NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  CONSTRAINT `blob_uploads_unique_with_upload_id_etag` UNIQUE (`upload_id`, `etag`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `artifact_artifacts` (
  `artifact_id` bigint NOT NULL,
  `artifact_index_id` bigint NOT NULL,
  PRIMARY KEY (`artifact_id`, `artifact_index_id`),
  CONSTRAINT `fk_artifact_artifacts_artifact` FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `fk_artifact_artifacts_artifact_index` FOREIGN KEY (`artifact_index_id`) REFERENCES `artifacts` (`id`)
);

CREATE TABLE IF NOT EXISTS `artifact_blobs` (
  `artifact_id` bigint NOT NULL,
  `blob_id` bigint NOT NULL,
  PRIMARY KEY (`artifact_id`, `blob_id`),
  CONSTRAINT `fk_artifact_blobs_artifact` FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `fk_artifact_blobs_blob` FOREIGN KEY (`blob_id`) REFERENCES `blobs` (`id`)
);

CREATE TABLE IF NOT EXISTS `daemon_logs` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `namespace_id` bigint,
  `type` ENUM ('Gc', 'Vulnerability', 'Sbom') NOT NULL,
  `action` ENUM ('create', 'update', 'delete', 'pull', 'push') NOT NULL,
  `resource` varchar(256) NOT NULL,
  `status` ENUM ('Success', 'Failed', 'Pending', 'Doing') NOT NULL,
  `message` BLOB,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0
);

CREATE TABLE `casbin_rules` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `ptype` varchar(100),
  `v0` varchar(100),
  `v1` varchar(100),
  `v2` varchar(100),
  `v3` varchar(100),
  `v4` varchar(100),
  `v5` varchar(100),
  CONSTRAINT `idx_casbin_rules` UNIQUE (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`)
);

-- ptype type
-- v0 sub
-- v1 dom
-- v2 url
-- v3 attr
-- v4 method
-- v5 allow or deny
INSERT INTO `casbin_rules` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`)
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

INSERT INTO `namespaces` (`name`, `visibility`)
  VALUES ('library', 'public');

CREATE TABLE IF NOT EXISTS `webhooks` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `namespace_id` bigint NOT NULL,
  `url` varchar(128) NOT NULL,
  `secret` varchar(63),
  `ssl_verify` tinyint NOT NULL DEFAULT 1,
  `retry_times` tinyint NOT NULL DEFAULT 3,
  `retry_duration` tinyint NOT NULL DEFAULT 5,
  `event_namespace` tinyint,
  `event_repository` tinyint NOT NULL DEFAULT 1,
  `event_tag` tinyint NOT NULL DEFAULT 1,
  `event_pull_push` tinyint NOT NULL DEFAULT 1,
  `event_member` tinyint NOT NULL DEFAULT 1,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS `webhook_logs` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `webhook_id` bigint NOT NULL,
  `event` varchar(128) NOT NULL,
  `status_code` smallint NOT NULL,
  `req_header` BLOB NOT NULL,
  `req_body` BLOB NOT NULL,
  `resp_header` BLOB NOT NULL,
  `resp_body` BLOB,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`webhook_id`) REFERENCES `webhooks` (`id`)
);

CREATE TABLE IF NOT EXISTS `builders` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `repository_id` bigint NOT NULL,
  `source` ENUM ('SelfCodeRepository', 'CodeRepository', 'Dockerfile') NOT NULL,
  -- source SelfCodeRepository
  `scm_repository` varchar(256),
  `scm_credential_type` varchar(16),
  `scm_ssh_key` BLOB,
  `scm_token` varchar(256),
  `scm_username` varchar(30),
  `scm_password` varchar(30),
  -- source CodeRepository
  `code_repository_id` bigint,
  -- source Dockerfile
  `dockerfile` BLOB,
  -- common settings
  `scm_branch` varchar(256),
  `scm_depth` MEDIUMINT,
  `scm_submodule` tinyint,
  -- cron settings
  `cron_rule` varchar(30),
  `cron_branch` varchar(256),
  `cron_tag_template` varchar(256),
  `cron_next_trigger` timestamp,
  -- webhook settings
  `webhook_branch_name` varchar(256),
  `webhook_branch_tag_template` varchar(256),
  `webhook_tag_tag_template` varchar(256),
  -- buildkit settings
  `buildkit_insecure_registries` varchar(256),
  `buildkit_context` varchar(30),
  `buildkit_dockerfile` varchar(256),
  `buildkit_platforms` varchar(256) NOT NULL DEFAULT 'linux/amd64',
  `buildkit_build_args` varchar(256),
  -- other fields
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`),
  FOREIGN KEY (`code_repository_id`) REFERENCES `code_repositories` (`id`),
  CONSTRAINT `builders_unique_with_repository` UNIQUE (`repository_id`, `deleted_at`)
);

-- TODO: buildx flags
CREATE TABLE IF NOT EXISTS `builder_runners` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `builder_id` bigint NOT NULL,
  `log` LONGBLOB,
  `status` ENUM ('Success', 'Failed', 'Pending', 'Scheduling', 'Building', 'Stopping', 'Stopped') NOT NULL DEFAULT 'Pending',
  -- common settings
  `tag` varchar(128),
  `raw_tag` varchar(255) NOT NULL,
  `description` varchar(255),
  `scm_branch` varchar(255),
  `started_at` timestamp,
  `ended_at` timestamp,
  `duration` bigint,
  -- other fields
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`builder_id`) REFERENCES `builders` (`id`)
);

CREATE TABLE IF NOT EXISTS `work_queues` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `topic` varchar(30) NOT NULL,
  `payload` BLOB NOT NULL,
  `times` MEDIUMINT NOT NULL DEFAULT 0,
  `version` varchar(36) NOT NULL,
  `status` ENUM ('Success', 'Failed', 'Pending', 'Doing') NOT NULL DEFAULT 'Pending',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS `caches` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `key` varchar(256) NOT NULL UNIQUE,
  `val` BLOB NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0
);

CREATE INDEX `idx_created_at` ON `caches` (`created_at`);

CREATE TABLE IF NOT EXISTS `settings` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `key` varchar(256) NOT NULL UNIQUE,
  `val` BLOB,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` bigint NOT NULL DEFAULT 0
);

