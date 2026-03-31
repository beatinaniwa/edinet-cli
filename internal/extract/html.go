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

func extractSectionsFromNodes(nodes []*html.Node) []Section {
	var sections []Section
	var currentSection *Section
	var currentText strings.Builder

	for _, doc := range nodes {
		walkForSections(doc, &sections, &currentSection, &currentText)
	}

	// Flush last section
	if currentSection != nil {
		currentSection.Text = normalizeWhitespace(currentText.String())
		sections = append(sections, *currentSection)
	}

	return sections
}

func walkForSections(n *html.Node, sections *[]Section, current **Section, text *strings.Builder) {
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
		if secDef := MatchSection(headingText); secDef != nil {
			// Flush previous section
			if *current != nil {
				(*current).Text = normalizeWhitespace(text.String())
				*sections = append(*sections, **current)
			}
			*current = &Section{
				ID:   secDef.ID,
				Name: headingText,
			}
			text.Reset()
			return
		}
	}

	// Collect text for current section
	if n.Type == html.TextNode && *current != nil {
		t := strings.TrimSpace(n.Data)
		if t != "" {
			text.WriteString(t)
			text.WriteString(" ")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkForSections(c, sections, current, text)
	}
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
