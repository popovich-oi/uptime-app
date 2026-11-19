package monitorcontroller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidURL(t *testing.T) {
	for _, value := range []string{"https://example.com", "http://localhost:8080/health"} {
		if !validURL(value) {
			t.Errorf("validURL(%q) = false", value)
		}
	}
	for _, value := range []string{"example.com", "ftp://example.com", "https://"} {
		if validURL(value) {
			t.Errorf("validURL(%q) = true", value)
		}
	}
}

func TestIntervalSeconds(t *testing.T) {
	for _, test := range []struct {
		value int
		unit  string
		want  int
	}{{5, "seconds", 5}, {10, "minutes", 600}, {2, "hours", 7200}} {
		got, err := intervalSeconds(test.value, test.unit)
		if err != nil || got != test.want {
			t.Errorf("intervalSeconds(%d, %q) = %d, %v; want %d", test.value, test.unit, got, err, test.want)
		}
	}
	for _, test := range []struct {
		value int
		unit  string
	}{{0, "seconds"}, {1, "days"}} {
		if _, err := intervalSeconds(test.value, test.unit); err == nil {
			t.Errorf("intervalSeconds(%d, %q) accepted invalid input", test.value, test.unit)
		}
	}
}

func TestInputLimits(t *testing.T) {
	for _, tc := range []struct {
		name, body        string
		tooLarge, invalid bool
	}{
		{"valid", `{"url":"https://example.com","interval_value":1,"interval_unit":"seconds"}`, false, false},
		{"multiple objects", `{} {}`, false, true},
		{"oversized", `{"url":"` + strings.Repeat("a", 17<<10) + `"}`, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var value input
			err := decodeInput(httptest.NewRecorder(), httptest.NewRequest("POST", "/monitors", strings.NewReader(tc.body)), &value)
			var limit *http.MaxBytesError
			if (err != nil) != tc.invalid || errors.As(err, &limit) != tc.tooLarge {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestURLAndIntervalBoundaries(t *testing.T) {
	base := "https://example.com/"
	if !validURL(base+strings.Repeat("a", 2048-len(base))) || validURL(base+strings.Repeat("a", 2049-len(base))) {
		t.Fatal("URL length boundary not enforced")
	}
	const max = 9007199254740991
	for _, tc := range []struct {
		unit       string
		multiplier int
	}{{"seconds", 1}, {"minutes", 60}, {"hours", 3600}} {
		if _, err := intervalSeconds(max/tc.multiplier, tc.unit); err != nil {
			t.Errorf("%s maximum rejected: %v", tc.unit, err)
		}
		if _, err := intervalSeconds(max/tc.multiplier+1, tc.unit); err == nil {
			t.Errorf("%s overflow accepted", tc.unit)
		}
	}
}

func TestRoutesRequireAuthorization(t *testing.T) {
	mux := http.NewServeMux()
	NewController(nil, nil).RegisterRoutes(mux)
	for _, method := range []string{"GET", "POST"} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, "/monitors", nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s: got %d", method, w.Code)
		}
	}
}
