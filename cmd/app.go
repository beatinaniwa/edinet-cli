package cmd

import (
	"io"

	"github.com/beatinaniwa/edinet-cli/internal/cache"
	"github.com/beatinaniwa/edinet-cli/internal/config"
)

// App holds application-level dependencies for all commands.
// Passed explicitly to subcommands for testability (no context.WithValue).
type App struct {
	Config *config.Config
	Cache  cache.Cache
	Stdout io.Writer
	Stderr io.Writer
}
