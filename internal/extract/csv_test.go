package extract

import (
	"testing"
)

func TestExtractCSVData_Normal(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"XBRL_TO_CSV/jpcrp030000-asr-001_E02144-000_2025-03-31_01_2025-06-20.csv": "要素ID,コンテキストID,値\njpcrp_cor:Revenue,CurrentYearDuration,1000000\njpcrp_cor:NetIncome,CurrentYearDuration,500000",
	})

	result, err := ExtractCSVData(data)
	if err != nil {
		t.Fatalf("ExtractCSVData() error = %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("len(Files) = %d, want 1", len(result.Files))
	}
	file := result.Files[0]
	if len(file.Headers) != 3 {
		t.Fatalf("len(Headers) = %d, want 3", len(file.Headers))
	}
	if file.Headers[0] != "要素ID" {
		t.Errorf("Headers[0] = %q, want %q", file.Headers[0], "要素ID")
	}
	if len(file.Rows) != 2 {
		t.Fatalf("len(Rows) = %d, want 2", len(file.Rows))
	}
	if file.Rows[0][0] != "jpcrp_cor:Revenue" {
		t.Errorf("Rows[0][0] = %q, want %q", file.Rows[0][0], "jpcrp_cor:Revenue")
	}
}

func TestExtractCSVData_MultipleFiles(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"XBRL_TO_CSV/file1.csv": "a,b\n1,2",
		"XBRL_TO_CSV/file2.csv": "c,d\n3,4",
	})

	result, err := ExtractCSVData(data)
	if err != nil {
		t.Fatalf("ExtractCSVData() error = %v", err)
	}
	if len(result.Files) != 2 {
		t.Errorf("len(Files) = %d, want 2", len(result.Files))
	}
}

func TestExtractCSVData_NoCSVDirectory(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"PublicDoc/main.htm": "<html>test</html>",
	})

	_, err := ExtractCSVData(data)
	if err == nil {
		t.Fatal("ExtractCSVData() should fail when XBRL_TO_CSV/ is missing")
	}
}

func TestExtractCSVData_UnevenRows(t *testing.T) {
	data := createTestZip(t, map[string]string{
		"XBRL_TO_CSV/data.csv": "a,b,c\n1,2\n3,4,5,6",
	})

	result, err := ExtractCSVData(data)
	if err != nil {
		t.Fatalf("ExtractCSVData() error = %v", err)
	}
	// Should still parse — uneven rows are preserved as-is
	if len(result.Files[0].Rows) != 2 {
		t.Errorf("len(Rows) = %d, want 2", len(result.Files[0].Rows))
	}
}
