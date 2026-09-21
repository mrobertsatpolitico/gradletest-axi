package results

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Fingerprint [sha256.Size]byte

type Snapshot map[string]Fingerprint

func TakeSnapshot(root, task string) (Snapshot, error) {
	files, err := DiscoverFiles(root, task)
	if err != nil {
		return nil, err
	}
	snapshot := make(Snapshot, len(files))
	for _, path := range files {
		fingerprint, fingerprintErr := fingerprintFile(path)
		if fingerprintErr != nil {
			return nil, fingerprintErr
		}
		snapshot[path] = fingerprint
	}
	return snapshot, nil
}

func DiscoverFiles(root, task string) ([]string, error) {
	reportTask := task
	if separator := strings.LastIndex(task, ":"); separator >= 0 {
		reportTask = task[separator+1:]
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if path != root && shouldSkipDirectory(entry.Name()) {
			return filepath.SkipDir
		}
		if entry.Name() != "build" {
			return nil
		}

		matches, globErr := filepath.Glob(filepath.Join(path, "test-results", reportTask, "TEST-*.xml"))
		if globErr != nil {
			return fmt.Errorf("find JUnit XML below %s: %w", path, globErr)
		}
		files = append(files, matches...)
		return filepath.SkipDir
	})
	if err != nil {
		return nil, fmt.Errorf("discover JUnit XML: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

func ChangedFiles(before Snapshot, after []string) ([]string, error) {
	var changed []string
	for _, path := range after {
		fingerprint, err := fingerprintFile(path)
		if err != nil {
			return nil, err
		}
		if previous, exists := before[path]; !exists || previous != fingerprint {
			changed = append(changed, path)
		}
	}
	return changed, nil
}

func shouldSkipDirectory(name string) bool {
	switch name {
	case ".git", ".gradle", ".idea", ".kotlin", "node_modules":
		return true
	default:
		return false
	}
}

func fingerprintFile(path string) (Fingerprint, error) {
	file, err := os.Open(path)
	if err != nil {
		return Fingerprint{}, fmt.Errorf("open JUnit XML %s: %w", path, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return Fingerprint{}, fmt.Errorf("fingerprint JUnit XML %s: %w", path, err)
	}
	var fingerprint Fingerprint
	copy(fingerprint[:], hash.Sum(nil))
	return fingerprint, nil
}
