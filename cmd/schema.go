package cmd

import (
	"github.com/beatinaniwa/edinet-cli/internal/schema"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Machine-readable CLI metadata for AI agents",
}

var schemaCommandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "List all CLI commands with flags and examples",
	RunE: func(cmd *cobra.Command, args []string) error {
		return outputResult(cmd.OutOrStdout(), schema.ListCommands())
	},
}

var schemaDocTypesCmd = &cobra.Command{
	Use:   "doc-types",
	Short: "List all EDINET document type codes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return outputResult(cmd.OutOrStdout(), schema.ListDocTypes())
	},
}

var schemaSectionsCmd = &cobra.Command{
	Use:   "sections",
	Short: "List known sections for text extraction",
	RunE: func(cmd *cobra.Command, args []string) error {
		return outputResult(cmd.OutOrStdout(), schema.ListSections())
	},
}

func init() {
	schemaCmd.AddCommand(schemaCommandsCmd)
	schemaCmd.AddCommand(schemaDocTypesCmd)
	schemaCmd.AddCommand(schemaSectionsCmd)
	rootCmd.AddCommand(schemaCmd)
}
