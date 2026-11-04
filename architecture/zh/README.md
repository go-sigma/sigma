# Architecture（中文）

本目录存放跨模块设计文档的中文版本；英文版本在上一级目录 `architecture/`。

## 目录结构

- `architecture/*.md`：英文设计文档。
- `architecture/zh/*.md`：中文设计文档；与英文版使用相同的 `<topic>` 文件名。

## 新增设计

1. 在 `architecture/` 新增 `<topic>.md`（英文），在 `architecture/zh/` 新增同名文件（中文）。
2. 在 [`architecture/README.md`](../README.md) 和本文件的索引表里各登记一行。
3. 实施进度以 PR / 代码为准；文档描述目标方案和已经确定的决策。

## 索引

| 文档 | 语言 | 状态 | 最后更新 | 范围 |
| --- | --- | --- | --- | --- |
| [size-quota.md](./size-quota.md) | 中文 | 已定稿 | 2026-11-04 | namespace / repository 的 size 与 tag/repository 数量配额、全局 `storage.capacity` |
