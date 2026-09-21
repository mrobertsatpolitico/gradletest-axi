# gradletest-axi

`gradletest-axi` runs one Gradle `Test`-compatible task in the nearest project
and emits one compact [TOON](https://github.com/toon-format/toon) document for
agents and automation. The task defaults to `test`; raw Gradle output never
appears on stdout.

## Install

### Go install

With Go 1.27.1 or newer, install the latest public source revision:

```sh
go install github.com/mrobertsatpolitico/gradletest-axi/cmd/gradletest-axi@latest
```

Use a semantic version after releases are available by replacing `latest`
with a tag such as `v1.0.0`.

The binary is written to `GOBIN`, or to the first `GOPATH` entry's `bin`
directory when `GOBIN` is unset. That directory must be on `PATH`.

### Prebuilt release

Tagged releases provide archives for macOS, Linux, and Windows on amd64 and
arm64 at the
[GitHub Releases page](https://github.com/mrobertsatpolitico/gradletest-axi/releases).
Each release also provides `checksums.txt`.

For example, download and verify the latest Apple Silicon archive with the
GitHub CLI:

```sh
gh release download \
  --repo mrobertsatpolitico/gradletest-axi \
  --pattern '*_darwin_arm64.tar.gz' \
  --pattern checksums.txt
shasum -a 256 --ignore-missing -c checksums.txt
tar -xzf gradletest-axi_*_darwin_arm64.tar.gz
mkdir -p ~/.local/bin
install -m 0755 gradletest-axi ~/.local/bin/gradletest-axi
```

Use `darwin_x86_64`, `linux_arm64`, or `linux_x86_64` for other Unix
targets. Windows releases use ZIP archives. The destination directory must be
on `PATH`.

### Local checkout

Install an unversioned development build from a checkout:

```sh
go install ./cmd/gradletest-axi
```

## Use

Run the default `test` task from anywhere beneath a project containing
`gradlew`:

```sh
gradletest-axi
```

Select another `Test`-compatible task with the optional positional parameter:

```sh
gradletest-axi integrationTest
gradletest-axi scraperTest
```

Alternate tasks may take longer when they start containers or other external
fixtures. They must accept the same Gradle test inputs and write Gradle JUnit
XML named `TEST-*.xml` beneath a project `build` directory.

Qualified task paths are also supported:

```sh
gradletest-axi :service:integrationTest
```

For a qualified path, the final component selects the conventional fallback
report directory. Pass arguments to the selected task after `--`:

```sh
gradletest-axi -- --tests com.example.WidgetTest
gradletest-axi integrationTest -- --tests com.example.WidgetIntegrationTest
```

Show every failure and complete JUnit failure text:

```sh
gradletest-axi integrationTest --full -- --tests com.example.WidgetIntegrationTest
```

Wrapper flags are `--full`, `--help`/`-h`, and `--version`/`-v`. Unknown
wrapper arguments, more than one task, and unsafe task paths are usage errors.
Wrapper flags may appear before or after the task.

## Output contract

Stdout contains exactly one TOON document after execution. Progress is written
to stderr. A successful result resembles:

```toon
status: passed
kind: test
task: test
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
task: integrationTest
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
results from being presented as the current failure. Reports created or
rewritten by the current invocation are discovered across project `build`
directories, including custom Gradle report locations. When a successful task
leaves reports untouched, discovery falls back to the selected task's
conventional `build/test-results/<task>/` directories.

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
state. It runs exactly one `Test`-compatible task per invocation.

## Development

```sh
gofmt -w ./cmd ./internal
go vet ./...
go test -race ./...
go build ./cmd/gradletest-axi
```

See [CHANGELOG.md](CHANGELOG.md) for curated user-facing changes and
[Releasing](docs/releases.md) for the tag-driven GitHub Release process.
