CREATE TABLE IF NOT EXISTS `auth_rules` (
  `id` varchar(26) PRIMARY KEY,
  `role` varchar(255) NOT NULL,
  `resource` varchar(255) NOT NULL,
  `action` varchar(255) NOT NULL,
  `effect` varchar(255) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `deleted_at` bigint DEFAULT NULL,
  PRIMARY KEY (id)) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `auth_roles` (
  `id` varchar(26) PRIMARY KEY,
  `role_id` varchar(26) NOT NULL,
  `scope_value` varchar(255) NOT NULL,
  `scope_type` varchar(255) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `deleted_at` bigint DEFAULT NULL,
  FOREIGN KEY (`role_id`) REFERENCES `auth_rules` (`id`)) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

