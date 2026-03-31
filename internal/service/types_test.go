package service

import (
	"testing"

	"github.com/beatinaniwa/edinet-cli/internal/api"
)

func strPtr(s string) *string { return &s }

func TestToDocumentInfo_FullDocument(t *testing.T) {
	doc := api.Document{
		SeqNumber:      1,
		DocID:          "S100ABCD",
		EdinetCode:     strPtr("E02144"),
		SecCode:        strPtr("72030"),
		FilerName:      strPtr("トヨタ自動車株式会社"),
		DocTypeCode:    strPtr("120"),
		DocDescription: strPtr("有価証券報告書"),
		PeriodStart:    strPtr("2024-04-01"),
		PeriodEnd:      strPtr("2025-03-31"),
		SubmitDateTime: strPtr("2025-06-20 12:34"),
		XbrlFlag:       strPtr("1"),
		PdfFlag:        strPtr("1"),
		CsvFlag:        strPtr("1"),
		LegalStatus:    strPtr("1"),
	}
	info := ToDocumentInfo(doc)

	if info.DocID != "S100ABCD" {
		t.Errorf("DocID = %q, want %q", info.DocID, "S100ABCD")
	}
	if info.EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %q, want %q", info.EdinetCode, "E02144")
	}
	if info.SecCode != "72030" {
		t.Errorf("SecCode = %q, want %q", info.SecCode, "72030")
	}
	if info.FilerName != "トヨタ自動車株式会社" {
		t.Errorf("FilerName = %q, want %q", info.FilerName, "トヨタ自動車株式会社")
	}
	if !info.HasXbrl {
		t.Error("HasXbrl = false, want true")
	}
	if !info.HasPdf {
		t.Error("HasPdf = false, want true")
	}
	if !info.HasCsv {
		t.Error("HasCsv = false, want true")
	}
	if info.LegalStatus != "active" {
		t.Errorf("LegalStatus = %q, want %q", info.LegalStatus, "active")
	}
}

func TestToDocumentInfo_NullFields(t *testing.T) {
	doc := api.Document{
		DocID: "S100XXXX",
		// All pointer fields are nil (expired document)
	}
	info := ToDocumentInfo(doc)

	if info.DocID != "S100XXXX" {
		t.Errorf("DocID = %q, want %q", info.DocID, "S100XXXX")
	}
	if info.EdinetCode != "" {
		t.Errorf("EdinetCode = %q, want empty", info.EdinetCode)
	}
	if info.HasXbrl {
		t.Error("HasXbrl = true, want false")
	}
	if info.LegalStatus != "" {
		t.Errorf("LegalStatus = %q, want empty", info.LegalStatus)
	}
}

func TestToDocumentInfo_LegalStatusMapping(t *testing.T) {
	tests := []struct {
		input *string
		want  string
	}{
		{strPtr("1"), "active"},
		{strPtr("2"), "extended"},
		{strPtr("0"), "expired"},
		{strPtr("9"), "9"},
		{nil, ""},
	}
	for _, tt := range tests {
		doc := api.Document{DocID: "S1", LegalStatus: tt.input}
		info := ToDocumentInfo(doc)
		if info.LegalStatus != tt.want {
			inputStr := "<nil>"
			if tt.input != nil {
				inputStr = *tt.input
			}
			t.Errorf("LegalStatus(%s) = %q, want %q", inputStr, info.LegalStatus, tt.want)
		}
	}
}

func TestToDocumentInfo_FlagConversion(t *testing.T) {
	tests := []struct {
		flag *string
		want bool
	}{
		{strPtr("1"), true},
		{strPtr("0"), false},
		{nil, false},
	}
	for _, tt := range tests {
		doc := api.Document{DocID: "S1", XbrlFlag: tt.flag}
		info := ToDocumentInfo(doc)
		if info.HasXbrl != tt.want {
			t.Errorf("HasXbrl(%v) = %v, want %v", tt.flag, info.HasXbrl, tt.want)
		}
	}
}
