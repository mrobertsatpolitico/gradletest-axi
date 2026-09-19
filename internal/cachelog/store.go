package cachelog

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const defaultKeep = 20

type Store struct {
	BaseDir string
	Keep    int
	Now     func() time.Time
}

func Default() (Store, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return Store{}, fmt.Errorf("resolve user cache directory: %w", err)
	}
	return Store{
		BaseDir: filepath.Join(cacheDir, "gradletest-axi"),
		Keep:    defaultKeep,
		Now:     time.Now,
	}, nil
}

func (s Store) Create(projectRoot string) (*os.File, string, error) {
	directory := filepath.Join(s.BaseDir, projectKey(projectRoot))
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, "", fmt.Errorf("create log directory: %w", err)
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	name := fmt.Sprintf("%s-%d.log", now().UTC().Format("20060102T150405.000000000Z"), os.Getpid())
	path := filepath.Join(directory, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, "", fmt.Errorf("create Gradle log: %w", err)
	}
	return file, path, nil
}

func (s Store) Prune(projectRoot string) error {
	keep := s.Keep
	if keep <= 0 {
		keep = defaultKeep
	}
	directory := filepath.Join(s.BaseDir, projectKey(projectRoot))
	entries, err := os.ReadDir(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read log directory: %w", err)
	}

	files := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".log" {
			files = append(files, entry)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})
	if len(files) <= keep {
		return nil
	}
	for _, entry := range files[keep:] {
		if err := os.Remove(filepath.Join(directory, entry.Name())); err != nil {
			return fmt.Errorf("remove old Gradle log: %w", err)
		}
	}
	return nil
}

func projectKey(root string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(root)))
	return hex.EncodeToString(sum[:8])
}
