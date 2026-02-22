package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticFileServing_ServesExistingFile(t *testing.T) {
	// Create temp directory with a test CSS file
	dir := t.TempDir()
	cssDir := filepath.Join(dir, "css")
	require.NoError(t, os.MkdirAll(cssDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cssDir, "main.css"), []byte("body { color: red; }"), 0o644))

	router := mux.NewRouter()
	staticFileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(dir)))
	router.PathPrefix("/static/").Handler(staticFileServer)

	req := httptest.NewRequest(http.MethodGet, "/static/css/main.css", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "body { color: red; }", rec.Body.String())
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/css")
}

func TestStaticFileServing_Returns404ForMissingFile(t *testing.T) {
	dir := t.TempDir()

	router := mux.NewRouter()
	staticFileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(dir)))
	router.PathPrefix("/static/").Handler(staticFileServer)

	req := httptest.NewRequest(http.MethodGet, "/static/nonexistent.css", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestStaticFileServing_ServesJavaScript(t *testing.T) {
	dir := t.TempDir()
	jsDir := filepath.Join(dir, "js")
	require.NoError(t, os.MkdirAll(jsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(jsDir, "main.js"), []byte("console.log('ok');"), 0o644))

	router := mux.NewRouter()
	staticFileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(dir)))
	router.PathPrefix("/static/").Handler(staticFileServer)

	req := httptest.NewRequest(http.MethodGet, "/static/js/main.js", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "console.log('ok');", rec.Body.String())
}

func TestStaticFileServing_ServesImages(t *testing.T) {
	dir := t.TempDir()
	imgDir := filepath.Join(dir, "images")
	require.NoError(t, os.MkdirAll(imgDir, 0o755))
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>`
	require.NoError(t, os.WriteFile(filepath.Join(imgDir, "favicon.svg"), []byte(svgContent), 0o644))

	router := mux.NewRouter()
	staticFileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(dir)))
	router.PathPrefix("/static/").Handler(staticFileServer)

	req := httptest.NewRequest(http.MethodGet, "/static/images/favicon.svg", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "image/svg+xml")
}
