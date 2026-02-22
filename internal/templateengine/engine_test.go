package templateengine

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestTemplates sets up a temporary template directory with test files.
func createTestTemplates(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	dirs := []string{
		filepath.Join(dir, "layouts"),
		filepath.Join(dir, "partials"),
		filepath.Join(dir, "pages"),
		filepath.Join(dir, "errors"),
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(d, 0o755))
	}

	// Layout
	writeFile(t, filepath.Join(dir, "layouts", "base.html"),
		`{{define "base"}}<!DOCTYPE html><html><head><title>{{block "title" .}}Test{{end}}</title></head><body>{{template "nav" .}}{{template "flash-messages" .}}{{block "content" .}}{{end}}{{template "footer" .}}</body></html>{{end}}`)

	writeFile(t, filepath.Join(dir, "layouts", "auth.html"),
		`{{define "auth"}}<!DOCTYPE html><html><head><title>{{block "title" .}}Auth{{end}}</title></head><body class="auth">{{template "flash-messages" .}}{{block "content" .}}{{end}}{{template "footer" .}}</body></html>{{end}}`)

	// Partials
	writeFile(t, filepath.Join(dir, "partials", "navigation.html"),
		`{{define "nav"}}<nav>Navigation</nav>{{end}}`)

	writeFile(t, filepath.Join(dir, "partials", "footer.html"),
		`{{define "footer"}}<footer>&copy; {{currentYear}}</footer>{{end}}`)

	writeFile(t, filepath.Join(dir, "partials", "flash-messages.html"),
		`{{define "flash-messages"}}{{end}}`)

	// Page
	writeFile(t, filepath.Join(dir, "pages", "home.html"),
		`{{define "title"}}Home{{end}}{{define "content"}}<h1>Welcome {{.Name}}</h1>{{end}}`)

	// Error page
	writeFile(t, filepath.Join(dir, "errors", "404.html"),
		`{{define "title"}}Not Found{{end}}{{define "content"}}<h1>404</h1>{{end}}`)

	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func TestNewEngine_CreatesInstance(t *testing.T) {
	e := NewEngine("/tmp/test", true)

	assert.NotNil(t, e)
	assert.Equal(t, "/tmp/test", e.templatesDir)
	assert.True(t, e.devMode)
	assert.NotNil(t, e.cache)
	assert.NotNil(t, e.funcMap)
}

func TestLoadTemplates_ParsesAllTemplates(t *testing.T) {
	dir := createTestTemplates(t)
	e := NewEngine(dir, false)

	err := e.LoadTemplates()

	require.NoError(t, err)
	assert.Contains(t, e.cache, "home.html")
	assert.Contains(t, e.cache, "errors/404.html")
}

func TestRender_PageWithBaseLayout(t *testing.T) {
	dir := createTestTemplates(t)
	e := NewEngine(dir, false)
	require.NoError(t, e.LoadTemplates())

	var buf bytes.Buffer
	data := map[string]interface{}{"Name": "TestUser"}
	err := e.Render(&buf, "home.html", "base", data)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "<title>Home</title>")
	assert.Contains(t, output, "<h1>Welcome TestUser</h1>")
	assert.Contains(t, output, "<nav>Navigation</nav>")
	assert.Contains(t, output, "<footer>")
}

func TestRender_PageWithAuthLayout(t *testing.T) {
	dir := createTestTemplates(t)
	e := NewEngine(dir, false)
	require.NoError(t, e.LoadTemplates())

	var buf bytes.Buffer
	err := e.Render(&buf, "home.html", "auth", nil)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, `class="auth"`)
	assert.Contains(t, output, "<h1>Welcome")
}

func TestRender_ErrorTemplate(t *testing.T) {
	dir := createTestTemplates(t)
	e := NewEngine(dir, false)
	require.NoError(t, e.LoadTemplates())

	var buf bytes.Buffer
	err := e.Render(&buf, "errors/404.html", "base", nil)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "<title>Not Found</title>")
	assert.Contains(t, output, "<h1>404</h1>")
}

func TestRender_TemplateNotFound_ReturnsError(t *testing.T) {
	dir := createTestTemplates(t)
	e := NewEngine(dir, false)
	require.NoError(t, e.LoadTemplates())

	var buf bytes.Buffer
	err := e.Render(&buf, "nonexistent.html", "base", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRender_DevMode_ReloadsTemplates(t *testing.T) {
	dir := createTestTemplates(t)
	e := NewEngine(dir, true)

	// Render initial version
	var buf bytes.Buffer
	data := map[string]interface{}{"Name": "Dev"}
	err := e.Render(&buf, "home.html", "base", data)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Welcome Dev")

	// Modify the template
	writeFile(t, filepath.Join(dir, "pages", "home.html"),
		`{{define "title"}}Updated{{end}}{{define "content"}}<h1>Updated {{.Name}}</h1>{{end}}`)

	// Render again – dev mode should pick up changes
	buf.Reset()
	err = e.Render(&buf, "home.html", "base", data)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Updated Dev")
	assert.Contains(t, buf.String(), "<title>Updated</title>")
}

func TestRender_DevMode_MissingFile_ReturnsError(t *testing.T) {
	dir := createTestTemplates(t)
	e := NewEngine(dir, true)

	var buf bytes.Buffer
	err := e.Render(&buf, "missing.html", "base", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoadTemplates_InvalidDir_ReturnsError(t *testing.T) {
	e := NewEngine("/nonexistent/path", false)
	err := e.LoadTemplates()

	assert.Error(t, err)
}
