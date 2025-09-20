# GitHub Copilot Instructions for experia-v10-exporter

Concise guidance for AI agents editing this repository. Focus on discoverable, repo-specific patterns, build/test workflows, and integration points.

# Ensure for each run modulescruture.md is applied to all relevant files.
- check .github/instructions/modulestucture.md is applied to all internal/**/*.go and cmd/**/*.go files

## Quick facts
- Language: Go (go.mod: go 1.25.1)
# GitHub Copilot Instructions for experia-v10-exporter

This file gives concise, up-to-date guidance for AI assistants and contributors working in this repository. It highlights repo layout, testing patterns, CI expectations, and recent changes (test helpers, Docker layout, Dependabot, branch strategy).

## Quick facts
- Language: Go (go.mod: go 1.25.1)
- Module: github.com/GrammaTonic/experia-v10-exporter
- Main binary: `cmd/experia-v10-exporter`
- Core package: `internal/collector`

## What changed recently (important)
- Test helpers consolidated into `internal/testutil/testutil.go` — use these exported helpers (RoundTripperFunc, MakeJSONHandler, SampleMibJSON, SampleStatsFmt, SimpleErr, ErrReadCloser) instead of reimplementing stubs.
- Docker artifacts were moved into `docker/` (Dockerfile, docker-compose.yml). CI was updated to build using `docker/Dockerfile` (workflow uses a `CI/Run tests` check name for unit tests).
- Dependabot is enabled (`.github/dependabot.yml`) with weekly updates and auto-rebase. An automerge workflow enables GitHub auto-merge for Dependabot PRs once CI passes.
- Branch protection has been enabled for `main` and `develop` to require the `CI/Run tests` status and enforce for admins.
- Git branching model: `main <- develop <- feature/*`. `develop` exists and feature branches should target `develop`.

## Repo layout (quick)
- `cmd/experia-v10-exporter/` — main program
- `docker/` — Dockerfile(s) and docker-compose
- `.github/` — CI and automation (workflows, dependabot, copilot instructions)

## Test & CI guidance
- Run unit tests locally: `go test ./... -v` (Ginkgo is used in some packages: `ginkgo -v ./...`).
- Tests must not hit the network — inject `http.Client` or `Transport` or use `httptest.Server`. Reuse helpers in `internal/testutil`.
- CI exposes a status context named `CI/Run tests` — branch protection depends on this context.

## Dependabot & automation
- Dependabot is configured to update Go modules, GitHub Actions, and Docker image references weekly and to auto-rebase PRs.
- There's a workflow that enables auto-merge for Dependabot PRs once branch protections and CI pass. The workflow only enables auto-merge for Dependabot-authored PRs.

## Branching & release flow
- Branch strategy: work in `feature/*` branches off `develop`. Open PRs to `develop` for feature work. Merge `develop` to `main` for releases.
- Branch protection: `main` and `develop` require passing `CI/Run tests` and enforce for admins.

## Coding & style
- Use Go idioms: short names, error wrapping (`fmt.Errorf("...: %w", err)`), small interfaces for testing, and `context.Context` for cancellations/timeouts.
- Run `gofmt` and `go vet` before committing.

## When editing CI or Docker
- Update `.github/workflows/ci.yml` if you move Dockerfiles or change the build context — CI currently points to `docker/Dockerfile`.
- If you add Docker-related steps, ensure Docker Buildx step includes `file: docker/Dockerfile` when referencing the relocated file and keep multi-arch caching for GHCR if needed.

## What NOT to do
- Don’t perform real network I/O in unit tests. Don’t commit secrets or credentials; add them to GitHub Secrets if CI needs them.

## Quick checklist for PR authors
1. Branch from `develop` (feature/*). Run `go test ./... -v` locally.
2. Reuse helpers from `internal/testutil` for HTTP/mocks.
3. Push branch and open PR to `develop` (PR will run CI). Fix any test failures.
4. Dependabot PRs are auto-rebased and can be auto-merged once CI passes.

If you want this shortened or expanded, or to include repository-specific examples (PR templates, CODEOWNERS, release steps), say which sections to adjust.
- Project layout: `cmd/experia-v10-exporter` for the main binary, `internal/collector` for the collector package.
