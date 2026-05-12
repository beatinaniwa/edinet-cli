package extract

import "strings"

// SectionDef defines a known section of Japanese financial filings.
type SectionDef struct {
	ID    string   `json:"id"`
	Names []string `json:"names"`
}

// Section represents an extracted section from a document.
type Section struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Text string `json:"text"`
}

// KnownSections defines the recognized sections in EDINET filings.
var KnownSections = []SectionDef{
	{ID: "business", Names: []string{"事業の内容", "事業の概要"}},
	{ID: "risk", Names: []string{"事業等のリスク"}},
	{ID: "mda", Names: []string{"経営者による財政状態", "経営成績等の状況の概要", "経営者による経営成績等の状況に関する分析・検討内容"}},
	{ID: "governance", Names: []string{"コーポレート・ガバナンス", "コーポレート・ガバナンスの概要"}},
	{ID: "financial", Names: []string{"財務諸表", "連結財務諸表"}},
	{ID: "employees", Names: []string{"従業員の状況"}},
	{ID: "facilities", Names: []string{"設備の状況", "設備の新設、除却等の計画"}},
	{ID: "history", Names: []string{"沿革"}},
	{ID: "shares", Names: []string{"株式等の状況", "株式の総数等"}},
	{ID: "dividends", Names: []string{"配当政策"}},
}

// normalizeForMatch normalizes a string for tolerant heading comparison.
// EDINET 提出企業ごとに 中黒 (・) の有無や前後の空白が揺れているため、
// マッチング前に正規化する。 例: 株式会社セブン＆アイ・ホールディングス は
// 「コーポレートガバナンス」 (中黒なし)、 日本マクドナルドホールディングス
// は「コーポレート・ガバナンス」 (中黒あり) で同じ章を表す。
func normalizeForMatch(s string) string {
	// Remove the katakana middle dot (・, U+30FB) and ASCII / full-width
	// whitespace so headings that differ only in these decorative chars still
	// match the same KnownSections entry.
	r := strings.NewReplacer(
		"・", "",
		" ", "",
		"　", "",
		"\t", "",
		"\n", "",
		"\r", "",
	)
	return r.Replace(s)
}

// MatchSection returns the SectionDef matching the given heading text, or nil if none match.
// Comparison is performed on a normalized form so that minor variations in
// 中黒 (・) usage and whitespace do not cause false negatives.
func MatchSection(heading string) *SectionDef {
	normHeading := normalizeForMatch(heading)
	for i := range KnownSections {
		for _, name := range KnownSections[i].Names {
			if strings.Contains(normHeading, normalizeForMatch(name)) {
				return &KnownSections[i]
			}
		}
	}
	return nil
}
