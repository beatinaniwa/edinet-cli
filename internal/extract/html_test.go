package extract

import (
	"strings"
	"testing"
)

func TestExtractText_BasicHTML(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/0000000_header.htm": `<html><body>
			<h2>【事業等のリスク】</h2>
			<p>当社グループの事業等のリスクについて記載します。</p>
			<h2>【従業員の状況】</h2>
			<p>従業員数は10,000人です。</p>
		</body></html>`,
	})

	text, err := ExtractText(data)
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}
	if !strings.Contains(text, "事業等のリスク") {
		t.Errorf("text missing '事業等のリスク'")
	}
	if !strings.Contains(text, "従業員数は10,000人") {
		t.Errorf("text missing '従業員数は10,000人'")
	}
}

func TestExtractText_StripScriptAndStyle(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/main.htm": `<html>
			<head><style>body { color: red; }</style></head>
			<body>
				<script>alert('test');</script>
				<p>重要なテキスト</p>
			</body>
		</html>`,
	})

	text, err := ExtractText(data)
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}
	if strings.Contains(text, "color: red") {
		t.Error("text should not contain style content")
	}
	if strings.Contains(text, "alert") {
		t.Error("text should not contain script content")
	}
	if !strings.Contains(text, "重要なテキスト") {
		t.Error("text missing '重要なテキスト'")
	}
}

func TestExtractText_InlineXBRLTags(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/main.htm": `<html><body>
			<ix:nonNumeric contextRef="CurrentYearDuration" name="jpcrp_cor:BusinessRisksTextBlock">
				リスク情報のテキスト
			</ix:nonNumeric>
		</body></html>`,
	})

	text, err := ExtractText(data)
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}
	if !strings.Contains(text, "リスク情報のテキスト") {
		t.Error("text should preserve content inside XBRL inline tags")
	}
}

func TestExtractText_NoPublicDoc(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"AttachDoc/readme.txt": "some attachment",
	})

	_, err := ExtractText(data)
	if err == nil {
		t.Fatal("ExtractText() should fail when PublicDoc/ is missing")
	}
}

func TestExtractSections_KnownSections(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/main.htm": `<html><body>
			<h2>【事業等のリスク】</h2>
			<p>リスクについての説明です。新たなリスク要因が追加されました。</p>
			<h2>【従業員の状況】</h2>
			<p>従業員数は10,000人です。平均年齢は40歳です。</p>
		</body></html>`,
	})

	sections, err := ExtractSections(data)
	if err != nil {
		t.Fatalf("ExtractSections() error = %v", err)
	}
	if len(sections) < 2 {
		t.Fatalf("len(sections) = %d, want >= 2", len(sections))
	}

	found := false
	for _, s := range sections {
		if s.ID == "risk" {
			found = true
			if !strings.Contains(s.Text, "リスクについての説明") {
				t.Errorf("risk section text = %q, missing expected content", s.Text)
			}
		}
	}
	if !found {
		t.Error("missing 'risk' section")
	}
}

func TestExtractText_MultipleHTMLFiles(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/0000001_main.htm":   "<html><body><p>第一章の内容</p></body></html>",
		"PublicDoc/0000002_detail.htm": "<html><body><p>第二章の内容</p></body></html>",
	})

	text, err := ExtractText(data)
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}
	if !strings.Contains(text, "第一章の内容") {
		t.Error("text missing first file content")
	}
	if !strings.Contains(text, "第二章の内容") {
		t.Error("text missing second file content")
	}
}

func TestMatchSection_KnownHeading(t *testing.T) {
	sec := MatchSection("【事業等のリスク】")
	if sec == nil {
		t.Fatal("MatchSection should match '事業等のリスク'")
	}
	if sec.ID != "risk" {
		t.Errorf("ID = %q, want %q", sec.ID, "risk")
	}
}

func TestMatchSection_Unknown(t *testing.T) {
	sec := MatchSection("不明なセクション")
	if sec != nil {
		t.Errorf("MatchSection should return nil for unknown heading, got %v", sec)
	}
}

// TestMatchSection_NakaguroVariants confirms tolerant matching across the
// 中黒 (・) variations seen in real EDINET filings. e.g., 株式会社セブン＆
// アイ・ホールディングス 第20期 (S100VT7P) uses「コーポレートガバナンス」
// (no middle dot) while 日本マクドナルドホールディングス 第55期 (S100XS22)
// uses「コーポレート・ガバナンス」 (with middle dot). Both must map to the
// governance section.
func TestMatchSection_NakaguroVariants(t *testing.T) {
	cases := []struct {
		heading string
		wantID  string
	}{
		{"４【コーポレート・ガバナンスの状況等】", "governance"},
		{"４【コーポレートガバナンスの状況等】", "governance"},
		{"（１）【コーポレート・ガバナンスの概要】", "governance"},
		{"（１）【コーポレートガバナンスの概要】", "governance"},
		{"４【コーポレート ガバナンスの状況等】", "governance"}, // half-width space
		{"４【コーポレート　ガバナンスの状況等】", "governance"}, // full-width space
	}
	for _, c := range cases {
		got := MatchSection(c.heading)
		if got == nil {
			t.Errorf("MatchSection(%q) = nil, want id=%q", c.heading, c.wantID)
			continue
		}
		if got.ID != c.wantID {
			t.Errorf("MatchSection(%q).ID = %q, want %q", c.heading, got.ID, c.wantID)
		}
	}
}

// TestExtractSections_BleedAcrossUnknownHeadings reproduces the bleed-truncated
// pattern observed in EDINET filings such as docID S100XS22 (日本マクドナルド
// HD 第55期): the "従業員の状況" section is followed by unknown headings
// (関係会社の状況, 第２【事業の状況】, 経営方針, サステナビリティ) before the
// next recognised heading (事業等のリスク). The depth-aware walker should flush
// the employees section at the next heading whose h-level is the same as the
// opening heading of employees, keeping employees content from bleeding into
// the following chapters.
func TestExtractSections_BleedAcrossUnknownHeadings(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/main.htm": `<html><body>
			<h3>５【従業員の状況】</h3>
			<p>従業員数は2,454名です。</p>
			<h3>４【関係会社の状況】</h3>
			<p>関係会社の説明です。</p>
			<h2>第２【事業の状況】</h2>
			<h3>１【経営方針、経営環境及び対処すべき課題等】</h3>
			<p>経営方針の説明です。</p>
			<h3>２【サステナビリティに関する考え方及び取組】</h3>
			<p>サステナビリティの説明です。</p>
			<h3>３【事業等のリスク】</h3>
			<p>リスクの説明です。</p>
		</body></html>`,
	})

	sections, err := ExtractSections(data)
	if err != nil {
		t.Fatalf("ExtractSections() error = %v", err)
	}

	var employees, risk *Section
	for i := range sections {
		s := &sections[i]
		if s.ID == "employees" {
			employees = s
		}
		if s.ID == "risk" {
			risk = s
		}
	}

	if employees == nil {
		t.Fatal("missing 'employees' section")
	}
	if !strings.Contains(employees.Text, "従業員数は2,454名") {
		t.Errorf("employees.Text = %q, missing expected content", employees.Text)
	}
	for _, leak := range []string{"関係会社の説明", "経営方針の説明", "サステナビリティの説明", "リスクの説明"} {
		if strings.Contains(employees.Text, leak) {
			t.Errorf("employees.Text bled into other chapter (found %q)", leak)
		}
	}
	if risk == nil {
		t.Fatal("missing 'risk' section")
	}
	if !strings.Contains(risk.Text, "リスクの説明") {
		t.Errorf("risk.Text = %q, missing expected content", risk.Text)
	}
}

// TestExtractSections_SameIDNestedHeading reproduces the empty-section pattern
// observed for governance: a parent heading "コーポレート・ガバナンスの状況等"
// (h3) is immediately followed by a child heading "コーポレート・ガバナンスの
// 概要" (h4) — both match the governance KnownSections entry. The previous
// implementation flushed and reset the section on the second match, leaving
// the parent section empty. With the same-ID continuation rule, the deeper
// heading is treated as a sub-heading inside the open governance section.
func TestExtractSections_SameIDNestedHeading(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/main.htm": `<html><body>
			<h3>４【コーポレート・ガバナンスの状況等】</h3>
			<h4>（１）【コーポレート・ガバナンスの概要】</h4>
			<p>ガバナンスの概要本文です。</p>
			<h4>（２）【役員の状況】</h4>
			<p>役員の状況本文です。</p>
			<h4>（３）【監査の状況】</h4>
			<p>監査の状況本文です。</p>
			<h3>５【提出会社の株式事務の概要】</h3>
			<p>株式事務の概要本文です。</p>
		</body></html>`,
	})

	sections, err := ExtractSections(data)
	if err != nil {
		t.Fatalf("ExtractSections() error = %v", err)
	}

	var governance *Section
	for i := range sections {
		s := &sections[i]
		if s.ID == "governance" {
			governance = s
		}
	}

	if governance == nil {
		t.Fatal("missing 'governance' section")
	}
	for _, want := range []string{"ガバナンスの概要本文", "役員の状況本文", "監査の状況本文"} {
		if !strings.Contains(governance.Text, want) {
			t.Errorf("governance.Text missing expected content %q (text=%q)", want, governance.Text)
		}
	}
	if strings.Contains(governance.Text, "株式事務の概要本文") {
		t.Errorf("governance.Text bled into next chapter (株式事務の概要)")
	}
}

// TestExtractSections_SubHeadingSameHTag reproduces the セコム 第64期
// (S100W3TS) pattern: the chapter heading "３【事業等のリスク】" and the
// sub-section heading "(1)事業環境に起因するリスク" are both marked up with
// <h3 class="smt_head2"> in the source HTML. Pure depth-based flushing would
// close risk at the first sub-heading and drop its content; the chapter-vs-
// sub-numbering check must keep them together.
func TestExtractSections_SubHeadingSameHTag(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/main.htm": `<html><body>
			<h3>３【事業等のリスク】</h3>
			<p>リスクのイントロ段落です。</p>
			<h3>(1)事業環境に起因するリスク</h3>
			<p>事業環境リスクの本文です。</p>
			<h3>①社会・経済</h3>
			<p>社会・経済の説明です。</p>
			<h3>(2)経営戦略に関するリスク</h3>
			<p>経営戦略リスクの本文です。</p>
			<h3>４【経営者による財政状態、経営成績及びキャッシュ・フローの状況の分析】</h3>
			<p>MD&A の本文です。</p>
		</body></html>`,
	})

	sections, err := ExtractSections(data)
	if err != nil {
		t.Fatalf("ExtractSections() error = %v", err)
	}

	var risk, mda *Section
	for i := range sections {
		s := &sections[i]
		if s.ID == "risk" {
			risk = s
		}
		if s.ID == "mda" {
			mda = s
		}
	}

	if risk == nil {
		t.Fatal("missing 'risk' section")
	}
	for _, want := range []string{"リスクのイントロ段落", "事業環境リスクの本文", "社会・経済の説明", "経営戦略リスクの本文"} {
		if !strings.Contains(risk.Text, want) {
			t.Errorf("risk.Text missing %q (text=%q)", want, risk.Text)
		}
	}
	if strings.Contains(risk.Text, "MD&A の本文") {
		t.Errorf("risk.Text bled into MD&A chapter (text=%q)", risk.Text)
	}
	if mda == nil {
		t.Fatal("missing 'mda' section")
	}
	if !strings.Contains(mda.Text, "MD&A の本文") {
		t.Errorf("mda.Text missing expected content (text=%q)", mda.Text)
	}
}

// TestIsChapterHeading covers the chapter-vs-sub-numbering predicate.
func TestIsChapterHeading(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"４【コーポレート・ガバナンスの状況等】", true},
		{"第２【事業の状況】", true},
		{"２【沿革】", true},
		{"５ 【従業員の状況】", true},
		{"第４ 【提出会社の状況】", true},
		{"(1)事業環境に起因するリスク", false},
		{"（１）【コーポレート・ガバナンスの概要】", false},
		{"①社会・経済", false},
		{"a. 受注実績", false},
		{"②キャッシュ・フローの状況", false},
	}
	for _, c := range cases {
		got := isChapterHeading(c.in)
		if got != c.want {
			t.Errorf("isChapterHeading(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestStripFilingFooter checks the EDINET filing-title footer is removed.
func TestStripFilingFooter(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"ガバナンスの本文です。 有価証券報告書（通常方式）_20260326141712", "ガバナンスの本文です。"},
		{"テキスト 有価証券報告書（通常方式）_20250523094900\n", "テキスト"},
		{"footerなし", "footerなし"},
		{"", ""},
	}
	for _, c := range cases {
		got := stripFilingFooter(c.in)
		if got != c.want {
			t.Errorf("stripFilingFooter(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestExtractSections_FilingFooterStripped confirms the filing-title footer
// artifact at the end of an HTML file is not retained in the last section's
// body text.
func TestExtractSections_FilingFooterStripped(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/main.htm": `<html><body>
			<h3>４【コーポレート・ガバナンスの状況等】</h3>
			<p>ガバナンスの本文です。</p>
			<p>有価証券報告書（通常方式）_20260326141712</p>
		</body></html>`,
	})

	sections, err := ExtractSections(data)
	if err != nil {
		t.Fatalf("ExtractSections() error = %v", err)
	}
	var governance *Section
	for i := range sections {
		if sections[i].ID == "governance" {
			governance = &sections[i]
		}
	}
	if governance == nil {
		t.Fatal("missing 'governance' section")
	}
	if strings.Contains(governance.Text, "有価証券報告書（通常方式）_") {
		t.Errorf("governance.Text still contains filing footer artifact (text=%q)", governance.Text)
	}
	if !strings.Contains(governance.Text, "ガバナンスの本文") {
		t.Errorf("governance.Text missing expected content (text=%q)", governance.Text)
	}
}

// TestMergeAdjacentSameIDSections checks the merge safety net directly.
func TestMergeAdjacentSameIDSections(t *testing.T) {
	in := []Section{
		{ID: "governance", Name: "コーポレート・ガバナンスの状況等", Text: ""},
		{ID: "governance", Name: "コーポレート・ガバナンスの概要", Text: "ガバナンス本文"},
		{ID: "financial", Name: "連結財務諸表", Text: "財務諸表本文"},
	}
	out := mergeAdjacentSameIDSections(in)
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2 (governance merged + financial)", len(out))
	}
	if out[0].ID != "governance" {
		t.Errorf("out[0].ID = %q, want governance", out[0].ID)
	}
	if !strings.Contains(out[0].Text, "ガバナンス本文") {
		t.Errorf("merged governance.Text = %q, missing content", out[0].Text)
	}
}
