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
		{name: "bare", want: Options{}},
		{name: "full", args: []string{"--full"}, want: Options{Full: true}},
		{name: "help short", args: []string{"-h"}, want: Options{Help: true}},
		{name: "version", args: []string{"--version"}, want: Options{Version: true}},
		{
			name: "passthrough",
			args: []string{"--full", "--", "--tests", "com.example.WidgetTest"},
			want: Options{Full: true, Passthrough: []string{"--tests", "com.example.WidgetTest"}},
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

func TestParseRejectsUnknownWrapperArgument(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"--tests", "WidgetTest"})
	var usageError *UsageError
	if !errors.As(err, &usageError) {
		t.Fatalf("Parse() error = %v, want UsageError", err)
	}
}
