package schema

// DocType represents a document type code entry.
type DocType struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CommandInfo describes a CLI command for machine consumption.
type CommandInfo struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Flags       []FlagInfo `json:"flags,omitempty"`
	Examples    []string   `json:"examples,omitempty"`
}

// FlagInfo describes a CLI flag.
type FlagInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

// SectionInfo describes a known section for text extraction.
type SectionInfo struct {
	ID    string   `json:"id"`
	Names []string `json:"names"`
}

// ListDocTypes returns all EDINET document type codes.
func ListDocTypes() []DocType {
	return []DocType{
		{Code: "010", Name: "有価証券通知書"},
		{Code: "020", Name: "変更通知書（有価証券通知書）"},
		{Code: "030", Name: "有価証券届出書"},
		{Code: "040", Name: "訂正有価証券届出書"},
		{Code: "050", Name: "届出の取下げ願い"},
		{Code: "060", Name: "発行登録通知書"},
		{Code: "070", Name: "変更通知書（発行登録通知書）"},
		{Code: "080", Name: "発行登録書"},
		{Code: "090", Name: "訂正発行登録書"},
		{Code: "100", Name: "発行登録追補書類"},
		{Code: "110", Name: "発行登録取下届出書"},
		{Code: "120", Name: "有価証券報告書"},
		{Code: "130", Name: "訂正有価証券報告書"},
		{Code: "135", Name: "確認書"},
		{Code: "136", Name: "訂正確認書"},
		{Code: "140", Name: "四半期報告書"},
		{Code: "150", Name: "訂正四半期報告書"},
		{Code: "160", Name: "半期報告書"},
		{Code: "170", Name: "訂正半期報告書"},
		{Code: "180", Name: "臨時報告書"},
		{Code: "190", Name: "訂正臨時報告書"},
		{Code: "200", Name: "親会社等状況報告書"},
		{Code: "210", Name: "訂正親会社等状況報告書"},
		{Code: "220", Name: "自己株券買付状況報告書"},
		{Code: "230", Name: "訂正自己株券買付状況報告書"},
		{Code: "235", Name: "内部統制報告書"},
		{Code: "236", Name: "訂正内部統制報告書"},
		{Code: "240", Name: "公開買付届出書"},
		{Code: "250", Name: "訂正公開買付届出書"},
		{Code: "260", Name: "公開買付撤回届出書"},
		{Code: "270", Name: "公開買付報告書"},
		{Code: "280", Name: "訂正公開買付報告書"},
		{Code: "290", Name: "意見表明報告書"},
		{Code: "300", Name: "訂正意見表明報告書"},
		{Code: "310", Name: "対質問回答報告書"},
		{Code: "320", Name: "訂正対質問回答報告書"},
		{Code: "330", Name: "別途買付け禁止の特例を受けるための申出書"},
		{Code: "340", Name: "訂正別途買付け禁止の特例を受けるための申出書"},
		{Code: "350", Name: "大量保有報告書"},
		{Code: "360", Name: "訂正大量保有報告書"},
		{Code: "370", Name: "基準日の届出書"},
		{Code: "380", Name: "変更の届出書"},
	}
}

// ListCommands returns descriptions of all CLI commands.
func ListCommands() []CommandInfo {
	return []CommandInfo{
		{
			Name:        "doc list",
			Description: "List documents for a date or date range",
			Flags: []FlagInfo{
				{Name: "--date", Type: "string", Description: "Single date (YYYY-MM-DD)"},
				{Name: "--from", Type: "string", Description: "Range start date"},
				{Name: "--to", Type: "string", Description: "Range end date"},
				{Name: "--doc-type", Type: "string", Description: "Filter by document type code"},
				{Name: "--sec-code", Type: "string", Description: "Filter by securities code"},
				{Name: "--edinet-code", Type: "string", Description: "Filter by EDINET code"},
				{Name: "--filer-name", Type: "string", Description: "Filter by filer name (substring)"},
				{Name: "--rate-limit", Type: "int", Default: "100", Description: "Request interval in ms"},
			},
			Examples: []string{
				"edinet doc list --date 2025-06-20",
				"edinet doc list --from 2025-06-01 --to 2025-06-30 --doc-type 120",
			},
		},
		{
			Name:        "doc get",
			Description: "Download a document (PDF, CSV, XBRL, etc.)",
			Flags: []FlagInfo{
				{Name: "--type", Type: "string", Required: true, Description: "xbrl|pdf|attach|english|csv"},
				{Name: "--out", Type: "string", Default: ".", Description: "Output directory"},
			},
			Examples: []string{
				"edinet doc get S100ABCD --type csv --out ./data/",
				"edinet doc get S100ABCD --type pdf",
			},
		},
		{
			Name:        "doc data",
			Description: "Extract financial data from CSV (experimental)",
			Examples:    []string{"edinet doc data S100ABCD"},
		},
		{
			Name:        "doc text",
			Description: "Extract text from document HTML (best-effort)",
			Flags: []FlagInfo{
				{Name: "--section", Type: "string", Description: "Section ID or heading pattern"},
				{Name: "--list-sections", Type: "bool", Description: "List available sections"},
			},
			Examples: []string{
				"edinet doc text S100ABCD",
				"edinet doc text S100ABCD --section risk",
				"edinet doc text --list-sections",
			},
		},
		{
			Name:        "doc financial",
			Description: "Extract structured financial statements from CSV",
			Flags: []FlagInfo{
				{Name: "--statement", Type: "string", Default: "all", Description: "Statement type: bs, pl, cf, all"},
				{Name: "--non-consolidated", Type: "bool", Description: "Prefer non-consolidated statements"},
				{Name: "--summary-only", Type: "bool", Description: "Output only summary metrics without detailed statements"},
			},
			Examples: []string{
				"edinet doc financial S100ABCD",
				"edinet doc financial S100ABCD --statement pl",
				"edinet doc financial S100ABCD --non-consolidated",
			},
		},
		{
			Name:        "company search",
			Description: "Search for companies by name, code, or industry",
			Flags: []FlagInfo{
				{Name: "--industry", Type: "string", Description: "Filter by industry name"},
			},
			Examples: []string{
				"edinet company search トヨタ",
				"edinet company search 7203",
			},
		},
		{
			Name:        "company filings",
			Description: "List filings for a company",
			Flags: []FlagInfo{
				{Name: "--doc-type", Type: "string", Description: "Filter by document type code"},
				{Name: "--from", Type: "string", Description: "Range start date (default: 365 days ago)"},
				{Name: "--to", Type: "string", Description: "Range end date (default: today)"},
				{Name: "--limit", Type: "int", Default: "0", Description: "Max results (0=unlimited)"},
			},
			Examples: []string{"edinet company filings 7203 --doc-type 120 --limit 5"},
		},
		{
			Name:        "company financials",
			Description: "Extract financial statements for multiple fiscal periods",
			Flags: []FlagInfo{
				{Name: "--periods", Type: "int", Default: "3", Description: "Number of fiscal periods (1-10)"},
				{Name: "--statement", Type: "string", Default: "all", Description: "Statement type: bs, pl, cf, all"},
				{Name: "--non-consolidated", Type: "bool", Description: "Prefer non-consolidated statements"},
				{Name: "--summary-only", Type: "bool", Description: "Output only summary metrics without detailed statements"},
			},
			Examples: []string{
				"edinet company financials E02144",
				"edinet company financials 7203 --periods 5",
				"edinet company financials トヨタ --statement pl",
			},
		},
		{
			Name:        "company update",
			Description: "Download and update the EDINET code list",
			Examples:    []string{"edinet company update"},
		},
		{
			Name:        "schema commands",
			Description: "List all CLI commands with flags",
		},
		{
			Name:        "schema derived-metrics",
			Description: "List all derived financial metrics with formulas",
		},
		{
			Name:        "schema doc-types",
			Description: "List all document type codes",
		},
		{
			Name:        "schema sections",
			Description: "List known sections for text extraction",
		},
		{
			Name:        "schema financial-elements",
			Description: "List all known financial XBRL element mappings",
		},
	}
}

// ListSections returns known sections for text extraction.
func ListSections() []SectionInfo {
	return []SectionInfo{
		{ID: "business", Names: []string{"事業の内容", "事業の概要"}},
		{ID: "risk", Names: []string{"事業等のリスク"}},
		{ID: "mda", Names: []string{"経営者による財政状態", "経営成績等の状況の概要"}},
		{ID: "governance", Names: []string{"コーポレート・ガバナンス"}},
		{ID: "financial", Names: []string{"財務諸表", "連結財務諸表"}},
		{ID: "employees", Names: []string{"従業員の状況"}},
		{ID: "facilities", Names: []string{"設備の状況"}},
		{ID: "history", Names: []string{"沿革"}},
		{ID: "shares", Names: []string{"株式等の状況", "株式の総数等"}},
		{ID: "dividends", Names: []string{"配当政策"}},
	}
}
