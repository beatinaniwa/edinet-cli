package schema

import (
	"testing"
)

func TestListDocTypes_HasMainTypes(t *testing.T) {
	types := ListDocTypes()
	if len(types) == 0 {
		t.Fatal("ListDocTypes() returned empty")
	}

	// Check for key document types
	wantCodes := map[string]string{
		"120": "有価証券報告書",
		"140": "四半期報告書",
		"160": "半期報告書",
		"180": "臨時報告書",
		"350": "大量保有報告書",
	}
	found := map[string]bool{}
	for _, dt := range types {
		if want, ok := wantCodes[dt.Code]; ok {
			if dt.Name != want {
				t.Errorf("code %s name = %q, want %q", dt.Code, dt.Name, want)
			}
			found[dt.Code] = true
		}
	}
	for code := range wantCodes {
		if !found[code] {
			t.Errorf("missing doc type code %s", code)
		}
	}
}

func TestListCommands_HasAllTopLevel(t *testing.T) {
	cmds := ListCommands()
	if len(cmds) == 0 {
		t.Fatal("ListCommands() returned empty")
	}

	wantNames := []string{"doc list", "doc get", "doc data", "doc text", "company search", "company filings", "company update", "schema commands", "schema doc-types", "schema sections"}
	cmdMap := map[string]bool{}
	for _, c := range cmds {
		cmdMap[c.Name] = true
	}
	for _, name := range wantNames {
		if !cmdMap[name] {
			t.Errorf("missing command %q", name)
		}
	}
}

func TestListSections_HasKnownSections(t *testing.T) {
	sections := ListSections()
	if len(sections) == 0 {
		t.Fatal("ListSections() returned empty")
	}

	wantIDs := []string{"business", "risk", "mda", "employees"}
	secMap := map[string]bool{}
	for _, s := range sections {
		secMap[s.ID] = true
	}
	for _, id := range wantIDs {
		if !secMap[id] {
			t.Errorf("missing section %q", id)
		}
	}
}
