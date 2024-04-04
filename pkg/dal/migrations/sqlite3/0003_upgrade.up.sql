CREATE TABLE IF NOT EXISTS `lockers` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `name` varchar(256) NOT NULL,
  `value` varchar(256) NOT NULL,
  `expire` integer NOT NULL DEFAULT 0,
  `created_at` integer NOT NULL DEFAULT (unixepoch () * 1000),
  `updated_at` integer NOT NULL DEFAULT (unixepoch () * 1000),
  `deleted_at` integer NOT NULL DEFAULT 0
);

