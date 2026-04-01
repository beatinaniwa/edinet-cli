package financial

import "testing"

func TestParseContextID_CurrentYearInstant(t *testing.T) {
	ctx := ParseContextID("CurrentYearInstant", "連結")
	if ctx.Period != "current" {
		t.Errorf("Period = %q, want %q", ctx.Period, "current")
	}
	if ctx.PointType != "instant" {
		t.Errorf("PointType = %q, want %q", ctx.PointType, "instant")
	}
	if ctx.Consolidated != "consolidated" {
		t.Errorf("Consolidated = %q, want %q", ctx.Consolidated, "consolidated")
	}
	if ctx.Member != "" {
		t.Errorf("Member = %q, want empty", ctx.Member)
	}
}

func TestParseContextID_Prior1YearDuration(t *testing.T) {
	ctx := ParseContextID("Prior1YearDuration", "連結")
	if ctx.Period != "prior1" {
		t.Errorf("Period = %q, want %q", ctx.Period, "prior1")
	}
	if ctx.PointType != "duration" {
		t.Errorf("PointType = %q, want %q", ctx.PointType, "duration")
	}
}

func TestParseContextID_Prior4YearDuration(t *testing.T) {
	ctx := ParseContextID("Prior4YearDuration", "その他")
	if ctx.Period != "prior4" {
		t.Errorf("Period = %q, want %q", ctx.Period, "prior4")
	}
	if ctx.PointType != "duration" {
		t.Errorf("PointType = %q, want %q", ctx.PointType, "duration")
	}
	if ctx.Consolidated != "other" {
		t.Errorf("Consolidated = %q, want %q", ctx.Consolidated, "other")
	}
}

func TestParseContextID_NonConsolidatedMember(t *testing.T) {
	ctx := ParseContextID("CurrentYearInstant_NonConsolidatedMember", "個別")
	if ctx.Period != "current" {
		t.Errorf("Period = %q, want %q", ctx.Period, "current")
	}
	if ctx.PointType != "instant" {
		t.Errorf("PointType = %q, want %q", ctx.PointType, "instant")
	}
	if ctx.Consolidated != "non_consolidated" {
		t.Errorf("Consolidated = %q, want %q", ctx.Consolidated, "non_consolidated")
	}
	if ctx.Member != "" {
		t.Errorf("Member = %q, want empty (NonConsolidatedMember is not a segment)", ctx.Member)
	}
}

func TestParseContextID_SegmentMember(t *testing.T) {
	ctx := ParseContextID("CurrentYearDuration_jpcrp030000-asr_E02144-000AutomotiveReportableSegmentMember", "その他")
	if ctx.Period != "current" {
		t.Errorf("Period = %q, want %q", ctx.Period, "current")
	}
	if ctx.PointType != "duration" {
		t.Errorf("PointType = %q, want %q", ctx.PointType, "duration")
	}
	if ctx.Member != "jpcrp030000-asr_E02144-000AutomotiveReportableSegmentMember" {
		t.Errorf("Member = %q, want segment member", ctx.Member)
	}
}

func TestParseContextID_ConsolidatedColumn(t *testing.T) {
	tests := []struct {
		name           string
		consolidatedCol string
		want           string
	}{
		{"連結", "連結", "consolidated"},
		{"個別", "個別", "non_consolidated"},
		{"その他 (IFRS consolidated)", "その他", "other"},
		{"empty", "", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ParseContextID("CurrentYearInstant", tt.consolidatedCol)
			if ctx.Consolidated != tt.want {
				t.Errorf("Consolidated = %q, want %q", ctx.Consolidated, tt.want)
			}
		})
	}
}

func TestParseContextID_FilingDateInstant(t *testing.T) {
	ctx := ParseContextID("FilingDateInstant", "その他")
	if ctx.Period != "filing_date" {
		t.Errorf("Period = %q, want %q", ctx.Period, "filing_date")
	}
	if ctx.PointType != "instant" {
		t.Errorf("PointType = %q, want %q", ctx.PointType, "instant")
	}
}

func TestParseContextID_Empty(t *testing.T) {
	ctx := ParseContextID("", "")
	if ctx.Period != "" {
		t.Errorf("Period = %q, want empty", ctx.Period)
	}
	if ctx.PointType != "" {
		t.Errorf("PointType = %q, want empty", ctx.PointType)
	}
}

func TestParseContextID_QuarterlyContexts(t *testing.T) {
	tests := []struct {
		contextID  string
		wantPeriod string
		wantType   string
	}{
		{"CurrentQuarterDuration", "current_quarter", "duration"},
		{"CurrentQuarterInstant", "current_quarter", "instant"},
		{"CurrentInterimInstant", "current_interim", "instant"},
		{"CurrentYTDDuration", "current_ytd", "duration"},
		{"Prior1QuarterDuration", "prior1_quarter", "duration"},
		{"Prior1InterimInstant", "prior1_interim", "instant"},
		{"Prior1YTDDuration", "prior1_ytd", "duration"},
	}
	for _, tt := range tests {
		t.Run(tt.contextID, func(t *testing.T) {
			ctx := ParseContextID(tt.contextID, "連結")
			if ctx.Period != tt.wantPeriod {
				t.Errorf("Period = %q, want %q", ctx.Period, tt.wantPeriod)
			}
			if ctx.PointType != tt.wantType {
				t.Errorf("PointType = %q, want %q", ctx.PointType, tt.wantType)
			}
		})
	}
}

func TestParseContextID_EquityMember(t *testing.T) {
	// Context IDs like CurrentYearDuration_RetainedEarningsIFRSMember
	ctx := ParseContextID("CurrentYearDuration_RetainedEarningsIFRSMember", "その他")
	if ctx.Period != "current" {
		t.Errorf("Period = %q, want %q", ctx.Period, "current")
	}
	if ctx.Member != "RetainedEarningsIFRSMember" {
		t.Errorf("Member = %q, want %q", ctx.Member, "RetainedEarningsIFRSMember")
	}
}
