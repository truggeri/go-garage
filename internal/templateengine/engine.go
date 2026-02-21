// Package templateengine provides HTML template rendering for the Go-Garage web interface.
// It supports template inheritance with layouts, partials, and page-specific content,
// along with template caching for production and automatic reloading for development.
package templateengine

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Engine manages HTML template parsing, caching, and rendering.
type Engine struct {
	templatesDir string
	cache        map[string]*template.Template
	mu           sync.RWMutex
	devMode      bool
	funcMap      template.FuncMap
}

// NewEngine creates a new template Engine rooted at the given templates directory.
// When devMode is true, templates are re-parsed on every render to support live reloading.
func NewEngine(templatesDir string, devMode bool) *Engine {
	e := &Engine{
		templatesDir: templatesDir,
		cache:        make(map[string]*template.Template),
		devMode:      devMode,
		funcMap:      buildFuncMap(),
	}
	return e
}

// LoadTemplates parses all page templates and caches them with their associated layouts and partials.
// It should be called once at startup in production mode.
func (e *Engine) LoadTemplates() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	layoutFiles, err := filepath.Glob(filepath.Join(e.templatesDir, "layouts", "*.html"))
	if err != nil {
		return fmt.Errorf("loading layout templates: %w", err)
	}

	partialFiles, err := filepath.Glob(filepath.Join(e.templatesDir, "partials", "*.html"))
	if err != nil {
		return fmt.Errorf("loading partial templates: %w", err)
	}

	errorFiles, err := filepath.Glob(filepath.Join(e.templatesDir, "errors", "*.html"))
	if err != nil {
		return fmt.Errorf("loading error templates: %w", err)
	}

	pageFiles, err := findFiles(filepath.Join(e.templatesDir, "pages"))
	if err != nil {
		return fmt.Errorf("finding page templates: %w", err)
	}

	sharedFiles := append(layoutFiles, partialFiles...)

	// Parse each page template together with shared templates
	for _, page := range pageFiles {
		name, err := filepath.Rel(filepath.Join(e.templatesDir, "pages"), page)
		if err != nil {
			return fmt.Errorf("computing relative path for %s: %w", page, err)
		}

		files := make([]string, 0, len(sharedFiles)+1)
		files = append(files, sharedFiles...)
		files = append(files, page)

		tmpl, err := template.New(filepath.Base(page)).Funcs(e.funcMap).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("parsing page template %s: %w", name, err)
		}
		e.cache[name] = tmpl
	}

	// Parse each error template together with shared templates
	for _, errPage := range errorFiles {
		name := "errors/" + filepath.Base(errPage)

		files := make([]string, 0, len(sharedFiles)+1)
		files = append(files, sharedFiles...)
		files = append(files, errPage)

		tmpl, err := template.New(filepath.Base(errPage)).Funcs(e.funcMap).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("parsing error template %s: %w", name, err)
		}
		e.cache[name] = tmpl
	}

	return nil
}

// Render executes the named template with the given data and writes the result to w.
// The name should be a page template path relative to the pages/ directory (e.g. "home.html")
// or an error template path prefixed with "errors/" (e.g. "errors/404.html").
// The layout parameter specifies which layout block to execute (e.g. "base" or "auth").
func (e *Engine) Render(w io.Writer, name, layout string, data interface{}) error {
	tmpl, err := e.getTemplate(name)
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, layout, data)
}

// getTemplate retrieves a cached template or parses it fresh in development mode.
func (e *Engine) getTemplate(name string) (*template.Template, error) {
	if e.devMode {
		return e.parseTemplate(name)
	}

	e.mu.RLock()
	tmpl, ok := e.cache[name]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("template %q not found in cache", name)
	}
	return tmpl, nil
}

// parseTemplate builds a template from disk by combining shared templates with the named page.
func (e *Engine) parseTemplate(name string) (*template.Template, error) {
	layoutFiles, err := filepath.Glob(filepath.Join(e.templatesDir, "layouts", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("loading layout templates: %w", err)
	}

	partialFiles, err := filepath.Glob(filepath.Join(e.templatesDir, "partials", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("loading partial templates: %w", err)
	}

	var pagePath string
	if len(name) > 7 && name[:7] == "errors/" {
		pagePath = filepath.Join(e.templatesDir, name)
	} else {
		pagePath = filepath.Join(e.templatesDir, "pages", name)
	}

	if _, statErr := os.Stat(pagePath); statErr != nil {
		return nil, fmt.Errorf("template file %q not found: %w", name, statErr)
	}

	files := make([]string, 0, len(layoutFiles)+len(partialFiles)+1)
	files = append(files, layoutFiles...)
	files = append(files, partialFiles...)
	files = append(files, pagePath)

	tmpl, err := template.New(filepath.Base(pagePath)).Funcs(e.funcMap).ParseFiles(files...)
	if err != nil {
		return nil, fmt.Errorf("parsing template %s: %w", name, err)
	}

	return tmpl, nil
}

// findFiles walks a directory tree and returns all .html files found.
func findFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
