package cli

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	DefaultTask = "test"
	Usage       = "gradletest-axi [task] [--full] [-- <task-args...>]"
)

type Options struct {
	Task        string
	Full        bool
	Help        bool
	Version     bool
	Passthrough []string
}

type UsageError struct {
	Message string
}

func (e *UsageError) Error() string {
	return e.Message
}

func Parse(args []string) (Options, error) {
	options := Options{Task: DefaultTask}
	taskSet := false
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch argument {
		case "--":
			options.Passthrough = append([]string(nil), args[index+1:]...)
			return options, nil
		case "--full":
			options.Full = true
		case "--help", "-h":
			options.Help = true
		case "--version", "-v":
			options.Version = true
		default:
			if strings.HasPrefix(argument, "-") {
				return Options{}, &UsageError{
					Message: fmt.Sprintf(
						"unknown argument %q; pass Gradle task arguments after `--`, for example `%s -- --tests ExampleTest`",
						argument,
						Usage,
					),
				}
			}
			if taskSet {
				return Options{}, &UsageError{
					Message: fmt.Sprintf("unexpected argument %q; select only one Gradle test task", argument),
				}
			}
			if err := validateTask(argument); err != nil {
				return Options{}, err
			}
			options.Task = argument
			taskSet = true
		}
	}
	return options, nil
}

func validateTask(task string) error {
	if task == "" || strings.ContainsAny(task, `/\*?[`) {
		return &UsageError{Message: fmt.Sprintf("invalid Gradle test task %q", task)}
	}
	for _, value := range task {
		if unicode.IsControl(value) || unicode.IsSpace(value) {
			return &UsageError{
				Message: fmt.Sprintf("invalid Gradle test task %q", task),
			}
		}
	}
	segments := strings.Split(task, ":")
	if segments[0] == "" {
		segments = segments[1:]
	}
	if len(segments) == 0 {
		return &UsageError{Message: fmt.Sprintf("invalid Gradle test task %q", task)}
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return &UsageError{Message: fmt.Sprintf("invalid Gradle test task %q", task)}
		}
	}
	return nil
}
