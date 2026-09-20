package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mrobertsatpolitico/gradletest-axi/internal/cachelog"
	"github.com/mrobertsatpolitico/gradletest-axi/internal/cli"
	"github.com/mrobertsatpolitico/gradletest-axi/internal/project"
	"github.com/mrobertsatpolitico/gradletest-axi/internal/report"
	"github.com/mrobertsatpolitico/gradletest-axi/internal/results"
	"github.com/mrobertsatpolitico/gradletest-axi/internal/runner"
)

const commandName = "gradletest-axi"

type Application struct {
	Version   string
	Stdout    io.Writer
	Stderr    io.Writer
	CacheBase string
}

func (a Application) Execute(ctx context.Context, args []string, workingDirectory string) int {
	options, err := cli.Parse(args)
	if err != nil {
		return a.renderError(err.Error(), 2, []string{"Usage: " + cli.Usage})
	}
	if options.Help {
		return a.renderHelp()
	}
	if options.Version {
		version := a.Version
		if version == "" {
			version = "dev"
		}
		if err := report.Render(a.stdout(), report.VersionDocument{Name: commandName, Version: version}); err != nil {
			return a.renderFallback(err)
		}
		return 0
	}

	gradleProject, err := project.Discover(workingDirectory)
	if err != nil {
		return a.renderError(err.Error(), 1, []string{
			"Run `gradletest-axi` inside a Gradle project containing an executable `gradlew`",
		})
	}

	store, err := a.cacheStore()
	if err != nil {
		return a.renderError(err.Error(), 1, []string{"Ensure the user cache directory is writable"})
	}
	logFile, logPath, err := store.Create(gradleProject.Root)
	if err != nil {
		return a.renderError(err.Error(), 1, []string{"Ensure the user cache directory is writable"})
	}

	before, snapshotErr := results.TakeSnapshot(gradleProject.Root)
	fmt.Fprintln(a.stderr(), "gradletest-axi: running Gradle tests")
	runResult := runner.Run(ctx, runner.Command{
		Executable: gradleProject.Wrapper,
		Args:       append([]string{"test", "--console=plain"}, options.Passthrough...),
		Directory:  gradleProject.Root,
	}, logFile)
	closeErr := logFile.Close()
	pruneErr := store.Prune(gradleProject.Root)

	document, exitCode := buildDocument(buildInput{
		Run:           runResult,
		Full:          options.Full,
		Root:          gradleProject.Root,
		LogPath:       logPath,
		Snapshot:      before,
		SnapshotError: snapshotErr,
		CloseError:    closeErr,
		PruneError:    pruneErr,
	})
	if err := report.Render(a.stdout(), document); err != nil {
		return a.renderFallback(err)
	}
	return exitCode
}

type buildInput struct {
	Run           runner.Result
	Full          bool
	Root          string
	LogPath       string
	Snapshot      results.Snapshot
	SnapshotError error
	CloseError    error
	PruneError    error
}

func buildDocument(input buildInput) (report.Document, int) {
	document := report.Document{
		Duration: input.Run.Duration,
		Report:   "unavailable",
	}
	if input.Run.HasExitCode {
		gradleExit := input.Run.ExitCode
		document.GradleExit = &gradleExit
	}

	if input.Run.StartError != nil {
		document.Status = "failed"
		document.Kind = "wrapper"
		document.ExitCode = 1
		document.Error = "could not start the Gradle wrapper: " + input.Run.StartError.Error()
		document.Log = input.LogPath
		document.Help = []string{"Verify the Gradle wrapper is executable and the configured Java runtime is available"}
		return document, 1
	}

	files, discoveryErr := results.DiscoverFiles(input.Root)
	selectedFiles := files
	provenance := "current"
	if input.Run.ExitCode != 0 {
		provenance = "fresh"
		if input.SnapshotError != nil {
			selectedFiles = nil
		} else if discoveryErr == nil {
			selectedFiles, discoveryErr = results.ChangedFiles(input.Snapshot, files)
		}
	}

	aggregate := results.Aggregate{}
	if discoveryErr == nil {
		aggregate = results.ParseFiles(selectedFiles)
	}
	validFiles := aggregate.Files - len(aggregate.ParseErrors)
	if validFiles > 0 {
		document.Report = "junit-" + provenance
		if len(aggregate.ParseErrors) > 0 {
			document.Report += "-partial"
		}
		document.Tests = report.CountsFromAggregate(aggregate)
		document.Failures, document.Truncated = report.FailuresFromAggregate(aggregate, input.Full)
	} else if aggregate.Files > 0 {
		document.Report = "malformed"
	}

	warnings := collectWarnings(input, discoveryErr, aggregate)
	if len(warnings) > 0 {
		document.Warning = strings.Join(warnings, "; ")
	}

	switch {
	case input.Run.Interrupted:
		document.Status = "interrupted"
		document.Kind = "process"
		document.ExitCode = input.Run.ExitCode
		if document.ExitCode < 128 {
			document.ExitCode = 130
		}
		document.Error = "Gradle test execution was interrupted"
	case aggregate.Failed > 0:
		document.Status = "failed"
		document.Kind = "test"
		document.ExitCode = 1
	case input.Run.ExitCode != 0:
		document.Status = "failed"
		document.Kind = "build"
		document.ExitCode = 1
		document.Error = report.SummarizeGradleFailure(input.Run.Tail)
	default:
		document.Status = "passed"
		document.Kind = "test"
		document.ExitCode = 0
	}

	needsLog := document.Kind == "build" || document.Kind == "wrapper" || document.Kind == "process" ||
		document.Warning != "" || document.Truncated
	if needsLog {
		document.Log = input.LogPath
	}
	if document.Truncated {
		document.Help = append(document.Help, "Run `gradletest-axi --full` with the same test filter to see all failure details")
	}
	if needsLog {
		document.Help = append(document.Help, fmt.Sprintf("Inspect `%s` for complete Gradle output", input.LogPath))
	}
	if document.Tests != nil && document.Tests.Total == 0 {
		document.Help = append(document.Help, "0 tests were reported; verify the test filter if tests were expected")
	}
	return document, document.ExitCode
}

func collectWarnings(input buildInput, discoveryErr error, aggregate results.Aggregate) []string {
	var warnings []string
	if input.SnapshotError != nil && input.Run.ExitCode != 0 {
		warnings = append(warnings, "could not snapshot existing JUnit reports before execution")
	}
	if discoveryErr != nil {
		warnings = append(warnings, "could not discover JUnit XML reports")
	} else if aggregate.Files == 0 {
		warnings = append(warnings, "no JUnit XML reports were available")
	} else if len(aggregate.ParseErrors) > 0 {
		warnings = append(warnings, fmt.Sprintf("%d JUnit XML report(s) could not be parsed", len(aggregate.ParseErrors)))
	}
	if input.CloseError != nil {
		warnings = append(warnings, "the Gradle log could not be closed cleanly")
	}
	if input.PruneError != nil {
		warnings = append(warnings, "old cached logs could not be pruned")
	}
	if input.Run.TerminateError != nil && !errors.Is(input.Run.TerminateError, os.ErrProcessDone) {
		warnings = append(warnings, "the Gradle process did not terminate cleanly")
	}
	return warnings
}

func (a Application) cacheStore() (cachelog.Store, error) {
	if a.CacheBase != "" {
		return cachelog.Store{BaseDir: a.CacheBase, Keep: 20}, nil
	}
	return cachelog.Default()
}

func (a Application) renderHelp() int {
	document := report.HelpDocument{
		Command:     commandName,
		Description: "Run the nearest Gradle project's test task and return a compact AXI result",
		Usage:       cli.Usage,
		Flags: []report.Flag{
			{Flag: "--full", Description: "include every failure and complete JUnit failure text"},
			{Flag: "--help, -h", Description: "show this concise reference"},
			{Flag: "--version, -v", Description: "show the installed version"},
			{Flag: "--", Description: "pass remaining arguments to the fixed Gradle test task"},
		},
		Help: []string{
			"Run `gradletest-axi` to execute tests",
			"Run `gradletest-axi -- --tests ExampleTest` to filter tests",
		},
	}
	if err := report.Render(a.stdout(), document); err != nil {
		return a.renderFallback(err)
	}
	return 0
}

func (a Application) renderError(message string, exitCode int, help []string) int {
	if err := report.Render(a.stdout(), report.ErrorDocument{
		Error:    message,
		ExitCode: exitCode,
		Help:     help,
	}); err != nil {
		return a.renderFallback(err)
	}
	return exitCode
}

func (a Application) renderFallback(err error) int {
	fmt.Fprintf(a.stdout(), "error: %s\nexit_code: 1\n", strings.ReplaceAll(err.Error(), "\n", " "))
	return 1
}

func (a Application) stdout() io.Writer {
	if a.Stdout != nil {
		return a.Stdout
	}
	return io.Discard
}

func (a Application) stderr() io.Writer {
	if a.Stderr != nil {
		return a.Stderr
	}
	return io.Discard
}
