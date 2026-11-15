package authcontroller

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"uptime-app/backend/internal/config"
	"uptime-app/backend/internal/storage"
)

func TestSaveAvatar(t *testing.T) {
	var encoded bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.White)
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		data  []byte
		valid bool
	}{{"png", encoded.Bytes(), true}, {"invalid", []byte("not an image"), false}, {"truncated", encoded.Bytes()[:20], false}, {"oversized", make([]byte, maxAvatarSize+1), false}} {
		t.Run(tc.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			part, err := writer.CreateFormFile("avatar", "avatar.png")
			if err != nil {
				t.Fatal(err)
			}
			if _, err = part.Write(tc.data); err != nil {
				t.Fatal(err)
			}
			if err = writer.Close(); err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("POST", "/auth/avatar", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			dir := t.TempDir()
			name, err := saveAvatar(dir, req)
			if (err == nil) != tc.valid {
				t.Fatalf("name=%q err=%v", name, err)
			}
			if tc.valid {
				saved, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil || !bytes.Equal(saved, tc.data) {
					t.Fatalf("saved image: %v", err)
				}
			}
		})
	}
}

func TestUploadAvatarRequiresAuthentication(t *testing.T) {
	mux := http.NewServeMux()
	NewController(nil, config.Config{AvatarDir: t.TempDir()}).RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/auth/avatar", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestFileDownload(t *testing.T) {
	dir := t.TempDir()
	name, err := storage.Save(dir, bytes.NewReader([]byte("<script>test</script>")), 100)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewController(nil, config.Config{AvatarDir: t.TempDir(), FileDir: dir}).RegisterRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/files/"+name, nil))
	if response.Code != http.StatusOK || response.Body.String() != "<script>test</script>" {
		t.Fatalf("response: %d %q", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Disposition") != "attachment" || response.Header().Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("headers: %v", response.Header())
	}
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/auth/files", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("upload status: %d", response.Code)
	}
}
