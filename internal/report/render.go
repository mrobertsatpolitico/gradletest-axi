package report

import (
	"fmt"
	"io"

	toon "github.com/toon-format/toon-go"
)

func Render(writer io.Writer, value any) error {
	ordered, err := toOrderedValue(value)
	if err != nil {
		return err
	}
	payload, err := toon.Marshal(ordered)
	if err != nil {
		return fmt.Errorf("encode TOON: %w", err)
	}
	if _, err := writer.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("write TOON: %w", err)
	}
	return nil
}

func toOrderedValue(value any) (any, error) {
	switch document := value.(type) {
	case Document:
		return documentObject(document), nil
	case ErrorDocument:
		fields := []toon.Field{
			{Key: "error", Value: document.Error},
			{Key: "exit_code", Value: document.ExitCode},
		}
		if len(document.Help) > 0 {
			fields = append(fields, toon.Field{Key: "help", Value: document.Help})
		}
		return toon.NewObject(fields...), nil
	case HelpDocument:
		flags := make([]toon.Object, 0, len(document.Flags))
		for _, flag := range document.Flags {
			flags = append(flags, toon.NewObject(
				toon.Field{Key: "flag", Value: flag.Flag},
				toon.Field{Key: "description", Value: flag.Description},
			))
		}
		return toon.NewObject(
			toon.Field{Key: "command", Value: document.Command},
			toon.Field{Key: "description", Value: document.Description},
			toon.Field{Key: "usage", Value: document.Usage},
			toon.Field{Key: "flags", Value: flags},
			toon.Field{Key: "help", Value: document.Help},
		), nil
	case VersionDocument:
		return toon.NewObject(
			toon.Field{Key: "name", Value: document.Name},
			toon.Field{Key: "version", Value: document.Version},
		), nil
	default:
		return nil, fmt.Errorf("unsupported report type %T", value)
	}
}

func documentObject(document Document) toon.Object {
	fields := []toon.Field{
		{Key: "status", Value: document.Status},
		{Key: "kind", Value: document.Kind},
		{Key: "exit_code", Value: document.ExitCode},
	}
	if document.GradleExit != nil {
		fields = append(fields, toon.Field{Key: "gradle_exit", Value: *document.GradleExit})
	}
	fields = append(fields,
		toon.Field{Key: "duration_ms", Value: document.Duration.Milliseconds()},
		toon.Field{Key: "report", Value: document.Report},
	)
	if document.Tests != nil {
		fields = append(fields, toon.Field{Key: "tests", Value: toon.NewObject(
			toon.Field{Key: "total", Value: document.Tests.Total},
			toon.Field{Key: "passed", Value: document.Tests.Passed},
			toon.Field{Key: "failed", Value: document.Tests.Failed},
			toon.Field{Key: "skipped", Value: document.Tests.Skipped},
			toon.Field{Key: "duration_ms", Value: document.Tests.DurationMS},
		)})
	}
	if len(document.Failures) > 0 {
		failures := make([]toon.Object, 0, len(document.Failures))
		for _, failure := range document.Failures {
			failureFields := []toon.Field{{Key: "test", Value: failure.Test}}
			if failure.Location != "" {
				failureFields = append(failureFields, toon.Field{Key: "location", Value: failure.Location})
			}
			failureFields = append(failureFields, toon.Field{Key: "message", Value: failure.Message})
			failures = append(failures, toon.NewObject(failureFields...))
		}
		fields = append(fields, toon.Field{Key: "failures", Value: failures})
	}
	if document.Error != "" {
		fields = append(fields, toon.Field{Key: "error", Value: document.Error})
	}
	if document.Warning != "" {
		fields = append(fields, toon.Field{Key: "warning", Value: document.Warning})
	}
	if document.Log != "" {
		fields = append(fields, toon.Field{Key: "log", Value: document.Log})
	}
	if document.Truncated {
		fields = append(fields, toon.Field{Key: "truncated", Value: true})
	}
	if len(document.Help) > 0 {
		fields = append(fields, toon.Field{Key: "help", Value: document.Help})
	}
	return toon.NewObject(fields...)
}
