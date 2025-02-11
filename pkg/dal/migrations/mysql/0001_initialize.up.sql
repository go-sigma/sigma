CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `username` varchar(64) NOT NULL,
  `password` varchar(256),
  `email` varchar(256),
  `last_login` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `namespace_limit` bigint NOT NULL DEFAULT 0,
  `namespace_count` bigint NOT NULL DEFAULT 0,
  `status` ENUM ('Active', 'Inactive') NOT NULL DEFAULT 'Active',
  `role` ENUM ('Root', 'Admin', 'User', 'Anonymous') NOT NULL DEFAULT 'User',
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  UNIQUE KEY `users_unique_with_username` (`username`, `deleted_at`),
  KEY `users_idx_status` (`status`),
  KEY `users_idx_role` (`role`),
  KEY `users_idx_last_login` (`last_login`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_3rdparty` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` bigint NOT NULL,
  `provider` ENUM ('github', 'gitlab', 'gitea') NOT NULL,
  `account_id` varchar(256),
  `token` varchar(256),
  `refresh_token` varchar(256),
  `cr_last_update_timestamp` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `cr_last_update_status` ENUM ('Success', 'Failed', 'Doing') NOT NULL DEFAULT 'Doing',
  `cr_last_update_message` varchar(256),
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `user_3rdparty_unique_with_account_id` UNIQUE (`provider`, `account_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `code_repository_clone_credentials` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_3rdparty_id` bigint NOT NULL,
  `type` ENUM ('none', 'ssh', 'username', 'token') NOT NULL,
  `ssh_key` BLOB,
  `username` varchar(256),
  `password` varchar(256),
  `token` varchar(256),
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_3rdparty_id`) REFERENCES `user_3rdparty` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `code_repository_owners` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_3rdparty_id` bigint NOT NULL,
  `is_org` tinyint NOT NULL DEFAULT 0,
  `owner_id` varchar(256) NOT NULL,
  `owner` varchar(256) NOT NULL,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_3rdparty_id`) REFERENCES `user_3rdparty` (`id`),
  CONSTRAINT `code_repository_owners_unique_with_name` UNIQUE (`user_3rdparty_id`, `owner_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

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
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_3rdparty_id`) REFERENCES `user_3rdparty` (`id`),
  CONSTRAINT `code_repositories_unique_with_name` UNIQUE (`user_3rdparty_id`, `owner_id`, `repository_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `code_repository_branches` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `code_repository_id` bigint NOT NULL,
  `name` varchar(256) NOT NULL,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`code_repository_id`) REFERENCES `code_repositories` (`id`),
  CONSTRAINT `code_repository_branches_unique_with_name` UNIQUE (`code_repository_id`, `name`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_recover_codes` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` bigint NOT NULL,
  `code` varchar(256) NOT NULL,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `user_recover_codes_unique_with_use_id` UNIQUE (`user_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `namespaces` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `name` varchar(64) NOT NULL,
  `description` varchar(256),
  `overview` BLOB,
  `visibility` ENUM ('public', 'private') NOT NULL DEFAULT 'private',
  `size_limit` bigint NOT NULL DEFAULT 0,
  `size` bigint NOT NULL DEFAULT 0,
  `repository_limit` bigint NOT NULL DEFAULT 0,
  `repository_count` bigint NOT NULL DEFAULT 0,
  `tag_limit` bigint NOT NULL DEFAULT 0,
  `tag_count` bigint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  CONSTRAINT `namespaces_unique_with_name` UNIQUE (`name`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `audits` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` bigint NOT NULL,
  `namespace_id` bigint,
  `action` ENUM ('Create', 'Update', 'Delete', 'Pull', 'Push') NOT NULL,
  `resource_type` ENUM ('Namespace', 'Repository', 'Tag', 'Builder', 'Webhook', 'NamespaceMember') NOT NULL,
  `resource` varchar(256) NOT NULL,
  `req_raw` BLOB,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `repositories` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `namespace_id` bigint NOT NULL,
  `name` varchar(64) NOT NULL,
  `description` varchar(255),
  `overview` BLOB,
  `size_limit` bigint NOT NULL DEFAULT 0,
  `size` bigint NOT NULL DEFAULT 0,
  `tag_limit` bigint NOT NULL DEFAULT 0,
  `tag_count` bigint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`),
  CONSTRAINT `repositories_unique_with_namespace` UNIQUE (`namespace_id`, `name`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

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
  `type` ENUM ('Image', 'ImageIndex', 'Chart', 'Cnab', 'Wasm', 'Provenance', 'Cosign', 'Sif', 'Unknown') NOT NULL DEFAULT 'Unknown',
  `pushed_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `last_pull` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `pull_times` bigint NOT NULL DEFAULT 0,
  `referrer_id` bigint,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`),
  FOREIGN KEY (`referrer_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `artifacts_unique_with_repo` UNIQUE (`repository_id`, `digest`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `artifact_sboms` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `artifact_id` bigint NOT NULL,
  `raw` MEDIUMBLOB,
  `result` MEDIUMBLOB,
  `status` varchar(64) NOT NULL,
  `stdout` MEDIUMBLOB,
  `stderr` MEDIUMBLOB,
  `message` varchar(256),
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `artifact_sbom_unique_with_artifact` UNIQUE (`artifact_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

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
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `artifact_vulnerability_unique_with_artifact` UNIQUE (`artifact_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `tags` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `repository_id` bigint NOT NULL,
  `artifact_id` bigint NOT NULL,
  `name` varchar(128) NOT NULL,
  `pushed_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `last_pull` bigint,
  `pull_times` bigint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`),
  FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `tags_unique_with_repo` UNIQUE (`repository_id`, `name`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `blobs` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `digest` varchar(256) NOT NULL UNIQUE,
  `size` bigint NOT NULL,
  `content_type` varchar(256) NOT NULL,
  `pushed_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `last_pull` bigint,
  `pull_times` bigint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  CONSTRAINT `blobs_unique_with_digest` UNIQUE (`digest`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `blob_uploads` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `part_number` int NOT NULL,
  `upload_id` varchar(256) NOT NULL,
  `etag` varchar(256) NOT NULL,
  `repository` varchar(256) NOT NULL,
  `file_id` varchar(256) NOT NULL,
  `size` bigint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  CONSTRAINT `blob_uploads_unique_with_upload_id_etag` UNIQUE (`upload_id`, `etag`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `artifact_artifacts` (
  `artifact_id` bigint NOT NULL,
  `artifact_index_id` bigint NOT NULL,
  PRIMARY KEY (`artifact_id`, `artifact_index_id`),
  CONSTRAINT `fk_artifact_artifacts_artifact` FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `fk_artifact_artifacts_artifact_index` FOREIGN KEY (`artifact_index_id`) REFERENCES `artifacts` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `artifact_blobs` (
  `artifact_id` bigint NOT NULL,
  `blob_id` bigint NOT NULL,
  PRIMARY KEY (`artifact_id`, `blob_id`),
  CONSTRAINT `fk_artifact_blobs_artifact` FOREIGN KEY (`artifact_id`) REFERENCES `artifacts` (`id`),
  CONSTRAINT `fk_artifact_blobs_blob` FOREIGN KEY (`blob_id`) REFERENCES `blobs` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_tag_rules` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `namespace_id` bigint,
  `is_running` tinyint NOT NULL DEFAULT 0,
  `cron_enabled` tinyint NOT NULL DEFAULT 0,
  `cron_rule` varchar(30),
  `cron_next_trigger` bigint,
  `retention_rule_type` ENUM ('Day', 'Quantity') NOT NULL DEFAULT 'Quantity',
  `retention_rule_amount` bigint NOT NULL DEFAULT 1,
  `retention_pattern` varchar(64),
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`),
  CONSTRAINT `daemon_gc_tag_rules_unique_with_ns` UNIQUE (`namespace_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_tag_runners` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `rule_id` bigint NOT NULL,
  `message` LONGBLOB,
  `status` ENUM ('Success', 'Failed', 'Pending', 'Doing') NOT NULL DEFAULT 'Pending',
  `operate_type` ENUM ('Automatic', 'Manual') NOT NULL DEFAULT 'Automatic',
  `operate_user_id` bigint,
  `started_at` bigint,
  `ended_at` bigint,
  `duration` bigint,
  `success_count` bigint,
  `failed_count` bigint,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`rule_id`) REFERENCES `daemon_gc_tag_rules` (`id`),
  FOREIGN KEY (`operate_user_id`) REFERENCES `users` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_tag_records` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `runner_id` bigint NOT NULL,
  `tag` varchar(128) NOT NULL,
  `status` ENUM ('Success', 'Failed') NOT NULL DEFAULT 'Success',
  `message` LONGBLOB,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`runner_id`) REFERENCES `daemon_gc_tag_runners` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_repository_rules` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `namespace_id` bigint,
  `is_running` tinyint NOT NULL DEFAULT 0,
  `retention_day` int NOT NULL DEFAULT 0,
  `cron_enabled` tinyint NOT NULL DEFAULT 0,
  `cron_rule` varchar(30),
  `cron_next_trigger` bigint,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`),
  CONSTRAINT `daemon_gc_repository_rules_unique_with_ns` UNIQUE (`namespace_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_repository_runners` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `rule_id` bigint NOT NULL,
  `message` LONGBLOB,
  `status` ENUM ('Success', 'Failed', 'Pending', 'Doing') NOT NULL DEFAULT 'Pending',
  `operate_type` ENUM ('Automatic', 'Manual') NOT NULL DEFAULT 'Automatic',
  `operate_user_id` bigint,
  `started_at` bigint,
  `ended_at` bigint,
  `duration` bigint,
  `success_count` bigint,
  `failed_count` bigint,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`rule_id`) REFERENCES `daemon_gc_repository_rules` (`id`),
  FOREIGN KEY (`operate_user_id`) REFERENCES `users` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_repository_records` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `runner_id` bigint NOT NULL,
  `repository` varchar(64) NOT NULL,
  `status` ENUM ('Success', 'Failed') NOT NULL DEFAULT 'Success',
  `message` LONGBLOB,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`runner_id`) REFERENCES `daemon_gc_repository_runners` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_artifact_rules` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `namespace_id` bigint,
  `is_running` tinyint NOT NULL DEFAULT 0,
  `retention_day` int NOT NULL DEFAULT 0,
  `cron_enabled` tinyint NOT NULL DEFAULT 0,
  `cron_rule` varchar(30),
  `cron_next_trigger` bigint,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`),
  CONSTRAINT `daemon_gc_artifact_rules_unique_with_ns` UNIQUE (`namespace_id`, `deleted_at`),
  KEY `daemon_gc_artifact_rules_idx_created_at` (`created_at`),
  KEY `daemon_gc_artifact_rules_idx_updated_at` (`updated_at`),
  KEY `daemon_gc_artifact_rules_idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_artifact_runners` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `rule_id` bigint NOT NULL,
  `message` LONGBLOB,
  `status` ENUM ('Success', 'Failed', 'Pending', 'Doing') NOT NULL DEFAULT 'Pending',
  `operate_type` ENUM ('Automatic', 'Manual') NOT NULL DEFAULT 'Automatic',
  `operate_user_id` bigint,
  `started_at` bigint,
  `ended_at` bigint,
  `duration` bigint,
  `success_count` bigint,
  `failed_count` bigint,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`rule_id`) REFERENCES `daemon_gc_artifact_rules` (`id`),
  FOREIGN KEY (`operate_user_id`) REFERENCES `users` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_artifact_records` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `runner_id` bigint NOT NULL,
  `digest` varchar(256) NOT NULL,
  `status` ENUM ('Success', 'Failed') NOT NULL DEFAULT 'Success',
  `message` LONGBLOB,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`runner_id`) REFERENCES `daemon_gc_artifact_runners` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_blob_rules` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `is_running` tinyint NOT NULL DEFAULT 0,
  `retention_day` int NOT NULL DEFAULT 0,
  `cron_enabled` tinyint NOT NULL DEFAULT 0,
  `cron_rule` varchar(30),
  `cron_next_trigger` bigint,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_blob_runners` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `rule_id` bigint NOT NULL,
  `message` LONGBLOB,
  `status` ENUM ('Success', 'Failed', 'Pending', 'Doing') NOT NULL DEFAULT 'Pending',
  `operate_type` ENUM ('Automatic', 'Manual') NOT NULL DEFAULT 'Automatic',
  `operate_user_id` bigint,
  `started_at` bigint,
  `ended_at` bigint,
  `duration` bigint,
  `success_count` bigint,
  `failed_count` bigint,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`rule_id`) REFERENCES `daemon_gc_blob_rules` (`id`),
  FOREIGN KEY (`operate_user_id`) REFERENCES `users` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_blob_records` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `runner_id` bigint NOT NULL,
  `digest` varchar(256) NOT NULL,
  `status` ENUM ('Success', 'Failed') NOT NULL DEFAULT 'Success',
  `message` LONGBLOB,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`runner_id`) REFERENCES `daemon_gc_blob_runners` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `casbin_rules` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `ptype` varchar(100),
  `v0` varchar(100),
  `v1` varchar(100),
  `v2` varchar(100),
  `v3` varchar(100),
  `v4` varchar(100),
  `v5` varchar(100),
  CONSTRAINT `idx_casbin_rules` UNIQUE (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `namespace_members` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` bigint NOT NULL,
  `namespace_id` bigint NOT NULL,
  `role` ENUM ('NamespaceReader', 'NamespaceManager', 'NamespaceAdmin') NOT NULL DEFAULT 'NamespaceReader',
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `namespace_members_unique_with_user_ns_role` UNIQUE (`user_id`, `namespace_id`, `role`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ptype type
-- v0 sub
-- v1 dom
-- v2 url
-- v3 attr
-- v4 method
-- v5 allow or deny
INSERT INTO `casbin_rules` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`)
  VALUES ('p', 'Root', '*', '*', '*', '*', 'allow'),
  ('p', 'Admin', '*', '*', '*', '*', 'allow'),
  ('p', 'Anonymous', '/*', '/v2/', 'public|private', 'GET', 'allow'),
  ('p', 'Anonymous', '/*', 'DS$*/**$blobs$*', 'public', 'GET|HEAD', 'allow'),
  ('p', 'Anonymous', '/*', 'DS$*/**$manifests$*', 'public', 'GET|HEAD', 'allow'),
  ('p', 'NamespaceReader', '/*', 'DS$*/**$blobs$*', 'public|private', 'GET|HEAD', 'allow'), -- get blob
  ('p', 'NamespaceReader', '/*', 'DS$*/**$manifests$*', 'public|private', 'GET|HEAD', 'allow'), -- get manifest
  ('p', 'NamespaceReader', '/*', 'API$*/**$namespaces/*', 'public|private', 'GET', 'allow'), -- get namespace
  ('p', 'NamespaceReader', '/*', 'API$*/**$namespaces/*/artifacts/*', 'public|private', 'GET', 'allow'), -- get artifact
  ('p', 'NamespaceReader', '/*', 'API$*/**$namespaces/*/artifacts/', 'public|private', 'GET', 'allow'), -- list artifacts
  ('p', 'NamespaceReader', '/*', 'API$*/**$namespaces/*/repositories/', 'public|private', 'GET', 'allow'), -- list repositories
  ('p', 'NamespaceReader', '/*', 'API$*/**$namespaces/*/repositories/*', 'public|private', 'GET', 'allow'), -- get repository
  ('p', 'NamespaceManager', '/*', '*', 'public', 'GET|HEAD', 'allow'),
  ('p', 'NamespaceAdmin', '/*', '*', 'public', 'GET|HEAD', 'allow');

INSERT INTO `namespaces` (`name`, `visibility`)
  VALUES ('library', 'public');

CREATE TABLE IF NOT EXISTS `webhooks` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `namespace_id` bigint,
  `url` varchar(128) NOT NULL,
  `secret` varchar(63),
  `enable` tinyint NOT NULL DEFAULT 1,
  `ssl_verify` tinyint NOT NULL DEFAULT 1,
  `retry_times` tinyint NOT NULL DEFAULT 1,
  `retry_duration` tinyint NOT NULL DEFAULT 5,
  `event_namespace` tinyint,
  `event_repository` tinyint NOT NULL DEFAULT 0,
  `event_tag` tinyint NOT NULL DEFAULT 0,
  `event_artifact` tinyint NOT NULL DEFAULT 0,
  `event_member` tinyint NOT NULL DEFAULT 0,
  `event_daemon_task_gc` tinyint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `webhook_logs` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `webhook_id` bigint,
  `resource_type` ENUM ('Webhook', 'Namespace', 'Repository', 'Tag', 'Artifact', 'Member', 'DaemonTaskGcRepositoryRule', 'DaemonTaskGcTagRule', 'DaemonTaskGcArtifactRule', 'DaemonTaskGcBlobRule', 'DaemonTaskGcRepositoryRunner', 'DaemonTaskGcTagRunner', 'DaemonTaskGcArtifactRunner', 'DaemonTaskGcBlobRunner') NOT NULL,
  `action` ENUM ('Create', 'Update', 'Delete', 'Add', 'Remove', 'Ping', 'Started', 'Finished') NOT NULL,
  `status_code` smallint NOT NULL,
  `req_header` BLOB NOT NULL,
  `req_body` BLOB NOT NULL,
  `resp_header` BLOB NOT NULL,
  `resp_body` BLOB,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`webhook_id`) REFERENCES `webhooks` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

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
  `cron_next_trigger` bigint,
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
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`),
  FOREIGN KEY (`code_repository_id`) REFERENCES `code_repositories` (`id`),
  CONSTRAINT `builders_unique_with_repository` UNIQUE (`repository_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- TODO: buildx flags
CREATE TABLE IF NOT EXISTS `builder_runners` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `builder_id` bigint NOT NULL,
  `log` LONGBLOB,
  `status` ENUM ('Success', 'Failed', 'Pending', 'Scheduling', 'Building', 'Stopping', 'Stopped') NOT NULL DEFAULT 'Pending',
  `status_message` varchar(255),
  -- common settings
  `tag` varchar(128),
  `raw_tag` varchar(255) NOT NULL,
  `description` varchar(255),
  `scm_branch` varchar(255),
  `started_at` bigint,
  `ended_at` bigint,
  `duration` bigint,
  -- other fields
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  FOREIGN KEY (`builder_id`) REFERENCES `builders` (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `work_queues` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `topic` varchar(30) NOT NULL,
  `payload` BLOB NOT NULL,
  `times` MEDIUMINT NOT NULL DEFAULT 0,
  `version` varchar(36) NOT NULL,
  `status` ENUM ('Success', 'Failed', 'Pending', 'Doing') NOT NULL DEFAULT 'Pending',
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `caches` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `key` varchar(256) NOT NULL UNIQUE,
  `val` BLOB NOT NULL,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE INDEX `idx_created_at` ON `caches` (`created_at`) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `settings` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `key` varchar(256) NOT NULL UNIQUE,
  `val` BLOB,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

