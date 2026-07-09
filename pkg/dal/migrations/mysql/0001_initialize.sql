-- +goose Up
CREATE TABLE IF NOT EXISTS `users` (
    `id` varchar(36) PRIMARY KEY,
    `username` varchar(64) NOT NULL,
    `password` varchar(256),
    `email` varchar(256),
    `last_login` bigint NOT NULL,
    `namespace_limit` bigint NOT NULL DEFAULT 0,
    `namespace_count` bigint NOT NULL DEFAULT 0,
    `status` varchar(64) NOT NULL DEFAULT 'Active',
    `role` varchar(64) NOT NULL DEFAULT 'User',
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    UNIQUE KEY `users_unique_with_username` (`username`, `deleted_at`),
    KEY `users_idx_status` (`status`),
    KEY `users_idx_role` (`role`),
    KEY `users_idx_last_login` (`last_login`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_3rdparty` (
    `id` varchar(36) PRIMARY KEY,
    `user_id` varchar(36) NOT NULL,
    `provider` varchar(64) NOT NULL,
    `account_id` varchar(256),
    `token` varchar(256),
    `refresh_token` varchar(256),
    `cr_last_update_timestamp` bigint NOT NULL,
    `cr_last_update_status` varchar(64) NOT NULL DEFAULT 'Doing',
    `cr_last_update_message` varchar(256),
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `user_3rdparty_unique_with_account_id` UNIQUE (
        `provider`,
        `account_id`,
        `deleted_at`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `code_repository_clone_credentials` (
    `id` varchar(36) PRIMARY KEY,
    `user_3rdparty_id` varchar(36) NOT NULL,
    `type` varchar(64) NOT NULL,
    `ssh_key` BLOB,
    `username` varchar(256),
    `password` varchar(256),
    `token` varchar(256),
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `code_repository_owners` (
    `id` varchar(36) PRIMARY KEY,
    `user_3rdparty_id` varchar(36) NOT NULL,
    `is_org` tinyint NOT NULL DEFAULT 0,
    `owner_id` varchar(256) NOT NULL,
    `owner` varchar(256) NOT NULL,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `code_repository_owners_unique_with_name` UNIQUE (
        `user_3rdparty_id`,
        `owner_id`,
        `deleted_at`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `code_repositories` (
    `id` varchar(36) PRIMARY KEY,
    `user_3rdparty_id` varchar(36) NOT NULL,
    `repository_id` varchar(256) NOT NULL,
    `is_org` tinyint NOT NULL DEFAULT 0,
    `owner_id` varchar(256) NOT NULL,
    `owner` varchar(256) NOT NULL,
    `name` varchar(256) NOT NULL,
    `ssh_url` varchar(256) NOT NULL,
    `clone_url` varchar(256) NOT NULL,
    `oci_repo_count` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `code_repositories_unique_with_name` UNIQUE (
        `user_3rdparty_id`,
        `owner_id`,
        `repository_id`,
        `deleted_at`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `code_repository_branches` (
    `id` varchar(36) PRIMARY KEY,
    `code_repository_id` varchar(36) NOT NULL,
    `name` varchar(256) NOT NULL,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `code_repository_branches_unique_with_name` UNIQUE (
        `code_repository_id`,
        `name`,
        `deleted_at`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_recover_codes` (
    `id` varchar(36) PRIMARY KEY,
    `user_id` varchar(36) NOT NULL,
    `code` varchar(256) NOT NULL,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `user_recover_codes_unique_with_use_id` UNIQUE (`user_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `namespaces` (
    `id` varchar(36) PRIMARY KEY,
    `name` varchar(64) NOT NULL,
    `description` varchar(256),
    `overview` BLOB,
    `visibility` varchar(64) NOT NULL DEFAULT 'private',
    `size_limit` bigint NOT NULL DEFAULT 0,
    `size` bigint NOT NULL DEFAULT 0,
    `size_dirty` tinyint(1) NOT NULL DEFAULT 0,
    `repository_limit` bigint NOT NULL DEFAULT 0,
    `repository_count` bigint NOT NULL DEFAULT 0,
    `tag_limit` bigint NOT NULL DEFAULT 0,
    `tag_count` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `namespaces_unique_with_name` UNIQUE (`name`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `audits` (
    `id` varchar(36) PRIMARY KEY,
    `user_id` varchar(36) NOT NULL,
    `username` varchar(128) NOT NULL DEFAULT '',
    `user_role` varchar(64) NOT NULL DEFAULT '',
    `namespace_id` varchar(36),
    `method` varchar(16) NOT NULL,
    `path` varchar(512) NOT NULL,
    `route` varchar(512) NOT NULL,
    `query` text,
    `status_code` integer NOT NULL,
    `client_ip` varchar(128) NOT NULL DEFAULT '',
    `user_agent` varchar(512) NOT NULL DEFAULT '',
    `latency_ms` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE INDEX `audits_idx_created_at` ON `audits` (`created_at`);

CREATE INDEX `audits_idx_user_created_at` ON `audits` (`user_id`, `created_at`);

CREATE TABLE IF NOT EXISTS `repositories` (
    `id` varchar(36) PRIMARY KEY,
    `namespace_id` varchar(36) NOT NULL,
    `name` varchar(64) NOT NULL,
    `description` varchar(255),
    `overview` BLOB,
    `size_limit` bigint NOT NULL DEFAULT 0,
    `size` bigint NOT NULL DEFAULT 0,
    `size_dirty` tinyint(1) NOT NULL DEFAULT 0,
    `tag_limit` bigint NOT NULL DEFAULT 0,
    `tag_count` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `repositories_unique_with_namespace` UNIQUE (
        `namespace_id`,
        `name`,
        `deleted_at`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE INDEX `repositories_idx_gc_deletable` ON `repositories` (`namespace_id`, `deleted_at`, `updated_at`, `id`);

CREATE TABLE IF NOT EXISTS `artifacts` (
    `id` varchar(36) PRIMARY KEY,
    `namespace_id` varchar(36) NOT NULL,
    `repository_id` varchar(36) NOT NULL,
    `digest` varchar(256) NOT NULL,
    `size` bigint NOT NULL DEFAULT 0,
    `blobs_size` bigint NOT NULL DEFAULT 0,
    `content_type` varchar(256) NOT NULL,
    `config_raw` MEDIUMBLOB,
    `config_media_type` varchar(256),
    `type` varchar(64) NOT NULL DEFAULT 'Unknown',
    `pushed_at` bigint NOT NULL,
    `last_pull` bigint NOT NULL,
    `pull_times` bigint NOT NULL DEFAULT 0,
    `referrer_id` varchar(36),
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `artifacts_unique_with_repo` UNIQUE (
        `repository_id`,
        `digest`,
        `deleted_at`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE INDEX `artifacts_idx_gc_deletable` ON `artifacts` (`repository_id`, `deleted_at`, `last_pull`, `updated_at`, `id`);
CREATE INDEX `artifacts_idx_referrer_deleted` ON `artifacts` (`referrer_id`, `deleted_at`);

CREATE TABLE IF NOT EXISTS `artifact_sboms` (
    `id` varchar(36) PRIMARY KEY,
    `artifact_id` varchar(36) NOT NULL,
    `raw` MEDIUMBLOB,
    `result` MEDIUMBLOB,
    `status` varchar(64) NOT NULL,
    `stdout` MEDIUMBLOB,
    `stderr` MEDIUMBLOB,
    `message` varchar(256),
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `artifact_sbom_unique_with_artifact` UNIQUE (`artifact_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `artifact_vulnerabilities` (
    `id` varchar(36) PRIMARY KEY,
    `artifact_id` varchar(36) NOT NULL,
    `version` bigint,
    `raw` MEDIUMBLOB,
    `result` MEDIUMBLOB,
    `status` varchar(64) NOT NULL,
    `stdout` MEDIUMBLOB,
    `stderr` MEDIUMBLOB,
    `message` varchar(256),
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `artifact_vulnerability_unique_with_artifact` UNIQUE (`artifact_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `tags` (
    `id` varchar(36) PRIMARY KEY,
    `repository_id` varchar(36) NOT NULL,
    `artifact_id` varchar(36) NOT NULL,
    `name` varchar(128) NOT NULL,
    `pushed_at` bigint NOT NULL,
    `last_pull` bigint,
    `pull_times` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `tags_unique_with_repo` UNIQUE (
        `repository_id`,
        `name`,
        `deleted_at`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE INDEX `tags_idx_gc_deletable` ON `tags` (`repository_id`, `deleted_at`, `updated_at`, `id`);
CREATE INDEX `tags_idx_artifact_deleted` ON `tags` (`artifact_id`, `deleted_at`);

CREATE TABLE IF NOT EXISTS `blobs` (
    `id` varchar(36) PRIMARY KEY,
    `digest` varchar(256) NOT NULL UNIQUE,
    `size` bigint NOT NULL,
    `content_type` varchar(256) NOT NULL,
    `pushed_at` bigint NOT NULL,
    `last_pull` bigint,
    `pull_times` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `blobs_unique_with_digest` UNIQUE (`digest`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE INDEX `blobs_idx_gc_deletable` ON `blobs` (`deleted_at`, `last_pull`, `updated_at`, `id`);

CREATE TABLE IF NOT EXISTS `blob_uploads` (
    `id` varchar(36) PRIMARY KEY,
    `part_number` int NOT NULL,
    `upload_id` varchar(256) NOT NULL,
    `etag` varchar(256) NOT NULL,
    `repository` varchar(256) NOT NULL,
    `file_id` varchar(256) NOT NULL,
    `size` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `blob_uploads_unique_with_upload_id_etag` UNIQUE (
        `upload_id`,
        `etag`,
        `deleted_at`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `artifact_artifacts` (
    `artifact_id` varchar(36) NOT NULL,
    `artifact_sub_id` varchar(36) NOT NULL,
    PRIMARY KEY (
        `artifact_id`,
        `artifact_sub_id`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE INDEX `artifact_artifacts_idx_sub_artifact` ON `artifact_artifacts` (`artifact_sub_id`, `artifact_id`);

CREATE TABLE IF NOT EXISTS `artifact_blobs` (
    `artifact_id` varchar(36) NOT NULL,
    `blob_id` varchar(36) NOT NULL,
    PRIMARY KEY (`artifact_id`, `blob_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE INDEX `artifact_blobs_idx_blob_artifact` ON `artifact_blobs` (`blob_id`, `artifact_id`);

CREATE TABLE IF NOT EXISTS `daemon_gc_tag_rules` (
    `id` varchar(36) PRIMARY KEY,
    `namespace_id` varchar(36),
    `is_running` tinyint NOT NULL DEFAULT 0,
    `cron_enabled` tinyint NOT NULL DEFAULT 0,
    `cron_rule` varchar(30),
    `cron_next_trigger` bigint,
    `retention_rule_type` varchar(64) NOT NULL DEFAULT 'Quantity',
    `retention_rule_amount` bigint NOT NULL DEFAULT 1,
    `retention_pattern` varchar(64),
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `daemon_gc_tag_rules_unique_with_ns` UNIQUE (`namespace_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_tag_runners` (
    `id` varchar(36) PRIMARY KEY,
    `rule_id` varchar(36) NOT NULL,
    `message` LONGBLOB,
    `status` varchar(64) NOT NULL DEFAULT 'Pending',
    `operate_type` varchar(64) NOT NULL DEFAULT 'Automatic',
    `operate_user_id` varchar(36),
    `started_at` bigint,
    `ended_at` bigint,
    `duration` bigint,
    `success_count` bigint,
    `failed_count` bigint,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_tag_records` (
    `id` varchar(36) PRIMARY KEY,
    `runner_id` varchar(36) NOT NULL,
    `tag` varchar(128) NOT NULL,
    `status` varchar(64) NOT NULL DEFAULT 'Success',
    `message` LONGBLOB,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_repository_rules` (
    `id` varchar(36) PRIMARY KEY,
    `namespace_id` varchar(36),
    `is_running` tinyint NOT NULL DEFAULT 0,
    `retention_day` int NOT NULL DEFAULT 0,
    `cron_enabled` tinyint NOT NULL DEFAULT 0,
    `cron_rule` varchar(30),
    `cron_next_trigger` bigint,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `daemon_gc_repository_rules_unique_with_ns` UNIQUE (`namespace_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_repository_runners` (
    `id` varchar(36) PRIMARY KEY,
    `rule_id` varchar(36) NOT NULL,
    `message` LONGBLOB,
    `status` varchar(64) NOT NULL DEFAULT 'Pending',
    `operate_type` varchar(64) NOT NULL DEFAULT 'Automatic',
    `operate_user_id` varchar(36),
    `started_at` bigint,
    `ended_at` bigint,
    `duration` bigint,
    `success_count` bigint,
    `failed_count` bigint,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_repository_records` (
    `id` varchar(36) PRIMARY KEY,
    `runner_id` varchar(36) NOT NULL,
    `repository` varchar(64) NOT NULL,
    `status` varchar(64) NOT NULL DEFAULT 'Success',
    `message` LONGBLOB,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_artifact_rules` (
    `id` varchar(36) PRIMARY KEY,
    `namespace_id` varchar(36),
    `is_running` tinyint NOT NULL DEFAULT 0,
    `retention_day` int NOT NULL DEFAULT 0,
    `cron_enabled` tinyint NOT NULL DEFAULT 0,
    `cron_rule` varchar(30),
    `cron_next_trigger` bigint,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `daemon_gc_artifact_rules_unique_with_ns` UNIQUE (`namespace_id`, `deleted_at`),
    KEY `daemon_gc_artifact_rules_idx_created_at` (`created_at`),
    KEY `daemon_gc_artifact_rules_idx_updated_at` (`updated_at`),
    KEY `daemon_gc_artifact_rules_idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_artifact_runners` (
    `id` varchar(36) PRIMARY KEY,
    `rule_id` varchar(36) NOT NULL,
    `message` LONGBLOB,
    `status` varchar(64) NOT NULL DEFAULT 'Pending',
    `operate_type` varchar(64) NOT NULL DEFAULT 'Automatic',
    `operate_user_id` varchar(36),
    `started_at` bigint,
    `ended_at` bigint,
    `duration` bigint,
    `success_count` bigint,
    `failed_count` bigint,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_artifact_records` (
    `id` varchar(36) PRIMARY KEY,
    `runner_id` varchar(36) NOT NULL,
    `digest` varchar(256) NOT NULL,
    `status` varchar(64) NOT NULL DEFAULT 'Success',
    `message` LONGBLOB,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_blob_rules` (
    `id` varchar(36) PRIMARY KEY,
    `is_running` tinyint NOT NULL DEFAULT 0,
    `retention_day` int NOT NULL DEFAULT 0,
    `cron_enabled` tinyint NOT NULL DEFAULT 0,
    `cron_rule` varchar(30),
    `cron_next_trigger` bigint,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_blob_runners` (
    `id` varchar(36) PRIMARY KEY,
    `rule_id` varchar(36) NOT NULL,
    `message` LONGBLOB,
    `status` varchar(64) NOT NULL DEFAULT 'Pending',
    `operate_type` varchar(64) NOT NULL DEFAULT 'Automatic',
    `operate_user_id` varchar(36),
    `started_at` bigint,
    `ended_at` bigint,
    `duration` bigint,
    `success_count` bigint,
    `failed_count` bigint,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `daemon_gc_blob_records` (
    `id` varchar(36) PRIMARY KEY,
    `runner_id` varchar(36) NOT NULL,
    `digest` varchar(256) NOT NULL,
    `status` varchar(64) NOT NULL DEFAULT 'Success',
    `message` LONGBLOB,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `gc_storage_deletion_tasks` (
    `id` varchar(36) PRIMARY KEY,
    `daemon` varchar(64) NOT NULL,
    `runner_id` varchar(36) NOT NULL,
    `resource_type` varchar(64) NOT NULL,
    `resource_id` varchar(36) NOT NULL,
    `storage_path` varchar(1024) NOT NULL,
    `storage_path_hash` varchar(64) NOT NULL,
    `status` varchar(64) NOT NULL DEFAULT 'Pending',
    `attempts` int NOT NULL DEFAULT 0,
    `message` LONGBLOB,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    UNIQUE KEY `gc_storage_deletion_tasks_unique_resource` (`resource_type`, `resource_id`, `storage_path_hash`, `deleted_at`),
    KEY `gc_storage_deletion_tasks_idx_status` (`status`, `deleted_at`, `updated_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `namespace_members` (
    `id` varchar(36) PRIMARY KEY,
    `user_id` varchar(36) NOT NULL,
    `namespace_id` varchar(36) NOT NULL,
    `role` varchar(64) NOT NULL DEFAULT 'NamespaceReader',
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `namespace_members_unique_with_user_ns_role` UNIQUE (
        `user_id`,
        `namespace_id`,
        `role`,
        `deleted_at`
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

INSERT INTO
    `namespaces` (
        `id`,
        `name`,
        `visibility`,
        `created_at`,
        `updated_at`
    )
VALUES (
        '019f56d8-633c-7d64-a115-f793aa8aad22',
        'library',
        'public',
        1775566388880,
        1775566388880
    );

CREATE TABLE IF NOT EXISTS `webhooks` (
    `id` varchar(36) PRIMARY KEY,
    `namespace_id` varchar(36),
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
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `webhook_logs` (
    `id` varchar(36) PRIMARY KEY,
    `webhook_id` varchar(36),
    `resource_type` varchar(64) NOT NULL,
    `action` varchar(64) NOT NULL,
    `status_code` smallint NOT NULL,
    `trace_context` BLOB,
    `req_header` BLOB NOT NULL,
    `req_body` BLOB NOT NULL,
    `resp_header` BLOB NOT NULL,
    `resp_body` BLOB,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `builders` (
    `id` varchar(36) PRIMARY KEY,
    `repository_id` varchar(36) NOT NULL,
    `source` varchar(64) NOT NULL,
    -- source SelfCodeRepository
    `scm_repository` varchar(256),
    `scm_credential_type` varchar(16),
    `scm_ssh_key` BLOB,
    `scm_token` varchar(256),
    `scm_username` varchar(30),
    `scm_password` varchar(30),
    -- source CodeRepository
    `code_repository_id` varchar(36),
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
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0,
    CONSTRAINT `builders_unique_with_repository` UNIQUE (`repository_id`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- TODO: buildx flags
CREATE TABLE IF NOT EXISTS `builder_runners` (
    `id` varchar(36) PRIMARY KEY,
    `builder_id` varchar(36) NOT NULL,
    `log` LONGBLOB,
    `status` varchar(64) NOT NULL DEFAULT 'Pending',
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
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `settings` (
    `id` varchar(36) PRIMARY KEY,
    `key` varchar(256) NOT NULL UNIQUE,
    `val` BLOB,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    `deleted_at` bigint NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_activity_hourlies` (
    `id` varchar(36) PRIMARY KEY,
    `hour` bigint NOT NULL,
    `user_id` varchar(36) NOT NULL,
    `push_count` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    CONSTRAINT `user_activity_hourlies_unique_with_hour_user` UNIQUE (`hour`, `user_id`),
    KEY `user_activity_hourlies_idx_user_hour` (`user_id`, `hour`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `namespace_activity_hourlies` (
    `id` varchar(36) PRIMARY KEY,
    `hour` bigint NOT NULL,
    `namespace_id` varchar(36) NOT NULL,
    `push_count` bigint NOT NULL DEFAULT 0,
    `pull_count` bigint NOT NULL DEFAULT 0,
    `size_delta` bigint NOT NULL DEFAULT 0,
    `tag_delta` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL,
    `updated_at` bigint NOT NULL,
    CONSTRAINT `namespace_activity_hourlies_unique_with_hour_namespace` UNIQUE (`hour`, `namespace_id`),
    KEY `namespace_activity_hourlies_idx_namespace_hour` (`namespace_id`, `hour`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `namespace_activity_hourlies`;

DROP TABLE IF EXISTS `user_activity_hourlies`;

DROP TABLE IF EXISTS `namespaces`;

DROP TABLE IF EXISTS `repositories`;

DROP TABLE IF EXISTS `artifacts`;

DROP TABLE IF EXISTS `tags`;

DROP TABLE IF EXISTS `blobs`;

DROP TABLE IF EXISTS `blob_uploads`;

DROP TABLE IF EXISTS `artifact_blobs`;

DROP TABLE IF EXISTS `users`;

DROP TABLE IF EXISTS `artifact_sbom`;

DROP TABLE IF EXISTS `artifact_vulnerability`;

DROP TABLE IF EXISTS `proxy_artifact_tasks`;

DROP TABLE IF EXISTS `proxy_artifact_blobs`;

DROP TABLE IF EXISTS `proxy_tag_tasks`;

DROP TABLE IF EXISTS `audits`;

DROP TABLE IF EXISTS `webhooks`;

DROP TABLE IF EXISTS `webhook_logs`;

DROP TABLE IF EXISTS `builders`;

DROP TABLE IF EXISTS `builder_logs`;

DROP TABLE IF EXISTS `user_3rdparty`;

DROP TABLE IF EXISTS `code_repository_clone_credentials`;

DROP TABLE IF EXISTS `code_repository_owners`;

DROP TABLE IF EXISTS `code_repositories`;

DROP TABLE IF EXISTS `user_recover_codes`;

DROP TABLE IF EXISTS `settings`;

DROP TABLE IF EXISTS `gc_storage_deletion_tasks`;

DROP TABLE IF EXISTS `daemon_gc_tag_rules`;

DROP TABLE IF EXISTS `daemon_gc_tag_runners`;

DROP TABLE IF EXISTS `daemon_gc_tag_records`;

DROP TABLE IF EXISTS `daemon_gc_repository_rules`;

DROP TABLE IF EXISTS `daemon_gc_repository_runners`;

DROP TABLE IF EXISTS `daemon_gc_repository_records`;

DROP TABLE IF EXISTS `daemon_gc_artifact_rules`;

DROP TABLE IF EXISTS `daemon_gc_artifact_runners`;

DROP TABLE IF EXISTS `daemon_gc_artifact_records`;

DROP TABLE IF EXISTS `daemon_gc_blob_rules`;

DROP TABLE IF EXISTS `daemon_gc_blob_runners`;

DROP TABLE IF EXISTS `daemon_gc_blob_runners`;
