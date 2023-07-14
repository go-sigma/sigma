CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `provider` ENUM ('local', 'github') NOT NULL DEFAULT 'local',
  `username` varchar(64) NOT NULL UNIQUE,
  `password` varchar(256),
  `email` varchar(256),
  `provider_account_id` varchar(256),
  `refresh_token` varchar(256),
  `access_token` varchar(256),
  `expires_at` bigint,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  CONSTRAINT `users_unique_with_username` UNIQUE (`username`, `deleted_at`),
  CONSTRAINT `users_unique_with_email` UNIQUE (`email`, `deleted_at`)
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
  `pushed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_pull` timestamp,
  `pull_times` bigint NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`),
  CONSTRAINT `artifacts_unique_with_repo` UNIQUE (`repository_id`, `digest`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `artifact_sboms` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `artifact_id` bigint NOT NULL,
  `raw` MEDIUMBLOB,
  `status` varchar(64) NOT NULL,
  `stdout` MEDIUMBLOB,
  `stderr` MEDIUMBLOB,
  `message` varchar(256),
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `artifact_sbom_unique_with_artifact` UNIQUE (`artifact_id`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `artifact_vulnerabilities` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `artifact_id` bigint NOT NULL,
  `metadata` BLOB,
  `raw` MEDIUMBLOB,
  `status` varchar(64) NOT NULL,
  `stdout` MEDIUMBLOB,
  `stderr` MEDIUMBLOB,
  `message` varchar(256),
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `artifact_vulnerability_unique_with_artifact` UNIQUE (`artifact_id`, `deleted_at`)
);

CREATE TABLE IF NOT EXISTS `tags` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `repository_id` bigint NOT NULL,
  `artifact_id` bigint NOT NULL,
  `name` varchar(64) NOT NULL,
  `pushed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_pull` timestamp,
  `pull_times` bigint NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
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
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
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
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
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

