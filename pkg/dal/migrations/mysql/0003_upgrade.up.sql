ALTER TABLE `artifacts`
  ADD COLUMN `namespace_id` bigint NOT NULL AFTER `id`;

ALTER TABLE `artifacts`
  ADD CONSTRAINT FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`);

