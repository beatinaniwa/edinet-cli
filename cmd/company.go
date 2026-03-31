package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/beatinaniwa/edinet-cli/internal/api"
	"github.com/beatinaniwa/edinet-cli/internal/company"
	"github.com/beatinaniwa/edinet-cli/internal/extract"
	"github.com/beatinaniwa/edinet-cli/internal/service"
	"github.com/spf13/cobra"
)

var (
	companyFilingsDocType string
	companyFilingsFrom    string
	companyFilingsTo      string
	companyFilingsLimit   int
	companySearchIndustry string
)

const (
	edinetCodeListURL     = "https://disclosure2dl.edinet-fsa.go.jp/searchdocument/codelist/Edinetcode.zip"
	codelistCacheKey      = "codelist/edinetcode.csv"
	codelistCacheTTL      = 7 * 24 * time.Hour
)

var companyCmd = &cobra.Command{
	Use:   "company",
	Short: "Company search and filings lookup",
}

var companySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for companies by name, code, or industry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		reg, err := loadRegistry()
		if err != nil {
			return err
		}

		results := reg.Search(query)
		if companySearchIndustry != "" {
			// Filter search results by industry
			var filtered []company.CompanyEntry
			for _, e := range results {
				if strings.Contains(e.IndustryName, companySearchIndustry) {
					filtered = append(filtered, e)
				}
			}
			results = filtered
		}

		return outputResult(cmd.OutOrStdout(), results)
	},
}

var companyFilingsCmd = &cobra.Command{
	Use:   "filings <code>",
	Short: "List filings for a company (by securities code or EDINET code)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		code := args[0]

		if companyFilingsFrom != "" {
			if err := validateDate(companyFilingsFrom); err != nil {
				return &api.EDINETError{Code: api.ErrValidation, Message: err.Error()}
			}
		}
		if companyFilingsTo != "" {
			if err := validateDate(companyFilingsTo); err != nil {
				return &api.EDINETError{Code: api.ErrValidation, Message: err.Error()}
			}
		}
		if companyFilingsFrom != "" && companyFilingsTo != "" {
			from, _ := time.Parse("2006-01-02", companyFilingsFrom)
			to, _ := time.Parse("2006-01-02", companyFilingsTo)
			if from.After(to) {
				return &api.EDINETError{Code: api.ErrValidation, Message: "--from must be before or equal to --to"}
			}
		}

		if app.Config.SubscriptionKey == "" {
			return &api.EDINETError{Code: api.ErrAuth, Message: "EDINET_API_KEY environment variable is required"}
		}

		reg, err := loadRegistry()
		if err != nil {
			return err
		}

		client := api.NewClient(app.Config.SubscriptionKey, "https://api.edinet-fsa.go.jp", app.Config.Debug)
		docSvc := service.NewDocumentService(client, app.Cache, cmd.ErrOrStderr())
		companySvc := service.NewCompanyService(reg, docSvc)

		result, err := companySvc.Filings(cmd.Context(), code, service.FilingsOptions{
			DocType:   companyFilingsDocType,
			From:      companyFilingsFrom,
			To:        companyFilingsTo,
			RateLimit: 100 * time.Millisecond,
			Limit:     companyFilingsLimit,
		})
		if err != nil {
			return err
		}

		return outputResult(cmd.OutOrStdout(), result)
	},
}

var companyUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Download and update the EDINET code list",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), `{"progress":"downloading EDINET code list..."}`)

		csvData, err := downloadCodeListCSV()
		if err != nil {
			return &api.EDINETError{Code: api.ErrNetwork, Message: err.Error()}
		}

		reg := &company.Registry{}
		if err := reg.LoadFromCSV(csvData); err != nil {
			return fmt.Errorf("failed to parse code list: %w", err)
		}

		cached := false
		if app != nil && app.Cache != nil {
			if err := app.Cache.Set(codelistCacheKey, csvData); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), `{"warning":"failed to cache code list: %s"}`+"\n", err)
			} else {
				cached = true
			}
		}

		return outputResult(cmd.OutOrStdout(), map[string]any{
			"updated": true,
			"entries": len(reg.Entries),
			"cached":  cached,
		})
	},
}

// loadRegistry loads the EDINET code list, using cache when available.
func loadRegistry() (*company.Registry, error) {
	// Try cache first
	if app != nil && app.Cache != nil {
		if data, err := app.Cache.Get(codelistCacheKey, codelistCacheTTL); err == nil {
			reg := &company.Registry{}
			if err := reg.LoadFromCSV(data); err == nil {
				return reg, nil
			}
		}
	}

	// Cache miss or parse error — download fresh
	csvData, err := downloadCodeListCSV()
	if err != nil {
		return nil, err
	}

	reg := &company.Registry{}
	if err := reg.LoadFromCSV(csvData); err != nil {
		return nil, fmt.Errorf("failed to parse code list: %w", err)
	}

	// Save to cache for next time
	if app != nil && app.Cache != nil {
		_ = app.Cache.Set(codelistCacheKey, csvData)
	}

	return reg, nil
}

// downloadCodeListCSV downloads the EDINET code list ZIP and extracts the CSV.
func downloadCodeListCSV() ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(edinetCodeListURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download code list: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read code list: %w", err)
	}

	entries, err := extract.ReadFromZip(body, "*.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to extract code list ZIP: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no CSV file found in code list ZIP")
	}

	return entries[0].Data, nil
}

func init() {
	companySearchCmd.Flags().StringVar(&companySearchIndustry, "industry", "", "Filter by industry name")

	companyFilingsCmd.Flags().StringVar(&companyFilingsDocType, "doc-type", "", "Filter by document type code")
	companyFilingsCmd.Flags().StringVar(&companyFilingsFrom, "from", "", "Range start date (default: 365 days ago)")
	companyFilingsCmd.Flags().StringVar(&companyFilingsTo, "to", "", "Range end date (default: today)")
	companyFilingsCmd.Flags().IntVar(&companyFilingsLimit, "limit", 0, "Maximum number of results (0=unlimited)")

	companyCmd.AddCommand(companySearchCmd)
	companyCmd.AddCommand(companyFilingsCmd)
	companyCmd.AddCommand(companyUpdateCmd)
	rootCmd.AddCommand(companyCmd)
}
