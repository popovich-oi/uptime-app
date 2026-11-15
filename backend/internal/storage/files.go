package storage

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var validName = regexp.MustCompile(`^[a-f0-9]{32}$`)
var ErrTooLarge = errors.New("file is too large")

// Save stores arbitrary file content under an opaque name, enforcing a size limit.
func Save(dir string, source io.Reader, limit int64) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	name := hex.EncodeToString(id)
	path := filepath.Join(dir, name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", err
	}
	n, copyErr := io.Copy(file, io.LimitReader(source, limit+1))
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil || n > limit {
		_ = os.Remove(path)
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		return "", ErrTooLarge
	}
	return name, nil
}

func Path(dir, name string) (string, bool) {
	if !validName.MatchString(name) {
		return "", false
	}
	return filepath.Join(dir, name), true
}
