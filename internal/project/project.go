package project

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type GradleProject struct {
	Root    string
	Wrapper string
}

func Discover(start string) (GradleProject, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return GradleProject{}, fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		wrapperName := "gradlew"
		if runtime.GOOS == "windows" {
			wrapperName = "gradlew.bat"
		}
		wrapper := filepath.Join(current, wrapperName)
		if info, statErr := os.Stat(wrapper); statErr == nil && !info.IsDir() {
			if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
				return GradleProject{}, fmt.Errorf("Gradle wrapper is not executable: %s", wrapper)
			}
			return GradleProject{Root: current, Wrapper: wrapper}, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return GradleProject{}, fmt.Errorf("no Gradle wrapper found from %s or its parents", start)
		}
		current = parent
	}
}
