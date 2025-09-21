# Release Guide for experia-v10-exporter

This document explains the steps and required repository secrets to perform a release using Goreleaser and GitHub Actions.

## Required repository secrets

- `GHCR_TOKEN` (recommended): a personal access token (PAT) that has the appropriate package permissions to push container images to GitHub Container Registry (GHCR). Scopes: `write:packages` and `read:packages` (or `packages:write` depending on GitHub's permission names). If you prefer to use `GITHUB_TOKEN` for release creation only, goreleaser can still create GitHub releases with the default `GITHUB_TOKEN` but pushing to GHCR may require a PAT.

- `GITHUB_TOKEN`: automatically provided in Actions; used by goreleaser to create GitHub Releases. No manual setup required.

## Tagging strategy

We follow semver tags. Tag format: `vMAJOR.MINOR.PATCH` (for example `v1.2.3`).

To create a release tag locally and push it:

```bash
# from the repository root on the branch you want to release (typically main)
git fetch --prune
git checkout main
git pull --ff-only origin main
# create tag
git tag -a v1.2.3 -m "Release v1.2.3"
# push tag
git push origin v1.2.3
```

Pushing the tag will trigger the `Release` workflow which runs goreleaser and publishes artifacts and Docker images (if `GHCR_TOKEN` is configured).

## Testing a release locally (snapshot)

You can test goreleaser locally without publishing by running the snapshot command:

```bash
# install goreleaser (if not already)
brew install goreleaser/tap/goreleaser || curl -sL https://git.io/goreleaser | bash

# run a snapshot locally
goreleaser --snapshot --rm-dist
```

This creates `dist/` artifacts and validates the config without publishing.

## Releasing from CI (GitHub Actions)

The repository contains `.github/workflows/release.yml` which:

- Runs a `snapshot` job for pull requests and pushes to `main` to validate the release process.
- Runs a `release` job on semver tag pushes that executes `goreleaser release --rm-dist`.

If you want to publish Docker images to GHCR, ensure `GHCR_TOKEN` is set in repository secrets. The workflow logs in to GHCR before running goreleaser.

## Post-release checks

After the release workflow completes, verify:

- A GitHub Release was created with the tag name and the expected changelog.
- `dist/` artifacts look correct (binaries and archives).
- Docker images are present in GHCR with the tag (if publishing enabled).

## Troubleshooting

- If goreleaser fails with schema/linter errors, run `goreleaser check` locally to get more detail.
- If Docker pushes fail due to permission errors, check that `GHCR_TOKEN` has the correct package permissions and belongs to a user with access to the repo or org.
- For runner caching issues, ensure your workflow cache keys are stable and avoid restoring archives on top of existing files that cause tar extraction conflicts.

## Rollback

If a release must be retracted, delete the GitHub Release and the Git tag, and optionally delete the GHCR package versions. Create a follow-up tag for the correct release.

## Contact

If you're unsure, ping the repository maintainer for guidance before performing production releases.
