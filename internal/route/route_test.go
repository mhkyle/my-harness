package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthcheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	Register(r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "OK" {
		t.Fatalf("expected body %q, got %q", "OK", w.Body.String())
	}
}
