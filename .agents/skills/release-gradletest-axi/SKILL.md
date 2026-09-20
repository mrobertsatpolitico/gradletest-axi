---
name: release-gradletest-axi
description: Prepare, publish, document, and verify a semantic version release of gradletest-axi. Use when asked to cut, tag, publish, or validate a gradletest-axi release.
compatibility: Requires git, authenticated GitHub CLI (gh), Go 1.27.1 or newer, Python 3, network access, and push permission for mrobertsatpolitico/gradletest-axi.
---

# Release gradletest-axi

Release only the canonical
`https://github.com/mrobertsatpolitico/gradletest-axi` repository. The
tag-triggered GitHub workflow owns GoReleaser publication; this skill prepares
the release commit, waits for CI, pushes the tag, synchronizes curated release
notes, and verifies the result.

## Required input

Obtain an explicit semantic version tag such as `v0.2.0`. If the user did not
provide one, review the unreleased changes, propose the next version, and wait
for confirmation before modifying files or creating commits.

The tag must be valid SemVer with a leading `v`. During `v0`, use a patch for
compatible fixes and a minor version for new or breaking behavior.

## Safety rules

- Never release from a branch other than `main`.
- Never include unrelated or pre-existing working-tree changes.
- Never tag a commit that has not passed the complete `ci` workflow.
- Never create or push a tag that already exists locally, remotely, or as a
  GitHub Release.
- Never move, replace, force-push, or delete a published tag.
- Never turn a failed release into different content under the same version.
  Fix the problem and prepare a new patch release.
- Stop immediately on a failed validation or workflow. Do not continue to the
  next release phase.

## 1. Preflight

1. Confirm the repository root, `origin` URL, `main` branch, authenticated
   `gh`, and push access.
2. Fetch `origin/main` and all tags.
3. Require an empty `git status --porcelain`.
4. Require local `HEAD` to equal both `main` and `origin/main`.
5. Validate the requested tag as SemVer.
6. Prove the tag is absent from local refs, remote refs, and GitHub Releases.
   Expected “not found” results are success for this preflight; any existing
   object is a hard stop.

Useful inspections:

```sh
git remote get-url origin
git branch --show-current
git fetch origin main --tags
git status --porcelain
git rev-parse HEAD
git rev-parse origin/main
gh auth status
gh repo view mrobertsatpolitico/gradletest-axi --json nameWithOwner,visibility,url
git tag --list "<tag>"
git ls-remote --tags origin "refs/tags/<tag>"
gh release view "<tag>" --repo mrobertsatpolitico/gradletest-axi
```

## 2. Compile the changelog

Find the most recent semantic version tag. For the first release, review all
history; otherwise review `<previous-tag>..HEAD`.

Read commit subjects and user-visible diffs, but write release notes for
people rather than copying the Git log. Update `CHANGELOG.md` as follows:

1. Keep a fresh `## [Unreleased]` section at the top.
2. Move its curated entries into `## [X.Y.Z] - YYYY-MM-DD`, omitting the
   leading `v` in the heading and using the current UTC date.
3. Group entries under only the applicable Keep a Changelog headings:
   `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, and `Security`.
4. Omit empty categories and internal-only implementation detail.
5. For the first release, link the version to
   `releases/tag/vX.Y.Z` and change `Unreleased` to compare
   `vX.Y.Z...HEAD`.
6. For later releases, link the version to
   `compare/<previous-tag>...vX.Y.Z` and update `Unreleased` to compare the
   new tag with `HEAD`.

Extract the prepared release notes and review them before committing:

```sh
python3 .agents/skills/release-gradletest-axi/scripts/extract_release_notes.py \
  CHANGELOG.md "<tag>"
```

## 3. Validate and publish the release commit

Run all local gates:

```sh
gofmt -d ./cmd ./internal
go mod tidy
go vet ./...
go test -race ./...
go build ./cmd/gradletest-axi
git diff --check
```

`go mod tidy` must not introduce an unplanned module change. Confirm that only
`CHANGELOG.md` changed during release preparation, then commit and push:

```sh
git add CHANGELOG.md
git diff --cached --name-only
git commit -m "docs: prepare <tag>"
git push origin main
```

Capture the release commit SHA. Find the `ci.yml` run for that exact SHA,
waiting briefly for GitHub to create it, then require success:

```sh
gh run list \
  --repo mrobertsatpolitico/gradletest-axi \
  --workflow ci.yml \
  --commit "<release-sha>" \
  --limit 1 \
  --json databaseId,headSha,status,conclusion
gh run watch "<run-id>" \
  --repo mrobertsatpolitico/gradletest-axi \
  --exit-status
```

Reconfirm that `origin/main` still points to the release commit before tagging.

## 4. Tag and publish

Create one annotated tag on the validated release commit and push only that
tag:

```sh
git tag -a "<tag>" "<release-sha>" -m "gradletest-axi <tag>"
git push origin "<tag>"
```

Find the `release.yml` run for the release commit and require success:

```sh
gh run list \
  --repo mrobertsatpolitico/gradletest-axi \
  --workflow release.yml \
  --commit "<release-sha>" \
  --limit 1 \
  --json databaseId,headSha,status,conclusion,url
gh run watch "<run-id>" \
  --repo mrobertsatpolitico/gradletest-axi \
  --exit-status
```

Do not create a second release manually. GoReleaser creates it from the tag.

## 5. Publish curated release notes

Extract the version body to a temporary file and replace GoReleaser's generated
commit list. This makes `CHANGELOG.md` authoritative:

```sh
python3 .agents/skills/release-gradletest-axi/scripts/extract_release_notes.py \
  CHANGELOG.md "<tag>" > "<notes-file>"
gh release edit "<tag>" \
  --repo mrobertsatpolitico/gradletest-axi \
  --verify-tag \
  --notes-file "<notes-file>"
```

Read the release body back and confirm it matches the extracted notes exactly.

## 6. Verify the release

Verify all of the following:

1. The local tag, remote tag, and GitHub Release point to the release commit.
2. A normal release is published and a pre-release tag is marked pre-release.
3. The release contains six platform archives and `checksums.txt`.
4. Every downloaded archive passes the checksum manifest.
5. The archive matching the current OS and architecture contains
   `gradletest-axi`, `README.md`, `CHANGELOG.md`, and the AXI documentation.
6. The packaged binary reports the exact tag through `--version`.
7. A clean public source install reports the exact tag:

```sh
GOBIN="<temporary-directory>" GOPROXY=direct \
  go install \
  "github.com/mrobertsatpolitico/gradletest-axi/cmd/gradletest-axi@<tag>"
"<temporary-directory>/gradletest-axi" --version
```

Use `gh release download <tag> --dir <temporary-directory>` to obtain all
assets. Verify checksums from inside that directory so manifest basenames
resolve correctly.

## 7. Report

Report the version, release commit, tag, GitHub Release URL, CI and release
workflow results, asset count, checksum result, packaged version, and public
`go install` result. Mention any verification that could not be completed.

If a transient workflow failure occurs after tagging, diagnose it and rerun
the same workflow only when no source or release content must change. Any
content change requires a new patch release.
