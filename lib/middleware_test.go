package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.DebugMode)
	tests := []struct {
		method    string
		outExists bool
	}{
		{"OPTIONS", true},
		{"GET", false},
		{"POST", false},
	}
	expectedHeader := []struct {
		key   string
		value string
	}{
		{"Access-Control-Allow-Origin", "http://localhost"},
		{"Access-Control-Max-Age", "86400"},
		{"Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE"},
		{"Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, x-access-token"},
		{"Access-Control-Expose-Headers", "Content-Length"},
		{"Access-Control-Allow-Credentials", "true"},
	}
	f := CORSMiddleware()
	for _, tt := range tests {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, err := http.NewRequest(tt.method, "/", nil)
		if err != nil {
			t.Error(err)
		}
		c.Request = req
		f(c)
		for _, h := range expectedHeader {
			if c.Writer.Header().Get(h.key) != h.value {
				t.Errorf("key: %s, want: %s, get: %s", h.key, c.Writer.Header().Get(h.key), h.value)
			}
		}
	}
}
