//go:build tools
// +build tools

// This file declares dependencies on tool binaries
package tools

import (
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/mattn/go-sqlite3"
)
