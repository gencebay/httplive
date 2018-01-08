package lib

import (
	"regexp"
	"testing"
)

func TestNewUUID(t *testing.T) {
	re := regexp.MustCompile("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
	for i := 0; i < 50; i++ {
		actual, err := NewUUID()
		if err != nil {
			t.Errorf("error is not nil: %v", err)
		}
		if !re.MatchString(actual) {
			t.Errorf("uuid: %s, len: %d", actual, len(actual))
		}
	}
}

func TestCreateEndpointKey(t *testing.T) {
	tests := []struct {
		method   string
		endpoint string
		out      string
	}{
		{"", "", ""},
		{"post", "foo", "postfoo"},
		{"post", "FOO", "postfoo"},
		{"POST", "foo", "postfoo"},
		{"POST", "FOO", "postfoo"},
		{"ÄËÏ", "ÖÜ", "äëïöü"},
		{"POST", "///", "post///"},
	}
	for _, tt := range tests {
		actual := CreateEndpointKey(tt.method, tt.endpoint)
		if tt.out != actual {
			t.Errorf("want: %s, give: %s, method: %s, endpoint: %s", tt.out, actual, tt.method, tt.endpoint)
		}
	}
}
