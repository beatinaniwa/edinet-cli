package cmd

import (
	"os"

	"github.com/beatinaniwa/edinet-cli/internal/cache"
	"github.com/beatinaniwa/edinet-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	formatFlag  string
	debugFlag   bool
	noCacheFlag bool
	app         *App
)

var rootCmd = &cobra.Command{
	Use:   "edinet",
	Short: "EDINET API v2 CLI for AI agents",
	Long:  "A structured wrapper around the EDINET API v2 for autonomous document retrieval and analysis by AI agents.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.NoCache = noCacheFlag
		cfg.Debug = debugFlag
		cfg.Format = formatFlag

		var c cache.Cache = cache.NoopCache{}
		if !cfg.NoCache {
			fs, err := cache.NewFileStore(cfg.CacheDir)
			if err == nil {
				c = fs
			}
			// If cache dir creation fails, silently fall back to NoopCache
		}

		app = &App{
			Config: cfg,
			Cache:  c,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "json", "Output format: json or table")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug output on stderr")
	rootCmd.PersistentFlags().BoolVar(&noCacheFlag, "no-cache", false, "Bypass local cache")
}

// Execute runs the root command and returns the exit code.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		return exitError(os.Stderr, err)
	}
	return 0
}
