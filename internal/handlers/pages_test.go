package handlers

import (
	"testing"

	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestPageHandler(t *testing.T, authSvc services.AuthenticationService) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, authSvc, nil, nil, nil, nil)
}
