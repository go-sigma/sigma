CREATE TABLE IF NOT EXISTS `lockers` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `key` varchar(256) NOT NULL,
  `value` varchar(256) NOT NULL,
  `expire` integer NOT NULL DEFAULT 0,
  `created_at` integer NOT NULL DEFAULT (unixepoch () * 1000),
  `updated_at` integer NOT NULL DEFAULT (unixepoch () * 1000),
  `deleted_at` integer NOT NULL DEFAULT 0,
  CONSTRAINT `idx_lockers_key` UNIQUE (`key`, `deleted_at`)
);

