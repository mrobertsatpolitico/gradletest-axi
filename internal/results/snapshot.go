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

type Fingerprint struct {
	Digest           [sha256.Size]byte
	Size             int64
	ModifiedUnixNano int64
}

type Snapshot map[string]Fingerprint

func TakeSnapshot(root string) (Snapshot, error) {
	files, err := DiscoverFiles(root)
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

func DiscoverFiles(root string) ([]string, error) {
	return discoverFiles(root, func(buildDirectory string) ([]string, error) {
		var matches []string
		err := filepath.WalkDir(buildDirectory, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if path != buildDirectory && shouldSkipDirectory(entry.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasPrefix(entry.Name(), "TEST-") && strings.HasSuffix(entry.Name(), ".xml") {
				matches = append(matches, path)
			}
			return nil
		})
		return matches, err
	})
}

func DiscoverConventionalFiles(root, task string) ([]string, error) {
	reportTask := task
	if separator := strings.LastIndex(task, ":"); separator >= 0 {
		reportTask = task[separator+1:]
	}
	return discoverFiles(root, func(buildDirectory string) ([]string, error) {
		return filepath.Glob(filepath.Join(buildDirectory, "test-results", reportTask, "TEST-*.xml"))
	})
}

func discoverFiles(root string, find func(buildDirectory string) ([]string, error)) ([]string, error) {
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

		matches, findErr := find(path)
		if findErr != nil {
			return fmt.Errorf("find JUnit XML below %s: %w", path, findErr)
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
	info, err := file.Stat()
	if err != nil {
		return Fingerprint{}, fmt.Errorf("stat JUnit XML %s: %w", path, err)
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return Fingerprint{}, fmt.Errorf("fingerprint JUnit XML %s: %w", path, err)
	}
	fingerprint := Fingerprint{
		Size:             info.Size(),
		ModifiedUnixNano: info.ModTime().UnixNano(),
	}
	copy(fingerprint.Digest[:], hash.Sum(nil))
	return fingerprint, nil
}
