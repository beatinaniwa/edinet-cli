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

// MatchSection returns the SectionDef matching the given heading text, or nil if none match.
func MatchSection(heading string) *SectionDef {
	for i := range KnownSections {
		for _, name := range KnownSections[i].Names {
			if strings.Contains(heading, name) {
				return &KnownSections[i]
			}
		}
	}
	return nil
}
