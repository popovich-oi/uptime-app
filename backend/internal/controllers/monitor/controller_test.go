package monitorcontroller

import "testing"

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
