# Contributing to sigma

Thank you for your interest in contributing to sigma. This guide explains how
to prepare changes, run validation, and submit contributions in a way that keeps
the project maintainable.

## Code of Conduct

All contributors are expected to follow the
[Code of Conduct](CODE_OF_CONDUCT.md). Be respectful, constructive, and focused
on improving the project.

## Getting Started

1. Fork the repository and create a topic branch from the active development
   branch.
2. Install the Go toolchain required by `go.mod`.
3. Use the sample configuration under `conf/` when running sigma locally.
4. Keep changes focused. Separate unrelated fixes, refactors, and features into
   different pull requests.

To run sigma locally:

```bash
go run . server -c conf/config.yaml
```

To build the binary:

```bash
make build
```

## Development Guidelines

- Follow the existing package structure and local patterns.
- Keep data access logic in the DAL layer and business orchestration in the
  service layer.
- Use dependency injection through `go.uber.org/dig` for application wiring.
- Use `log/slog` for logging.
- Pass `context.Context` through request and background task call chains.
- Generate application IDs with `pkg/utils/uuid.NewV7String()`.
- Avoid committing generated files unless the change requires regeneration.
- Do not commit secrets, credentials, local databases, or environment-specific
  configuration.

## Testing and Linting

The project uses build tags for tests, linting, and builds:

```text
netgo,timetzdata,exclude_graphdriver_btrfs,containers_image_openpgp
```

Run Go lint checks:

```bash
make lint-go
```

Run the full lint target:

```bash
make lint
```

Run tests with sqlite:

```bash
CI_DATABASE_TYPE=sqlite3 go test -parallel 1 -failfast \
  -tags "netgo,timetzdata,exclude_graphdriver_btrfs,containers_image_openpgp" \
  -timeout 30m ./...
```

When a change only affects a small package, it is fine to run a focused test
first. Before submitting a pull request, run the broader validation that matches
the change's risk.

## Generated Code

Some files are generated and should not be edited by hand.

- Regenerate GORM query code with `make gormgen`.
- Regenerate Swagger documentation with `make swagen`.
- Regenerate mocks with the existing `go:generate` directives when interfaces
  change.

If generated output changes, include it in the same pull request as the source
change that required it.

## Commit Messages

Commit messages must follow Conventional Commits:

```text
<type>(<scope>): <subject>
```

Examples:

```text
feat(distribution): add manifest referrer filtering
fix(auth): handle expired bearer tokens
refactor(repository): move creation orchestration to service layer
docs: add contributing guidelines
```

Use an English subject, keep it concise, and avoid vague messages such as
`update code` or `fix bug`. Add a body when the change needs context, and include
validation commands under `Verification:` only when you actually ran them.

## Pull Requests

Before opening a pull request:

- Rebase or merge the latest target branch.
- Confirm the worktree does not contain unrelated local files.
- Run formatting, linting, and tests appropriate for the change.
- Update documentation, examples, configuration, Helm values, or migrations when
  behavior changes.
- Explain what changed, why it changed, and how it was verified.

Pull requests should be small enough to review effectively. Large changes are
easier to merge when split into preparatory refactors, behavior changes, and
follow-up cleanup.

## Reporting Issues

When reporting a bug, include:

- The sigma version or commit hash.
- The deployment mode and storage/database/cache backends in use.
- Steps to reproduce the issue.
- Expected behavior and actual behavior.
- Relevant logs or error messages, with secrets removed.

For feature requests, describe the use case, the current limitation, and any
operational constraints that affect the design.
