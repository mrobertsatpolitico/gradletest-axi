package cachelog

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreCreatesPrivateLogAndPrunesOldFiles(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	root := t.TempDir()
	times := []time.Time{
		time.Unix(1, 0),
		time.Unix(2, 0),
		time.Unix(3, 0),
	}
	index := 0
	store := Store{
		BaseDir: base,
		Keep:    2,
		Now: func() time.Time {
			value := times[index]
			index++
			return value
		},
	}
	for range times {
		file, _, err := store.Create(root)
		if err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Prune(root); err != nil {
		t.Fatal(err)
	}

	directory := filepath.Join(base, projectKey(root))
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("retained logs = %d, want 2", len(entries))
	}
	info, err := entries[0].Info()
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("log permissions = %o, want 600", info.Mode().Perm())
	}
}
