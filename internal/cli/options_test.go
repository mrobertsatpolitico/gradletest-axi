package cli

import (
	"errors"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want Options
	}{
		{name: "bare", want: Options{Task: DefaultTask}},
		{name: "full", args: []string{"--full"}, want: Options{Task: DefaultTask, Full: true}},
		{name: "help short", args: []string{"-h"}, want: Options{Task: DefaultTask, Help: true}},
		{name: "version", args: []string{"--version"}, want: Options{Task: DefaultTask, Version: true}},
		{name: "task", args: []string{"integrationTest"}, want: Options{Task: "integrationTest"}},
		{
			name: "task after flag",
			args: []string{"--full", "integrationTest"},
			want: Options{Task: "integrationTest", Full: true},
		},
		{
			name: "qualified task before flag",
			args: []string{":service:scraperTest", "--full"},
			want: Options{Task: ":service:scraperTest", Full: true},
		},
		{
			name: "passthrough",
			args: []string{"scraperTest", "--full", "--", "--tests", "com.example.WidgetTest"},
			want: Options{
				Task:        "scraperTest",
				Full:        true,
				Passthrough: []string{"--tests", "com.example.WidgetTest"},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(test.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Parse() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseRejectsInvalidArguments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown wrapper flag", args: []string{"--tests", "WidgetTest"}},
		{name: "multiple tasks", args: []string{"test", "integrationTest"}},
		{name: "empty qualified segment", args: []string{":service::integrationTest"}},
		{name: "path separator", args: []string{"service/integrationTest"}},
		{name: "glob metacharacter", args: []string{"integration[Test"}},
		{name: "parent directory", args: []string{".."}},
		{name: "whitespace", args: []string{"integration Test"}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(test.args)
			var usageError *UsageError
			if !errors.As(err, &usageError) {
				t.Fatalf("Parse() error = %v, want UsageError", err)
			}
		})
	}
}
