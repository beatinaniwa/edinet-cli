package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/beatinaniwa/edinet-cli/internal/api"
	"github.com/beatinaniwa/edinet-cli/internal/extract"
	"github.com/beatinaniwa/edinet-cli/internal/service"
	"github.com/spf13/cobra"
)

var (
	docListDate       string
	docListFrom       string
	docListTo         string
	docListDocType    string
	docListSecCode    string
	docListEdinetCode string
	docListFilerName  string
	docListRateLimit  int

	docGetType string
	docGetOut  string

	docTextSection      string
	docTextListSections bool

	docFinancialStatement      string
	docFinancialNonConsolidated bool
)

var downloadTypeMap = map[string]int{
	"xbrl":    1,
	"pdf":     2,
	"attach":  3,
	"english": 4,
	"csv":     5,
}

var downloadExtMap = map[string]string{
	"xbrl":    "zip",
	"pdf":     "pdf",
	"attach":  "zip",
	"english": "zip",
	"csv":     "zip",
}

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Document operations (list, get, data, text)",
}

var docListCmd = &cobra.Command{
	Use:   "list",
	Short: "List documents for a date or date range",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateDocListFlags(); err != nil {
			return &api.EDINETError{Code: api.ErrValidation, Message: err.Error()}
		}

		if app.Config.SubscriptionKey == "" {
			return &api.EDINETError{Code: api.ErrAuth, Message: "EDINET_API_KEY environment variable is required"}
		}

		client := api.NewClient(app.Config.SubscriptionKey, "https://api.edinet-fsa.go.jp", app.Config.Debug)
		svc := service.NewDocumentService(client, app.Cache, cmd.ErrOrStderr())

		result, err := svc.List(cmd.Context(), service.ListOptions{
			Date:       docListDate,
			From:       docListFrom,
			To:         docListTo,
			DocType:    docListDocType,
			SecCode:    docListSecCode,
			EdinetCode: docListEdinetCode,
			FilerName:  docListFilerName,
			RateLimit:  time.Duration(docListRateLimit) * time.Millisecond,
		})
		if err != nil {
			return err
		}

		return outputResult(cmd.OutOrStdout(), result)
	},
}

var docGetCmd = &cobra.Command{
	Use:   "get <docID>",
	Short: "Download a document (PDF, CSV, XBRL, etc.)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]
		if !isValidDocID(docID) {
			return &api.EDINETError{Code: api.ErrValidation, Message: fmt.Sprintf("invalid docID %q: must match S followed by digits (e.g. S100ABCD)", docID)}
		}

		apiType, ok := downloadTypeMap[docGetType]
		if !ok || docGetType == "" {
			return &api.EDINETError{Code: api.ErrValidation, Message: fmt.Sprintf("invalid --type %q, must be one of: xbrl, pdf, attach, english, csv", docGetType)}
		}

		if app.Config.SubscriptionKey == "" {
			return &api.EDINETError{Code: api.ErrAuth, Message: "EDINET_API_KEY environment variable is required"}
		}

		client := api.NewClient(app.Config.SubscriptionKey, "https://api.edinet-fsa.go.jp", app.Config.Debug)
		body, _, err := client.DownloadDocument(cmd.Context(), docID, apiType)
		if err != nil {
			return err
		}

		// Create output directory
		if err := os.MkdirAll(docGetOut, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Atomic write: temp + rename
		ext := downloadExtMap[docGetType]
		filename := fmt.Sprintf("%s_%s.%s", docID, docGetType, ext)
		destPath := filepath.Join(docGetOut, filename)
		absPath, _ := filepath.Abs(destPath)

		tmpFile, err := os.CreateTemp(docGetOut, ".download-*")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		if _, err := tmpFile.Write(body); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpFile.Name())
			return fmt.Errorf("failed to write file: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			_ = os.Remove(tmpFile.Name())
			return fmt.Errorf("failed to close temp file: %w", err)
		}
		if err := os.Rename(tmpFile.Name(), destPath); err != nil {
			_ = os.Remove(tmpFile.Name())
			return fmt.Errorf("failed to rename file: %w", err)
		}

		return outputResult(cmd.OutOrStdout(), map[string]any{
			"path": absPath,
			"size": len(body),
			"type": docGetType,
		})
	},
}

var docDataCmd = &cobra.Command{
	Use:   "data <docID>",
	Short: "Extract financial data from CSV (experimental)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]
		if !isValidDocID(docID) {
			return &api.EDINETError{Code: api.ErrValidation, Message: fmt.Sprintf("invalid docID %q", docID)}
		}

		if app.Config.SubscriptionKey == "" {
			return &api.EDINETError{Code: api.ErrAuth, Message: "EDINET_API_KEY environment variable is required"}
		}

		client := api.NewClient(app.Config.SubscriptionKey, "https://api.edinet-fsa.go.jp", app.Config.Debug)
		body, _, err := client.DownloadDocument(cmd.Context(), docID, 5) // type=5 CSV
		if err != nil {
			return err
		}

		result, err := extract.ExtractCSVData(body)
		if err != nil {
			return fmt.Errorf("failed to extract CSV data: %w", err)
		}

		return outputResult(cmd.OutOrStdout(), result)
	},
}

var docFinancialCmd = &cobra.Command{
	Use:   "financial <docID>",
	Short: "Extract structured financial statements from CSV",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]
		if !isValidDocID(docID) {
			return &api.EDINETError{Code: api.ErrValidation, Message: fmt.Sprintf("invalid docID %q: must match S followed by digits (e.g. S100ABCD)", docID)}
		}
		if err := validateStatement(docFinancialStatement); err != nil {
			return err
		}
		if app.Config.SubscriptionKey == "" {
			return &api.EDINETError{Code: api.ErrAuth, Message: "EDINET_API_KEY environment variable is required"}
		}

		client := api.NewClient(app.Config.SubscriptionKey, "https://api.edinet-fsa.go.jp", app.Config.Debug)
		svc := service.NewFinancialService(client, app.Cache)

		opts := service.StatementOpts{
			Statement: docFinancialStatement,
		}
		if docFinancialNonConsolidated {
			opts.Consolidated = ptrBool(false)
		}

		result, err := svc.GetStatements(cmd.Context(), docID, opts)
		if err != nil {
			return err
		}
		return outputResult(cmd.OutOrStdout(), result)
	},
}

var docTextCmd = &cobra.Command{
	Use:   "text [docID]",
	Short: "Extract text from document HTML (best-effort)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// --list-sections mode: no docID needed
		if docTextListSections {
			return outputResult(cmd.OutOrStdout(), extract.KnownSections)
		}

		if len(args) == 0 {
			return &api.EDINETError{Code: api.ErrValidation, Message: "docID argument is required"}
		}
		docID := args[0]
		if !isValidDocID(docID) {
			return &api.EDINETError{Code: api.ErrValidation, Message: fmt.Sprintf("invalid docID %q", docID)}
		}

		if app.Config.SubscriptionKey == "" {
			return &api.EDINETError{Code: api.ErrAuth, Message: "EDINET_API_KEY environment variable is required"}
		}

		client := api.NewClient(app.Config.SubscriptionKey, "https://api.edinet-fsa.go.jp", app.Config.Debug)
		body, _, err := client.DownloadDocument(cmd.Context(), docID, 1) // type=1 XBRL
		if err != nil {
			return err
		}

		if docTextSection != "" {
			sections, err := extract.ExtractSections(body)
			if err != nil {
				return fmt.Errorf("failed to extract sections: %w", err)
			}
			for _, s := range sections {
				if s.ID == docTextSection || strings.Contains(s.Name, docTextSection) {
					return outputResult(cmd.OutOrStdout(), map[string]string{
						"section": s.ID,
						"text":    s.Text,
					})
				}
			}
			// Section not found — return full text with warning
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), `{"warning":"section '%s' not found, returning full text"}`+"\n", docTextSection)
		}

		text, err := extract.ExtractText(body)
		if err != nil {
			return fmt.Errorf("failed to extract text: %w", err)
		}
		return outputResult(cmd.OutOrStdout(), map[string]string{"text": text})
	},
}

func validateDocListFlags() error {
	hasDate := docListDate != ""
	hasFrom := docListFrom != ""
	hasTo := docListTo != ""

	if !hasDate && !hasFrom && !hasTo {
		return fmt.Errorf("either --date or --from/--to is required")
	}
	if hasDate && (hasFrom || hasTo) {
		return fmt.Errorf("--date and --from/--to are mutually exclusive")
	}
	if hasFrom != hasTo {
		return fmt.Errorf("--from and --to must be specified together")
	}

	// Validate date format and range
	dates := []string{}
	if hasDate {
		dates = append(dates, docListDate)
	}
	if hasFrom {
		dates = append(dates, docListFrom, docListTo)
	}
	for _, d := range dates {
		if err := validateDate(d); err != nil {
			return err
		}
	}

	if hasFrom && hasTo {
		from, _ := time.Parse("2006-01-02", docListFrom)
		to, _ := time.Parse("2006-01-02", docListTo)
		if from.After(to) {
			return fmt.Errorf("--from must be before or equal to --to")
		}
	}

	return nil
}

// jst is the Asia/Tokyo timezone used for EDINET date validation.
var jst = time.FixedZone("JST", 9*60*60)

var docIDPattern = regexp.MustCompile(`^S\w+$`)

// isValidDocID checks that the docID contains no path separators or traversal characters.
func isValidDocID(id string) bool {
	return docIDPattern.MatchString(id)
}

func validateDate(d string) error {
	t, err := time.ParseInLocation("2006-01-02", d, jst)
	if err != nil {
		return fmt.Errorf("invalid date format %q, expected YYYY-MM-DD", d)
	}
	now := time.Now().In(jst)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)
	if t.After(today) {
		return fmt.Errorf("date %q is in the future", d)
	}
	tenYearsAgo := today.AddDate(-10, 0, 0)
	if t.Before(tenYearsAgo) {
		return fmt.Errorf("date %q is more than 10 years ago (EDINET limit)", d)
	}
	return nil
}

func init() {
	docListCmd.Flags().StringVar(&docListDate, "date", "", "Single date (YYYY-MM-DD)")
	docListCmd.Flags().StringVar(&docListFrom, "from", "", "Range start date (YYYY-MM-DD)")
	docListCmd.Flags().StringVar(&docListTo, "to", "", "Range end date (YYYY-MM-DD)")
	docListCmd.Flags().StringVar(&docListDocType, "doc-type", "", "Filter by document type code")
	docListCmd.Flags().StringVar(&docListSecCode, "sec-code", "", "Filter by securities code")
	docListCmd.Flags().StringVar(&docListEdinetCode, "edinet-code", "", "Filter by EDINET code")
	docListCmd.Flags().StringVar(&docListFilerName, "filer-name", "", "Filter by filer name (substring match)")
	docListCmd.Flags().IntVar(&docListRateLimit, "rate-limit", 100, "Rate limit between requests in ms")

	docGetCmd.Flags().StringVar(&docGetType, "type", "", "Document type: xbrl, pdf, attach, english, csv (required)")
	docGetCmd.Flags().StringVar(&docGetOut, "out", ".", "Output directory")

	docTextCmd.Flags().StringVar(&docTextSection, "section", "", "Section ID or heading pattern")
	docTextCmd.Flags().BoolVar(&docTextListSections, "list-sections", false, "List available sections")

	docFinancialCmd.Flags().StringVar(&docFinancialStatement, "statement", "all", "Statement type: bs, pl, cf, all")
	docFinancialCmd.Flags().BoolVar(&docFinancialNonConsolidated, "non-consolidated", false, "Prefer non-consolidated statements")

	docCmd.AddCommand(docListCmd)
	docCmd.AddCommand(docGetCmd)
	docCmd.AddCommand(docDataCmd)
	docCmd.AddCommand(docTextCmd)
	docCmd.AddCommand(docFinancialCmd)
	rootCmd.AddCommand(docCmd)
}
