package buildinfo

import "testing"

func TestResolveVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		linkedVersion string
		moduleVersion string
		want          string
	}{
		{
			name:          "release linker metadata wins",
			linkedVersion: "v1.2.3",
			moduleVersion: "v1.2.2",
			want:          "v1.2.3",
		},
		{
			name:          "go install module version",
			moduleVersion: "v0.4.0",
			want:          "v0.4.0",
		},
		{
			name:          "go install pseudo version",
			moduleVersion: "v0.0.0-20260920120000-abcdef123456",
			want:          "v0.0.0-20260920120000-abcdef123456",
		},
		{
			name:          "linked development value falls back to module",
			linkedVersion: developmentVersion,
			moduleVersion: "v0.3.0",
			want:          "v0.3.0",
		},
		{
			name:          "local build",
			moduleVersion: "(devel)",
			want:          developmentVersion,
		},
		{
			name: "missing build metadata",
			want: developmentVersion,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveVersion(test.linkedVersion, test.moduleVersion); got != test.want {
				t.Fatalf("resolveVersion(%q, %q) = %q, want %q", test.linkedVersion, test.moduleVersion, got, test.want)
			}
		})
	}
}
