CREATE TABLE IF NOT EXISTS `lockers` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `key` varchar(256) NOT NULL,
  `value` varchar(256) NOT NULL,
  `expire` bigint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `updated_at` bigint NOT NULL DEFAULT (UNIX_TIMESTAMP (CURRENT_TIMESTAMP()) * 1000),
  `deleted_at` bigint NOT NULL DEFAULT 0,
  CONSTRAINT `idx_lockers_key` UNIQUE (`key`, `deleted_at`)
);

