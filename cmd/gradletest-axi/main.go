package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mrobertsatpolitico/gradletest-axi/internal/app"
	"github.com/mrobertsatpolitico/gradletest-axi/internal/buildinfo"
	"github.com/mrobertsatpolitico/gradletest-axi/internal/report"
)

var version string

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
		Version: buildinfo.Version(version),
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
	os.Exit(application.Execute(ctx, os.Args[1:], workingDirectory))
}
