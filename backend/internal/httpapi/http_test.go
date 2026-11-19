package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORS(t *testing.T) {
	const allowed = "http://localhost:3000"
	for _, tc := range []struct {
		name, origin, method, requestMethod, wantOrigin string
		wantStatus                                      int
		wantCalled                                      bool
	}{
		{"allowed request", allowed, "GET", "", allowed, http.StatusTeapot, true},
		{"unrelated origin", "https://example.com", "GET", "", "", http.StatusTeapot, true},
		{"no origin", "", "GET", "", "", http.StatusTeapot, true},
		{"allowed preflight", allowed, "OPTIONS", "PATCH", allowed, http.StatusNoContent, false},
		{"denied preflight", "https://example.com", "OPTIONS", "POST", "", http.StatusForbidden, false},
		{"ordinary options", allowed, "OPTIONS", "", allowed, http.StatusTeapot, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusTeapot)
			}), allowed)
			req := httptest.NewRequest(tc.method, "/auth/profile", nil)
			req.Header.Set("Origin", tc.origin)
			if tc.requestMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tc.requestMethod)
				req.Header.Set("Access-Control-Request-Headers", "authorization, content-type")
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tc.wantStatus || called != tc.wantCalled {
				t.Fatalf("status=%d called=%t", response.Code, called)
			}
			if got := response.Header().Get("Access-Control-Allow-Origin"); got != tc.wantOrigin {
				t.Errorf("allow origin = %q, want %q", got, tc.wantOrigin)
			}
			if got := response.Header().Get("Access-Control-Allow-Credentials"); (got == "true") != (tc.wantOrigin != "") {
				t.Errorf("allow credentials = %q", got)
			}
			if !strings.Contains(strings.Join(response.Header().Values("Vary"), ","), "Origin") {
				t.Error("missing Vary: Origin")
			}
			if tc.name == "allowed preflight" && (response.Header().Get("Access-Control-Allow-Methods") != "GET, HEAD, POST, PATCH" || response.Header().Get("Access-Control-Allow-Headers") != "Authorization, Content-Type") {
				t.Error("missing preflight permissions")
			}
		})
	}
}

func TestCORSDoesNotAllowWildcardWithCredentials(t *testing.T) {
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), "*")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "*")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Header().Get("Access-Control-Allow-Origin") != "" || response.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatal("wildcard must not allow credentialed requests")
	}
}
