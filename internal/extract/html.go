package extract

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// ExtractText extracts plain text from HTML files in a type=1 XBRL ZIP archive.
// Reads all .htm files under PublicDoc/, sorted by name, and concatenates their text.
func ExtractText(zipData []byte) (string, error) {
	entries, err := readHTMLEntries(zipData)
	if err != nil {
		return "", err
	}

	var allText strings.Builder
	for i, entry := range entries {
		if i > 0 {
			allText.WriteString("\n\n")
		}
		text, err := extractHTMLText(entry.Data)
		if err != nil {
			// Best-effort: log error but continue
			fmt.Fprintf(&allText, "[parse error: %s: %v]\n", entry.Name, err)
			continue
		}
		allText.WriteString(text)
	}

	return normalizeWhitespace(allText.String()), nil
}

// ExtractSections extracts named sections from HTML files in a type=1 XBRL ZIP.
// Sections are detected by heading elements (h1-h3) matching known section names.
func ExtractSections(zipData []byte) ([]Section, error) {
	entries, err := readHTMLEntries(zipData)
	if err != nil {
		return nil, err
	}

	var allNodes []*html.Node
	for _, entry := range entries {
		doc, err := html.Parse(bytes.NewReader(entry.Data))
		if err != nil {
			continue
		}
		allNodes = append(allNodes, doc)
	}

	return extractSectionsFromNodes(allNodes), nil
}

func extractHTMLText(data []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	extractTextFromNode(doc, &buf)
	return buf.String(), nil
}

func extractTextFromNode(n *html.Node, buf *strings.Builder) {
	if n == nil {
		return
	}

	// Skip style and script elements
	if n.Type == html.ElementNode && (n.Data == "style" || n.Data == "script") {
		return
	}

	// Add text content
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			buf.WriteString(text)
			buf.WriteString(" ")
		}
	}

	// Add line break for block elements
	if n.Type == html.ElementNode && isBlockElement(n.Data) {
		buf.WriteString("\n")
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractTextFromNode(c, buf)
	}

	if n.Type == html.ElementNode && isBlockElement(n.Data) {
		buf.WriteString("\n")
	}
}

// sectionWalkState tracks the currently-open section and the heading depth
// (h-level: 1..6) at which it was opened. Depth lets us flush on non-matching
// sibling/parent headings while keeping sub-headings (deeper h-levels) as part
// of the open section.
type sectionWalkState struct {
	current *Section
	depth   int // h-level (1..6) where the current section was opened; 0 = no section open
}

func extractSectionsFromNodes(nodes []*html.Node) []Section {
	var sections []Section
	state := &sectionWalkState{}
	var currentText strings.Builder

	for _, doc := range nodes {
		walkForSections(doc, &sections, state, &currentText)
	}

	// Flush last section
	if state.current != nil {
		state.current.Text = normalizeWhitespace(currentText.String())
		sections = append(sections, *state.current)
	}

	return mergeAdjacentSameIDSections(sections)
}

// headingLevel returns the heading depth (1..6) for h1..h6 tags, or 0 otherwise.
func headingLevel(tag string) int {
	switch tag {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	}
	return 0
}

func walkForSections(n *html.Node, sections *[]Section, state *sectionWalkState, text *strings.Builder) {
	if n == nil {
		return
	}

	// Skip style and script
	if n.Type == html.ElementNode && (n.Data == "style" || n.Data == "script") {
		return
	}

	// Check if this is a heading element
	if n.Type == html.ElementNode && isHeadingElement(n.Data) {
		headingText := getNodeText(n)
		level := headingLevel(n.Data)

		if secDef := MatchSection(headingText); secDef != nil {
			// If the current section has the same ID, treat this heading as a
			// sub-heading inside the same section: do not flush, do not reset
			// text. This handles EDINET filings where a parent heading like
			// "コーポレート・ガバナンスの状況等" is immediately followed by
			// child headings like "コーポレート・ガバナンスの概要" that also
			// match the same KnownSections entry.
			if state.current != nil && state.current.ID == secDef.ID {
				return
			}

			// Different section: flush previous and start new.
			if state.current != nil {
				state.current.Text = normalizeWhitespace(text.String())
				*sections = append(*sections, *state.current)
			}
			state.current = &Section{
				ID:   secDef.ID,
				Name: headingText,
			}
			state.depth = level
			text.Reset()
			return
		}

		// Non-matching heading: if it is at the same depth as (or shallower
		// than) the heading that opened the current section, treat it as a
		// section boundary and flush. Deeper headings are sub-headings of the
		// current section (e.g., "（２）役員の状況" inside a governance section
		// anchored at h3) and should keep accumulating text.
		if state.current != nil && level > 0 && level <= state.depth {
			state.current.Text = normalizeWhitespace(text.String())
			*sections = append(*sections, *state.current)
			state.current = nil
			state.depth = 0
			text.Reset()
			// Fall through so the heading's own text is not collected into
			// any section.
			return
		}
	}

	// Collect text for current section
	if n.Type == html.TextNode && state.current != nil {
		t := strings.TrimSpace(n.Data)
		if t != "" {
			text.WriteString(t)
			text.WriteString(" ")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkForSections(c, sections, state, text)
	}
}

// mergeAdjacentSameIDSections concatenates consecutive sections that share
// the same ID. This is a safety net for cases where the same section opens
// twice in a row (e.g., split across HTML files) — the depth-aware walker
// already prevents most occurrences, but merging guards against edge cases.
func mergeAdjacentSameIDSections(in []Section) []Section {
	if len(in) <= 1 {
		return in
	}
	out := make([]Section, 0, len(in))
	out = append(out, in[0])
	for i := 1; i < len(in); i++ {
		last := &out[len(out)-1]
		if last.ID != "" && last.ID == in[i].ID {
			if in[i].Text != "" {
				if last.Text != "" {
					last.Text += " " + in[i].Text
				} else {
					last.Text = in[i].Text
				}
			}
			continue
		}
		out = append(out, in[i])
	}
	return out
}

func getNodeText(n *html.Node) string {
	var buf strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(buf.String())
}

func isHeadingElement(tag string) bool {
	switch tag {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return true
	}
	return false
}

func isBlockElement(tag string) bool {
	switch tag {
	case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6",
		"li", "tr", "td", "th", "br", "hr", "table",
		"section", "article", "header", "footer", "nav":
		return true
	}
	return false
}

// readHTMLEntries finds HTML files under PublicDoc/ at any nesting depth.
// EDINET type=1 ZIPs may place HTML under PublicDoc/, XBRL/PublicDoc/, or deeper paths.
func readHTMLEntries(zipData []byte) ([]ZipEntry, error) {
	entries, err := ReadFromZipFunc(zipData, func(name string) bool {
		return (strings.Contains(name, "PublicDoc/") || strings.Contains(name, "PublicDoc\\")) &&
			(strings.HasSuffix(name, ".htm") || strings.HasSuffix(name, ".html"))
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read HTML from ZIP: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no .htm files found in archive under PublicDoc/")
	}
	return entries, nil
}

func normalizeWhitespace(s string) string {
	// Collapse multiple newlines to max 2
	lines := strings.Split(s, "\n")
	var result []string
	emptyCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			emptyCount++
			if emptyCount <= 1 {
				result = append(result, "")
			}
		} else {
			emptyCount = 0
			result = append(result, trimmed)
		}
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}
