package company

import (
	"os"
	"testing"
)

func loadTestRegistry(t *testing.T) *Registry {
	t.Helper()
	data, err := os.ReadFile("../../testdata/edinetcode_sample.csv")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	r := &Registry{}
	if err := r.LoadFromCSV(data); err != nil {
		t.Fatalf("LoadFromCSV() error = %v", err)
	}
	return r
}

func TestSearch_ByEdinetCode(t *testing.T) {
	r := loadTestRegistry(t)
	results := r.Search("E02144")
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].SubmitterName != "トヨタ自動車株式会社" {
		t.Errorf("SubmitterName = %q, want トヨタ自動車株式会社", results[0].SubmitterName)
	}
}

func TestSearch_BySecCode5Digit(t *testing.T) {
	r := loadTestRegistry(t)
	results := r.Search("72030")
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %q, want E02144", results[0].EdinetCode)
	}
}

func TestSearch_BySecCode4Digit(t *testing.T) {
	r := loadTestRegistry(t)
	results := r.Search("7203")
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %q, want E02144", results[0].EdinetCode)
	}
}

func TestSearch_ByJapaneseName(t *testing.T) {
	r := loadTestRegistry(t)
	results := r.Search("トヨタ")
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %q, want E02144", results[0].EdinetCode)
	}
}

func TestSearch_ByEnglishName(t *testing.T) {
	r := loadTestRegistry(t)
	results := r.Search("TOYOTA")
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
}

func TestSearch_ByKanaName(t *testing.T) {
	r := loadTestRegistry(t)
	results := r.Search("ソニー")
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].EdinetCode != "E01777" {
		t.Errorf("EdinetCode = %q, want E01777", results[0].EdinetCode)
	}
}

func TestSearch_NoMatch(t *testing.T) {
	r := loadTestRegistry(t)
	results := r.Search("存在しない会社")
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}

func TestSearch_MultipleResults(t *testing.T) {
	r := loadTestRegistry(t)
	// "株式会社" is in all company names
	results := r.Search("株式会社")
	if len(results) < 3 {
		t.Errorf("len(results) = %d, want >= 3", len(results))
	}
}

func TestSearch_ByIndustry(t *testing.T) {
	r := loadTestRegistry(t)
	results := r.SearchByIndustry("輸送用機器")
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2 (Toyota + Honda)", len(results))
	}
}
