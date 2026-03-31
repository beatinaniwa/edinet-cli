package company

import (
	"os"
	"testing"
)

func TestLoadFromCSV_Normal(t *testing.T) {
	data, err := os.ReadFile("../../testdata/edinetcode_sample.csv")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	r := &Registry{}
	if err := r.LoadFromCSV(data); err != nil {
		t.Fatalf("LoadFromCSV() error = %v", err)
	}
	if len(r.Entries) != 4 {
		t.Fatalf("len(Entries) = %d, want 4", len(r.Entries))
	}

	toyota := r.Entries[0]
	if toyota.EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %q, want %q", toyota.EdinetCode, "E02144")
	}
	if toyota.SubmitterName != "トヨタ自動車株式会社" {
		t.Errorf("SubmitterName = %q, want %q", toyota.SubmitterName, "トヨタ自動車株式会社")
	}
	if toyota.SecCode != "72030" {
		t.Errorf("SecCode = %q, want %q", toyota.SecCode, "72030")
	}
	if toyota.IndustryName != "輸送用機器" {
		t.Errorf("IndustryName = %q, want %q", toyota.IndustryName, "輸送用機器")
	}
}

func TestLoadFromCSV_HeaderValidation(t *testing.T) {
	data := []byte("タイトル行\n不正ヘッダ1,不正ヘッダ2\ndata1,data2\n")
	r := &Registry{}
	err := r.LoadFromCSV(data)
	if err == nil {
		t.Fatal("LoadFromCSV() should fail with invalid headers")
	}
}

func TestLoadFromCSV_EmptyInput(t *testing.T) {
	r := &Registry{}
	err := r.LoadFromCSV([]byte{})
	if err == nil {
		t.Fatal("LoadFromCSV() should fail with empty input")
	}
}

func TestLoadFromCSV_TitleRowOnly(t *testing.T) {
	r := &Registry{}
	err := r.LoadFromCSV([]byte("タイトル行\n"))
	if err == nil {
		t.Fatal("LoadFromCSV() should fail with only title row")
	}
}

func TestRegistry_FindByEdinetCode(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/edinetcode_sample.csv")
	r := &Registry{}
	_ = r.LoadFromCSV(data)

	entry, err := r.FindByEdinetCode("E02144")
	if err != nil {
		t.Fatalf("FindByEdinetCode() error = %v", err)
	}
	if entry.SubmitterName != "トヨタ自動車株式会社" {
		t.Errorf("SubmitterName = %q, want トヨタ自動車株式会社", entry.SubmitterName)
	}
}

func TestRegistry_FindByEdinetCode_NotFound(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/edinetcode_sample.csv")
	r := &Registry{}
	_ = r.LoadFromCSV(data)

	_, err := r.FindByEdinetCode("E99999")
	if err == nil {
		t.Fatal("FindByEdinetCode() should fail for unknown code")
	}
}

func TestRegistry_FindBySecCode(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/edinetcode_sample.csv")
	r := &Registry{}
	_ = r.LoadFromCSV(data)

	entry, err := r.FindBySecCode("72030")
	if err != nil {
		t.Fatalf("FindBySecCode() error = %v", err)
	}
	if entry.EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %q, want E02144", entry.EdinetCode)
	}
}

func TestLoadFromCSV_ShiftJIS(t *testing.T) {
	data, err := os.ReadFile("../../testdata/edinetcode_sample_sjis.csv")
	if err != nil {
		t.Fatalf("failed to read Shift-JIS fixture: %v", err)
	}

	r := &Registry{}
	if err := r.LoadFromCSV(data); err != nil {
		t.Fatalf("LoadFromCSV(Shift-JIS) error = %v", err)
	}
	if len(r.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(r.Entries))
	}
	if r.Entries[0].EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %q, want E02144", r.Entries[0].EdinetCode)
	}
	if r.Entries[0].SubmitterName != "トヨタ自動車株式会社" {
		t.Errorf("SubmitterName = %q, want トヨタ自動車株式会社", r.Entries[0].SubmitterName)
	}
}

func TestRegistry_FindBySecCode_FourDigit(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/edinetcode_sample.csv")
	r := &Registry{}
	_ = r.LoadFromCSV(data)

	// Search with 4-digit code, matching 5-digit "72030"
	entry, err := r.FindBySecCode("7203")
	if err != nil {
		t.Fatalf("FindBySecCode(4-digit) error = %v", err)
	}
	if entry.EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %q, want E02144", entry.EdinetCode)
	}
}
