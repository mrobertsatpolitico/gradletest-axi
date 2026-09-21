# AXI principles in gradletest-axi

[`gradletest-axi`](../README.md) applies the ten principles of the
[Agent eXperience Interface (AXI)](https://github.com/kunchenguid/axi) to one
focused operation: running Gradle tests. The behaviors below describe the
public CLI contract rather than its internal implementation.

## 1. Token-efficient output

Every invocation writes one deterministic TOON document to stdout. TOON keeps
the result readable while using fewer tokens than an equivalent JSON response.
Repeated failure records use TOON's tabular form, with a test name, optional
source location, and message.

Gradle console output does not compete with the structured response. Progress
goes to stderr, while complete Gradle stdout and stderr are captured in a
private cached log. An agent can therefore consume stdout as data without
filtering banners, stack traces, download progress, or task logs.

## 2. Minimal default schemas

The top level contains only the fields needed to decide what happened:
`status`, `kind`, the selected `task`, the wrapper `exit_code`, the child
`gradle_exit` when available, elapsed time, and report provenance. Optional
sections appear only when they carry information for that result.

Test aggregates contain total, passed, failed, and skipped counts plus elapsed
test time. Each failure contains only its test name, message, and a location
when JUnit supplied a useful source frame. Warnings, log paths, truncation
markers, and help are omitted when they are not relevant.

## 3. Content truncation

The default response includes at most five failures. Individual failure
messages are bounded to 800 characters and include the original character
count when truncated, so an agent knows that more detail exists. In that case,
the response adds `truncated: true` and suggests rerunning with `--full`.

`--full` returns every JUnit failure and its complete failure text. Build and
configuration failures use a bounded actionable summary instead of dumping
Gradle's console tail. When machine-readable detail is incomplete or console
diagnostics are needed, the response supplies the cached raw-log path as the
full-fidelity escape hatch.

## 4. Pre-computed aggregates

The response reports the overall status and failure kind, wrapper and Gradle
exit values, command duration, and whether JUnit data is current, fresh,
partial, malformed, or unavailable. Agents do not need to infer success from
console text or trust an outer terminal's exit-code display.

When JUnit XML is available, `gradletest-axi` aggregates all discovered Gradle
test projects into total, passed, failed, and skipped counts plus cumulative
test duration. Failure and error elements are both counted as failed tests.
These summaries usually answer whether the run passed and how much work
remains without another command.

## 5. Definitive empty states

A valid JUnit run with no test cases reports explicit zero counts and suggests
checking the test filter if tests were expected. It never represents “zero
tests” as empty stdout.

Missing and unreadable result data are different states. `report: unavailable`
identifies an absent trusted report, `report: malformed` identifies reports
that could not be parsed, and a `-partial` provenance suffix identifies a mix
of usable and unusable reports. Discovery is limited to the selected task's
conventional `build/test-results/<task>/` directories. On a failed Gradle run,
only new or changed JUnit XML is trusted, so an old report cannot silently
become the current answer.

## 6. Structured errors & exit codes

Results and errors use TOON on stdout. Progress stays on stderr, and raw Gradle
output stays in the cache log. Wrapper errors translate the actionable problem
and provide relevant recovery guidance rather than leaking an unbounded child
process transcript.

Wrapper flags and the optional positional task are validated before Gradle
starts. Unknown flags, additional positionals, and unsafe task paths fail
loudly with exit code 2 and include valid usage; Gradle test arguments must be
placed after `--`. Execution is non-interactive and runs exactly one selected
task, defaulting to `test`.

Exit codes have a stable meaning:

- `0` means Gradle completed successfully and no JUnit failures were reported.
- `1` means a test, build, process, or wrapper operation failed.
- `2` means the wrapper invocation was invalid.
- Interrupted runs preserve a conventional signal exit when available and
  otherwise use `130`.

The TOON `exit_code` records the wrapper decision, while `gradle_exit` records
the child result whenever Gradle started. This remains authoritative even when
the surrounding terminal reports an unknown exit status.

## 7. Ambient context

Test execution is expensive and its result becomes stale as soon as relevant
code changes. For that reason, `gradletest-axi` deliberately does not install
a session-start hook, run tests automatically, or inject a cached pass/fail
result into every agent session.

Instead, the project ships an on-demand `gradletest-axi` Agent Skill. Compatible
agents can discover the command when a Gradle testing task arises, invoke it
at the point where a fresh result is useful, and interpret its status,
provenance, truncation, and log guidance consistently. This preserves AXI's
discovery goal without spending tokens and compute on unsolicited or stale
test context.

## 8. Content first

Running `gradletest-axi` with no arguments performs the useful operation rather
than printing a usage manual. It walks upward to the nearest executable Gradle
wrapper, runs the default `test` task with plain console mode, and returns the
structured result.

Agents can select another `Test`-compatible task with one positional argument,
including a qualified path such as `:service:integrationTest`, and narrow that
operation by passing Gradle test options after `--`. For example,
`gradletest-axi integrationTest -- --tests com.example.WidgetTest`. Alternate
tasks retain Gradle's conventional test input and JUnit output contract but
may take longer when they start containers. Help and version information
remain explicit requests rather than replacing the bare invocation's live test
result.

## 9. Contextual disclosure

Next steps appear when the current result creates a useful action:

- Truncated failures suggest the equivalent `--full` invocation.
- Build, process, missing-report, malformed-report, and partial-report results
  provide the cached log path for complete Gradle diagnostics.
- A zero-test result suggests checking the requested test filter.
- Project discovery, cache access, and wrapper startup errors include guidance
  specific to the failed prerequisite.
- Invalid wrapper input includes valid usage and an example of placing Gradle
  arguments after `--`.

A clean, complete pass does not repeat static advice. This keeps disclosure
specific to the result instead of charging every successful invocation for
irrelevant boilerplate.

## 10. Consistent way to get help

`gradletest-axi --help` and `gradletest-axi -h` return the same concise TOON
reference: the command's purpose, canonical usage, supported wrapper flags,
the optional task, the `--` separator, and examples for default, alternate,
and filtered runs. The tool has one focused operation and no subcommand tree,
so there is one predictable help surface.

Usage errors embed enough of that contract to self-correct without a separate
help call. `gradletest-axi --version` and `gradletest-axi -v` provide the
installed version through the same structured output channel.
