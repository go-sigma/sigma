# Security Policy

## Supported Versions

Security fixes are provided for the active development branch and the latest
released version of sigma.

If you are using an older release, upgrade to the latest release before
reporting a vulnerability unless the issue is clearly present in the current
codebase.

## Reporting a Vulnerability

Please do not report security vulnerabilities through public GitHub issues,
pull requests, discussions, or chat channels.

Use GitHub's private vulnerability reporting feature for this repository when it
is available. If private reporting is not available, open a minimal public issue
requesting a private contact path from the maintainers, but do not include
exploit details, secrets, private deployment information, or proof-of-concept
payloads in the public issue.

When reporting a vulnerability, include as much of the following information as
you can safely share:

- Affected sigma version or commit hash
- Affected component, API, or deployment mode
- Clear steps to reproduce the issue
- Expected impact and security boundary crossed
- Relevant logs, configuration, or requests with secrets removed
- Whether the issue is already publicly known or actively exploited

## Response Process

The maintainers will review vulnerability reports as promptly as possible. After
triage, maintainers may request additional details, confirm whether the issue is
accepted, and coordinate a fix and disclosure timeline with the reporter.

Accepted vulnerabilities will be fixed in the active codebase and released as
appropriate for the severity and exploitability of the issue.

## Disclosure Guidelines

Please allow maintainers reasonable time to investigate and release a fix before
publicly disclosing a vulnerability.

Do not publish exploit details, proof-of-concept code, or operational indicators
until a fix or mitigation has been made available, unless maintainers explicitly
agree to another disclosure plan.

## Security Best Practices for Operators

- Change default credentials before exposing sigma to untrusted networks.
- Store database, Redis, and object storage credentials in a secret manager or
  Kubernetes Secret.
- Use HTTPS for registry traffic in production.
- Restrict access to administrative APIs.
- Keep sigma and its dependencies up to date.
- Review logs and audit records for suspicious activity.
