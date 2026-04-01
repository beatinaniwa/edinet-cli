package financial

import "testing"

func assertMetricNil(t *testing.T, s Summary, key string) {
	t.Helper()
	if v := s[key]; v != nil {
		t.Errorf("Summary[%q] = %v, want nil", key, *v)
	}
}

func TestDeriveMetrics_AllMetrics(t *testing.T) {
	s := Summary{
		"revenue":              ptrFloat(10000),
		"gross_profit":         ptrFloat(5000),
		"operating_income":     ptrFloat(2000),
		"net_income":           ptrFloat(1000),
		"equity":               ptrFloat(8000),
		"total_assets":         ptrFloat(20000),
		"current_assets":       ptrFloat(6000),
		"current_liabilities":  ptrFloat(3000),
		"operating_cf":         ptrFloat(3000),
		"investing_cf":         ptrFloat(-1000),
		"interest_bearing_debt": ptrFloat(4000),
	}

	DeriveMetrics(s)

	assertSummaryValue(t, s, "gross_margin", 0.5)
	assertSummaryValue(t, s, "operating_margin", 0.2)
	assertSummaryValue(t, s, "net_margin", 0.1)
	assertSummaryValue(t, s, "roe", 0.125)          // 1000/8000
	assertSummaryValue(t, s, "roa", 0.05)           // 1000/20000
	assertSummaryValue(t, s, "equity_ratio", 0.4)   // 8000/20000
	assertSummaryValue(t, s, "current_ratio", 2.0)  // 6000/3000
	assertSummaryValue(t, s, "fcf", 2000)           // 3000 + (-1000)
	assertSummaryValue(t, s, "debt_to_equity", 0.5) // 4000/8000
}

func TestDeriveMetrics_MissingPrerequisites(t *testing.T) {
	s := Summary{
		"revenue": ptrFloat(10000),
		// no other keys
	}

	DeriveMetrics(s)

	// Only gross_margin should be skipped (no gross_profit)
	assertMetricNil(t, s, "gross_margin")
	assertMetricNil(t, s, "roe")
	assertMetricNil(t, s, "roa")
	assertMetricNil(t, s, "fcf")
}

func TestDeriveMetrics_ZeroDenominator(t *testing.T) {
	s := Summary{
		"revenue":             ptrFloat(0),
		"gross_profit":        ptrFloat(0),
		"net_income":          ptrFloat(1000),
		"equity":              ptrFloat(0),
		"total_assets":        ptrFloat(0),
		"current_liabilities": ptrFloat(0),
	}

	DeriveMetrics(s) // should not panic

	assertMetricNil(t, s, "gross_margin")
	assertMetricNil(t, s, "roe")
	assertMetricNil(t, s, "roa")
	assertMetricNil(t, s, "current_ratio")
}

func TestDeriveMetrics_NegativeValues(t *testing.T) {
	s := Summary{
		"revenue":    ptrFloat(10000),
		"net_income": ptrFloat(-500),
		"equity":     ptrFloat(8000),
	}

	DeriveMetrics(s)

	assertSummaryValue(t, s, "net_margin", -0.05) // -500/10000
	assertSummaryValue(t, s, "roe", -0.0625)      // -500/8000
}

func TestDeriveMetrics_DoesNotOverwrite(t *testing.T) {
	s := Summary{
		"revenue":          ptrFloat(10000),
		"operating_income": ptrFloat(2000),
		"operating_margin": ptrFloat(0.999), // pre-existing
	}

	DeriveMetrics(s)

	assertSummaryValue(t, s, "operating_margin", 0.999) // should NOT be overwritten
}

func TestDeriveMetrics_EmptySummary(t *testing.T) {
	s := Summary{}
	DeriveMetrics(s) // should not panic
	if len(s) != 0 {
		t.Errorf("empty summary should remain empty, got %d keys", len(s))
	}
}

func TestDeriveMetrics_EquityFallbackToNetAssets(t *testing.T) {
	s := Summary{
		"net_income":  ptrFloat(1000),
		"net_assets":  ptrFloat(5000),
		"total_assets": ptrFloat(20000),
		// no "equity" key
	}

	DeriveMetrics(s)

	assertSummaryValue(t, s, "roe", 0.2)           // 1000/5000 (net_assets fallback)
	assertSummaryValue(t, s, "equity_ratio", 0.25) // 5000/20000
}

func TestDeriveMetrics_EquityPreferredOverNetAssets(t *testing.T) {
	s := Summary{
		"net_income":  ptrFloat(1000),
		"equity":      ptrFloat(8000),
		"net_assets":  ptrFloat(10000), // should be ignored
		"total_assets": ptrFloat(20000),
	}

	DeriveMetrics(s)

	assertSummaryValue(t, s, "roe", 0.125)        // 1000/8000 (equity preferred)
	assertSummaryValue(t, s, "equity_ratio", 0.4) // 8000/20000
}

func TestDerivedMetrics_MetadataCompleteness(t *testing.T) {
	defs := DerivedMetricDefs()
	if len(defs) == 0 {
		t.Fatal("DerivedMetricDefs() returned empty slice")
	}
	for _, d := range defs {
		if d.Key == "" {
			t.Error("DerivedMetricDef has empty Key")
		}
		if d.Formula == "" {
			t.Errorf("DerivedMetricDef %q has empty Formula", d.Key)
		}
		if d.Description == "" {
			t.Errorf("DerivedMetricDef %q has empty Description", d.Key)
		}
		if d.Requires == nil {
			t.Errorf("DerivedMetricDef %q has nil Requires", d.Key)
		}
	}
}
