# gradletest-axi

`gradletest-axi` runs the nearest Gradle project's fixed `test` task and emits
one compact [TOON](https://github.com/toon-format/toon) document for agents and
automation. Raw Gradle output never appears on stdout.

## Install

Build the local command:

```sh
go install ./cmd/gradletest-axi
```

Or build a versioned binary:

```sh
go build -ldflags "-X main.version=$(git describe --always --dirty)" -o bin/gradletest-axi ./cmd/gradletest-axi
```

## Use

Run all tests from anywhere beneath a project containing `gradlew`:

```sh
gradletest-axi
```

Pass arguments to Gradle's fixed `test` task after `--`:

```sh
gradletest-axi -- --tests com.example.WidgetTest
```

Show every failure and complete JUnit failure text:

```sh
gradletest-axi --full -- --tests com.example.WidgetTest
```

Wrapper flags are `--full`, `--help`/`-h`, and `--version`/`-v`. Unknown
wrapper arguments are usage errors.

## Output contract

Stdout contains exactly one TOON document after execution. Progress is written
to stderr. A successful result resembles:

```toon
status: passed
kind: test
exit_code: 0
gradle_exit: 0
duration_ms: 824
report: junit-current
tests:
  total: 18
  passed: 18
  failed: 0
  skipped: 0
  duration_ms: 411
```

A failing result with two JUnit-derived failures resembles:

```toon
status: failed
kind: test
exit_code: 1
gradle_exit: 1
duration_ms: 1096
report: junit-fresh
tests:
  total: 18
  passed: 15
  failed: 2
  skipped: 1
  duration_ms: 673
failures[2]{test,location,message}:
  com.example.WidgetTest.rejectsInvalidInput,"WidgetTest.kt:42",expected validation error
  com.example.WidgetTest.savesChanges,"WidgetTest.kt:87",expected true but was false
```

The default failure response includes at most five deterministically ordered
failures and truncates long messages. `--full` removes both limits. Missing,
malformed, or only partially readable JUnit XML is explicit; a failed Gradle
run only trusts reports that are new or changed since launch, preventing stale
results from being presented as the current failure.

The fields `exit_code` and `gradle_exit` make both the wrapper decision and the
child process result authoritative even when an outer execution harness cannot
report an exit status.

## Exit codes

- `0`: Gradle succeeded and no JUnit failures were reported.
- `1`: test, build, process, or wrapper failure.
- `2`: invalid wrapper usage.
- `128 + signal`: interrupted execution when the operating system exposes the
  child signal; otherwise interruption uses `130`.

## Cached logs

Combined Gradle stdout and stderr are saved beneath the operating system's user
cache directory in `gradletest-axi/<project-hash>/`. Files are created with
user-only permissions and the newest 20 logs per project are retained.

The TOON response includes a log path only when full diagnostics are useful:
build/process failures, unavailable or partial machine-readable results, or
truncated failure output. The tested repository is not modified by logging.

## Design

The command follows [AXI](https://github.com/kunchenguid/axi): content-first
bare invocation, deterministic structured output, bounded defaults, explicit
empty/error states, contextual recovery guidance, and full-fidelity data
available on demand. See [AXI principles in gradletest-axi](docs/axi-principles.md)
for the detailed public behavior behind each principle. The command
intentionally does not act as a general Gradle proxy and does not install
automatic session hooks, which could run expensive tests or surface stale
state.

## Development

```sh
gofmt -w ./cmd ./internal
go vet ./...
go test -race ./...
go build ./cmd/gradletest-axi
```
