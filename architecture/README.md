# Architecture

Cross-module design documents.

## Layout

- `architecture/*.md` — English design documents.
- `architecture/zh/*.md` — Chinese design documents for the same topics.

## Adding a design

1. Add `<topic>.md` in `architecture/` (English) and/or `architecture/zh/<topic>.md` (Chinese).
   Keep the same `<topic>` name in both languages.
2. Register it in the index tables below and in [`zh/README.md`](./zh/README.md).
3. Implementation status is tracked by PRs / code; a document describes the target design and
   the decisions that have been made.

## Index

| Document | Language | Status | Last updated | Scope |
| --- | --- | --- | --- | --- |
| [size-quota.md](./size-quota.md) | English | Final | 2026-11-04 | namespace / repository size and tag/repository count quotas, global `storage.capacity` |
| [zh/size-quota.md](./zh/size-quota.md) | Chinese | Final | 2026-11-04 | namespace / repository size and tag/repository count quotas, global `storage.capacity` |
