package storage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndReadFiles(t *testing.T) {
	dir := t.TempDir()
	for _, data := range [][]byte{[]byte("text document"), {0, 1, 2, 255}} {
		name, err := Save(dir, bytes.NewReader(data), 100)
		if err != nil {
			t.Fatal(err)
		}
		path, ok := Path(dir, name)
		if !ok {
			t.Fatal("generated name rejected")
		}
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, data) {
			t.Fatalf("content %v, error %v", got, err)
		}
	}
	if _, ok := Path(dir, "../secret"); ok {
		t.Fatal("accepted path traversal")
	}
}

func TestSaveRejectsOversizedFileAndCleansUp(t *testing.T) {
	dir := t.TempDir()
	_, err := Save(dir, bytes.NewReader([]byte("12345")), 4)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("error: %v", err)
	}
	entries, err := os.ReadDir(filepath.Clean(dir))
	if err != nil || len(entries) != 0 {
		t.Fatalf("entries: %v, error: %v", entries, err)
	}
}
