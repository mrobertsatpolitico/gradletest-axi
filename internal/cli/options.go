package cli

import (
	"fmt"
	"strings"
)

const Usage = "gradletest-axi [--full] [-- <gradle-test-args...>]"

type Options struct {
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
	var options Options
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
			hint := ""
			if strings.HasPrefix(argument, "-") {
				hint = fmt.Sprintf("; pass Gradle test arguments after `--`, for example `%s -- --tests ExampleTest`", Usage)
			}
			return Options{}, &UsageError{
				Message: fmt.Sprintf("unknown argument %q%s", argument, hint),
			}
		}
	}
	return options, nil
}
