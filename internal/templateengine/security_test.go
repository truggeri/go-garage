package templateengine

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestXSSEscaping_TemplateEngine verifies that Go's html/template package
// auto-escapes dangerous content in various HTML contexts.
func TestXSSEscaping_TemplateEngine(t *testing.T) {
	xssPayloads := []struct {
		name    string
		input   string
		escaped string // substring that must appear (HTML-entity-escaped form)
	}{
		{
			name:    "script tag",
			input:   `<script>alert('xss')</script>`,
			escaped: "&lt;script&gt;",
		},
		{
			name:    "img onerror",
			input:   `"><img src=x onerror=alert(1)>`,
			escaped: "&lt;img",
		},
		{
			name:    "svg onload",
			input:   `<svg onload=alert(1)>`,
			escaped: "&lt;svg",
		},
		{
			name:    "event handler attribute",
			input:   `" onmouseover="alert(1)"`,
			escaped: "&#34;",
		},
	}

	for _, tc := range xssPayloads {
		t.Run("escapes "+tc.name+" in element content", func(t *testing.T) {
			tmpl, err := template.New("test").Parse(`<div>{{.Content}}</div>`)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, map[string]string{"Content": tc.input})
			require.NoError(t, err)

			result := buf.String()
			assert.NotContains(t, result, tc.input, "raw payload must be escaped")
			assert.Contains(t, result, tc.escaped, "escaped form must be present")
		})

		t.Run("escapes "+tc.name+" in attribute value", func(t *testing.T) {
			tmpl, err := template.New("test").Parse(`<input value="{{.Value}}">`)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, map[string]string{"Value": tc.input})
			require.NoError(t, err)

			result := buf.String()
			assert.NotContains(t, result, tc.input, "raw payload must be escaped in attribute")
		})
	}

	t.Run("escapes script injection in href context", func(t *testing.T) {
		tmpl, err := template.New("test").Parse(`<a href="{{.URL}}">link</a>`)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]string{"URL": "javascript:alert(1)"})
		require.NoError(t, err)

		result := buf.String()
		// Go's html/template sanitises javascript: URIs to #ZgotmplZ.
		assert.NotContains(t, result, "javascript:", "javascript: URI must be sanitised")
	})
}

// TestSafeHTMLHelper verifies that the safeHTML helper only bypasses escaping
// when explicitly called, and that normal rendering still escapes.
func TestSafeHTMLHelper(t *testing.T) {
	t.Run("safeHTML marks content as safe", func(t *testing.T) {
		result := safeHTML("<b>bold</b>")
		assert.Equal(t, template.HTML("<b>bold</b>"), result)
	})

	t.Run("normal template rendering escapes HTML", func(t *testing.T) {
		funcMap := buildFuncMap()
		tmpl, err := template.New("test").Funcs(funcMap).Parse(`{{.Content}}`)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]string{"Content": "<script>alert(1)</script>"})
		require.NoError(t, err)

		assert.True(t, strings.Contains(buf.String(), "&lt;script&gt;"))
		assert.False(t, strings.Contains(buf.String(), "<script>"))
	})
}
