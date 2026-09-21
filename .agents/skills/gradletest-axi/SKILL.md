---
name: gradletest-axi
description: Run Gradle tests with compact, authoritative AXI output instead of raw Gradle console noise.
---

# Gradle test AXI

Use `gradletest-axi` instead of invoking Gradle `Test` tasks directly when this
command is available.

## Workflow

1. Run `gradletest-axi` from any directory beneath the Gradle project for the
   default `test` task, or pass one alternate task such as `integrationTest` or
   `scraperTest`.
2. Read the TOON `status`, `kind`, `task`, `exit_code`, `gradle_exit`, and test
   counts before acting.
3. Fix the bounded failures first. Treat their JUnit-derived locations and
   messages as current when `report` is `junit-current` or `junit-fresh`.
4. If `truncated: true`, rerun the relevant filter with `--full`.
5. Follow a returned `log` path only for diagnostics not represented by JUnit
   XML, such as compilation, configuration, process, malformed-report, or
   missing-report failures.

Pass test filters after the separator:

```sh
gradletest-axi -- --tests com.example.WidgetTest
gradletest-axi integrationTest -- --tests com.example.WidgetIntegrationTest
```

Qualified task paths such as `:service:integrationTest` are supported.
Alternate tasks must be `Test`-compatible and emit conventional JUnit XML;
they may run longer when starting test containers. Do not use this wrapper for
arbitrary non-test Gradle tasks.

Do not infer a test result from stale XML, raw log fragments, or an outer
terminal's unknown exit status. The response's `exit_code` is authoritative
for the wrapper and `gradle_exit` records the child result.
