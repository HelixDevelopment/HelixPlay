# Contributing to HelixPlay

Thank you for your interest in HelixPlay! This document outlines the standards and processes for contributing to the project.

## Prime Directive (Non-Negotiable)

> "Execution of tests and Challenges MUST guarantee the quality, the completion and full usability by end users of the product. We had been in position that all tests do execute with success and all Challenges as well, but in reality the most of the features does not work and can't be used! This MUST NOT be the case."

Any contribution that allows green tests on broken features is a violation and will be rejected.

## Branch Strategy

- `main` — stable, CI-passing, release-ready
- `001-helixplay-system` — current feature branch for HelixPlay system implementation
- `feature/<name>` — short-lived feature branches
- `hotfix/<name>` — emergency fixes

## Commit Conventions

Use conventional commits:
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `test`, `refactor`, `docs`, `chore`, `perf`, `security`

Scopes: `host-agent`, `core`, `client-wails`, `client-web`, `client-flutter`, `<submodule-name>`

## Pull Request Template

```markdown
## Summary
Brief description of the change.

## Type
- [ ] Feature
- [ ] Bug fix
- [ ] Test
- [ ] Refactor
- [ ] Documentation

## Constitution Compliance
- [ ] Anti-bluff scan passes (`make anti-bluff`)
- [ ] Unit tests ≥95% coverage (`make test-unit`)
- [ ] No TODO/FIXME/placeholders
- [ ] Tests verify observable behavior (≥60% observable assertions)
- [ ] Mocks only in Unit tests
- [ ] Documentation updated

## Testing
How was this tested? Include test commands and results.

## Breaking Changes
List any breaking changes and migration steps.
```

## Code Review Requirements

- All PRs require at least 2 approvals
- CI must pass (including anti-bluff scan)
- Security scan must pass (no CRITICAL/HIGH vulns)
- Performance regression check (if applicable)

## Anti-Bluff Pledge

Before submitting a PR, confirm:
- [ ] My tests would fail if the feature were broken
- [ ] My tests exercise real behavior, not just constructors
- [ ] Integration/E2E tests use real dependencies, not mocks
- [ ] I have run `make anti-bluff` and it passes
- [ ] I have run `make test` and all tests pass

## Contact

- Discord: [HelixPlay Dev](https://discord.gg/helixplay)
- Matrix: #helixplay-dev:matrix.org
