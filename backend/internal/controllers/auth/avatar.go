package authcontroller

import (
	"bytes"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"uptime-app/backend/internal/storage"
)

const maxAvatarSize = 5 << 20

// saveAvatar validates the entire image and stores it under an unpredictable name.
func saveAvatar(dir string, r *http.Request) (string, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxAvatarSize+(1<<20))
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		return "", err
	}
	defer r.MultipartForm.RemoveAll()
	file, _, err := r.FormFile("avatar")
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxAvatarSize+1))
	if err != nil || len(data) > maxAvatarSize {
		return "", errors.New("image is too large")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") || cfg.Width > 4096 || cfg.Height > 4096 || cfg.Width <= 0 || cfg.Height <= 0 {
		return "", errors.New("invalid image")
	}
	// Decode verifies the contents, not just the image header.
	if _, _, err = image.Decode(bytes.NewReader(data)); err != nil {
		return "", err
	}
	name, err := storage.Save(dir, bytes.NewReader(data), maxAvatarSize)
	return name, err
}
