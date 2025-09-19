```markdown
Title: test: reach 100% coverage for collector and cmd packages

Summary
-------
This branch adds tests, small refactors, and test-only overrides to bring the test coverage of
`internal/collector` and `cmd/experia-v10-exporter` to 100%.

What changed (high level)
- Add tests covering error branches in `authenticate()` and `fetchURL()`.
- Factor cookie-setting logic into `internal/collector/cookie_helper.go` and add tests for fallback behavior.
- Add `newRequest` indirection (production and test variants) to allow simulating `http.NewRequest` failures.
- Use `jsonMarshal` indirection in `authenticate()` so tests can simulate json.Marshal failures in both test and non-test builds.
- Make `main` testable by moving `exitOnError` to a separate file and adding test hooks.
- Clean up duplicate test definitions and centralize test helpers.

Why
---
Some error paths and edge cases were previously untested. The changes are minimal, low-risk refactors that make those branches testable without affecting production behaviour.

Files touched (representative)
- internal/collector/cookie_helper.go — helper to set cookies from an HTTP response (tested)
- internal/collector/newrequest_prod.go — production default for newRequest indirection
- internal/collector/test_newrequest_override.go — test build default for newRequest
- internal/collector/json_prod.go / test_overrides.go — jsonMarshal indirection (existing)
- internal/collector/metrics_test.go — non-test build tests exercising edge cases
- internal/collector/metrics_test_overrides_test.go — -tags test tests
- cmd/experia-v10-exporter/exitonerror_prod.go — move exitOnError for testability

Tests
-----
Run both test matrices locally to validate:

With test tag (includes test-only overrides):
  go test -tags test ./... -v

Without test tag (normal):
  go test ./... -v

Both runs should report 100% coverage.

How to create the PR locally (recommended)
----------------------------------------
I prepared this branch locally. To make a single squashed commit and open a PR from your machine, run:

  git checkout --orphan pr/tests-100-coverage
  git reset --hard
  git add -A
  git commit -m "chore(tests): add tests and small refactors to reach 100% coverage; see .github/PR_TESTS_100_COVERAGE.md for details"
  git push -u origin pr/tests-100-coverage

Then open a PR on GitHub comparing `pr/tests-100-coverage` -> `main` with the above title.

If you prefer I push or open the PR for you, grant push access or run the push command above from your environment.

GHCR (GitHub Container Registry) notes
-------------------------------------
This repository's CI now builds and pushes Docker images to GitHub Container Registry (GHCR) on pushes to `main`.

- The workflow uses the repository's `GITHUB_TOKEN` (no additional secrets required) and the repository owner's namespace `ghcr.io/<owner>/experia-v10-exporter`.
- If you want to push to a different account/org or use a PAT, create a secret (for example `GHCR_PAT`) with `write:packages` scope and update the workflow to use it instead of `GITHUB_TOKEN`.
- To publish images for pull requests as well, remove or change the job condition `if: github.event_name == 'push'` in `.github/workflows/ci.yml`.
- Tags created by CI: `:ci` and `:latest`. Adjust the `tags` field in the workflow to add additional tags (for example a commit SHA or Git tag).

Example: change the `tags` input in `.github/workflows/ci.yml` to include the short SHA:

  tags: |
    ghcr.io/${{ github.repository_owner }}/experia-v10-exporter:${{ github.sha }}
    ghcr.io/${{ github.repository_owner }}/experia-v10-exporter:latest

You can view pushed packages at https://github.com/orgs/<owner>/packages or https://ghcr.io/ depending on your account.

```
Title: test: reach 100% coverage for collector and cmd packages

Summary
-------
This branch adds tests, small refactors, and test-only overrides to bring the test coverage of
`internal/collector` and `cmd/experia-v10-exporter` to 100%.

What changed (high level)
- Add tests covering error branches in `authenticate()` and `fetchURL()`.
- Factor cookie-setting logic into `internal/collector/cookie_helper.go` and add tests for fallback behavior.
- Add `newRequest` indirection (production and test variants) to allow simulating `http.NewRequest` failures.
- Use `jsonMarshal` indirection in `authenticate()` so tests can simulate json.Marshal failures in both test and non-test builds.
- Make `main` testable by moving `exitOnError` to a separate file and adding test hooks.
- Clean up duplicate test definitions and centralize test helpers.

Why
---
Some error paths and edge cases were previously untested. The changes are minimal, low-risk refactors that make those branches testable without affecting production behaviour.

Files touched (representative)
- internal/collector/cookie_helper.go — helper to set cookies from an HTTP response (tested)
- internal/collector/newrequest_prod.go — production default for newRequest indirection
- internal/collector/test_newrequest_override.go — test build default for newRequest
- internal/collector/json_prod.go / test_overrides.go — jsonMarshal indirection (existing)
- internal/collector/metrics_test.go — non-test build tests exercising edge cases
- internal/collector/metrics_test_overrides_test.go — -tags test tests
- cmd/experia-v10-exporter/exitonerror_prod.go — move exitOnError for testability

Tests
-----
Run both test matrices locally to validate:

With test tag (includes test-only overrides):
  go test -tags test ./... -v

Without test tag (normal):
  go test ./... -v

Both runs should report 100% coverage.

How to create the PR locally (recommended)
----------------------------------------
I prepared this branch locally. To make a single squashed commit and open a PR from your machine, run:

  git checkout --orphan pr/tests-100-coverage
  git reset --hard
  git add -A
  git commit -m "chore(tests): add tests and small refactors to reach 100% coverage; see .github/PR_TESTS_100_COVERAGE.md for details"
  git push -u origin pr/tests-100-coverage

Then open a PR on GitHub comparing `pr/tests-100-coverage` -> `main` with the above title.

If you prefer I push or open the PR for you, grant push access or run the push command above from your environment.
