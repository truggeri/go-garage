package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestPageHandler(t *testing.T, authSvc services.AuthenticationService) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, authSvc, nil, nil, nil, nil, nil)
}

func serveHTMXPage(t *testing.T, handler http.HandlerFunc, rec *httptest.ResponseRecorder, req *http.Request) {
	t.Helper()
	middleware.CSRFProtection("test-csrf-secret")(handler).ServeHTTP(rec, req)
	assert.Regexp(t, `<meta name="csrf-token" content="[0-9a-f]+\.[0-9a-f]+">`, rec.Body.String())
}
