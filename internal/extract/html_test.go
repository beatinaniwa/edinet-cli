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
