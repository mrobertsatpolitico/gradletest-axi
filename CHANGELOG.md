# Changelog

All notable changes to this project will be documented in this file.

The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2026-09-21

### Added

- An optional positional Gradle `Test` task, including qualified task paths,
  with `test` preserved as the default.
- Task-aware JUnit discovery and a self-describing `task` field in execution
  results.

## [1.0.0] - 2026-09-20

### Added

- A fixed-operation Gradle `test` wrapper that discovers the nearest project
  wrapper and supports explicit test arguments after `--`.
- Deterministic TOON results with authoritative wrapper and Gradle exit
  values, aggregate test counts, bounded failure details, and `--full`.
- Multi-project JUnit XML aggregation with explicit zero, missing, malformed,
  and partial report states plus stale-report protection on failed builds.
- Non-interactive process supervision with signal forwarding, cached raw
  Gradle logs, bounded console summaries, and per-project log retention.
- Structured help, version, usage errors, contextual recovery guidance, and
  an on-demand Agent Skill for Gradle testing.
- AXI principle documentation, integration tests, race validation, and
  macOS/Linux continuous integration.
- Public `go install` support and tag-driven, checksummed release archives for
  macOS, Linux, and Windows on amd64 and arm64.

### Changed

- The CLI contract is declared stable for the first `v1.0.0` release.
- The development, CI, and source-install baseline now requires Go 1.27.1.

[Unreleased]: https://github.com/mrobertsatpolitico/gradletest-axi/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/mrobertsatpolitico/gradletest-axi/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/mrobertsatpolitico/gradletest-axi/releases/tag/v1.0.0
