package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gradletest-axi/internal/app"
	"gradletest-axi/internal/report"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	workingDirectory, err := os.Getwd()
	if err != nil {
		_ = report.Render(os.Stdout, report.ErrorDocument{
			Error:    "resolve working directory: " + err.Error(),
			ExitCode: 1,
			Help:     []string{"Run the command from an accessible Gradle project directory"},
		})
		os.Exit(1)
	}

	application := app.Application{
		Version: version,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
	os.Exit(application.Execute(ctx, os.Args[1:], workingDirectory))
}
