package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStoppedHandlerRendersDomainAndPort(t *testing.T) {
	handler := NewStoppedHandler("cayacart.test", 3000)

	req := httptest.NewRequest("GET", "http://cayacart.test/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(body, "cayacart.test") {
		t.Error("response body should contain domain name")
	}
	if !strings.Contains(body, "3000") {
		t.Error("response body should contain port number")
	}
	if !strings.Contains(body, "__hatch/ws") {
		t.Error("response body should contain websocket path")
	}
}
