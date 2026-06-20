// Package mail provides SMTP email delivery utilities including
// a simple Markdown-to-HTML renderer for email templates.
package mail

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// RenderMarkdown converts a simple Markdown string to HTML.
// Supported syntax: headings (#), bold (**), italic (*), links, inline code,
// fenced code blocks, unordered lists, and blank-line separated paragraphs.
func RenderMarkdown(md string) string {
	md = strings.TrimSpace(md)
	if md == "" {
		return ""
	}

	// Split into blocks separated by blank lines
	blocks := splitBlocks(md)

	var htmlParts []string
	for _, block := range blocks {
		htmlParts = append(htmlParts, renderBlock(strings.TrimSpace(block)))
	}

	return strings.Join(htmlParts, "\n")
}

// RenderTemplate renders a markdown template with variable substitution.
// Variables in the template use {{.VarName}} syntax.
func RenderTemplate(md string, vars map[string]any) string {
	content := md
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		content = strings.ReplaceAll(content, placeholder, fmt.Sprint(value))
	}
	return RenderMarkdown(content)
}

func splitBlocks(md string) []string {
	lines := strings.Split(md, "\n")
	var blocks []string
	var current []string

	flush := func() {
		if len(current) > 0 {
			blocks = append(blocks, strings.Join(current, "\n"))
			current = nil
		}
	}

	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track fenced code blocks
		if strings.HasPrefix(trimmed, "```") {
			if inFence {
				current = append(current, line)
				flush()
				inFence = false
				continue
			}
			flush()
			current = append(current, line)
			inFence = true
			continue
		}

		if inFence {
			current = append(current, line)
			continue
		}

		if trimmed == "" {
			flush()
			continue
		}

		current = append(current, line)
	}
	flush()

	return blocks
}

func renderBlock(block string) string {
	// Fenced code block
	if strings.HasPrefix(block, "```") {
		return renderCodeBlock(block)
	}

	// Heading
	if strings.HasPrefix(block, "#") {
		return renderHeading(block)
	}

	// Unordered list
	if isListItem(block) {
		return renderList(block)
	}

	// Regular paragraph
	return renderParagraph(block)
}

func renderHeading(block string) string {
	level := 0
	for i, ch := range block {
		if ch == '#' {
			level++
		} else {
			break
		}
		if i >= 6 {
			break
		}
	}
	if level == 0 || level > 6 {
		level = 1
	}
	text := strings.TrimSpace(block[level:])
	if level > 6 {
		level = 6
	}
	return fmt.Sprintf("<h%d>%s</h%d>", level, renderInline(text), level)
}

func renderCodeBlock(block string) string {
	lines := strings.Split(block, "\n")
	if len(lines) < 2 {
		return ""
	}
	// Remove first and last ``` lines
	codeLines := lines[1 : len(lines)-1]
	code := html.EscapeString(strings.Join(codeLines, "\n"))
	return fmt.Sprintf("<pre><code>%s</code></pre>", code)
}

func isListItem(block string) bool {
	lines := strings.Split(block, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		return strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") ||
			regexp.MustCompile(`^\d+\.\s`).MatchString(trimmed)
	}
	return false
}

func renderList(block string) string {
	lines := strings.Split(block, "\n")
	var items []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Strip list marker
		content := trimmed
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			content = trimmed[2:]
		} else if match := regexp.MustCompile(`^\d+\.\s`).FindString(trimmed); match != "" {
			content = trimmed[len(match):]
		}
		items = append(items, fmt.Sprintf("<li>%s</li>", renderInline(content)))
	}
	return fmt.Sprintf("<ul>\n%s\n</ul>", strings.Join(items, "\n"))
}

func renderParagraph(block string) string {
	lines := strings.Split(block, "\n")
	var inlineParts []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		inlineParts = append(inlineParts, renderInline(trimmed))
	}
	return fmt.Sprintf("<p>%s</p>", strings.Join(inlineParts, " "))
}

var (
	boldRe          = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRe        = regexp.MustCompile(`\*(.+?)\*`)
	inlineCodeRe    = regexp.MustCompile("`([^`]+)`")
	linkRe          = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	imageRe         = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	strikethroughRe = regexp.MustCompile(`~~(.+?)~~`)
)

func renderInline(text string) string {
	// Escape HTML first, then apply markdown formatting
	text = html.EscapeString(text)

	// Images
	text = imageRe.ReplaceAllString(text, `<img src="$2" alt="$1" style="max-width:100%%;">`)

	// Links
	text = linkRe.ReplaceAllString(text, `<a href="$2" style="color:#2563eb;">$1</a>`)

	// Bold
	text = boldRe.ReplaceAllString(text, `<strong>$1</strong>`)

	// Italic (after bold to avoid conflict)
	text = italicRe.ReplaceAllString(text, `<em>$1</em>`)

	// Strikethrough
	text = strikethroughRe.ReplaceAllString(text, `<del>$1</del>`)

	// Inline code
	text = inlineCodeRe.ReplaceAllString(text, `<code style="background:#f1f5f9;padding:2px 6px;border-radius:4px;font-size:0.95em;">$1</code>`)

	return text
}
