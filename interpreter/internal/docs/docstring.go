package docs

import (
	"regexp"
	"strings"
	"unicode"
)

// Dedent normalizes a Starlark docstring for rendering as Markdown.
//
// The first line of a """...""" literal begins immediately after the quotes and
// so carries no indentation, while continuation lines carry the indentation of
// the enclosing def. Only the continuation lines contribute to the common
// prefix that gets stripped.
func Dedent(doc string) string {
	if doc == "" {
		return ""
	}

	lines := strings.Split(strings.ReplaceAll(doc, "\t", "    "), "\n")

	indent := -1
	for _, line := range lines[1:] {
		trimmed := strings.TrimLeft(line, " ")
		if trimmed == "" {
			continue // blank lines never constrain the common prefix
		}
		if n := len(line) - len(trimmed); indent < 0 || n < indent {
			indent = n
		}
	}

	out := make([]string, 0, len(lines))
	out = append(out, strings.TrimSpace(lines[0]))
	for _, line := range lines[1:] {
		// Strip exactly the common prefix, never more: relative indentation
		// carries meaning in nested lists and fenced examples. A line too short
		// to carry the prefix is whitespace-only, since every other line has at
		// least that much.
		if indent > 0 {
			if len(line) >= indent {
				line = line[indent:]
			} else {
				line = ""
			}
		}
		out = append(out, strings.TrimRight(line, " "))
	}

	for len(out) > 0 && out[0] == "" {
		out = out[1:]
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}

	return strings.Join(out, "\n")
}

var sectionHeading = regexp.MustCompile(`^(Args|Arguments|Returns|Raises|Example|Examples|Note|Notes):$`)

// FormatDoc converts Google-style sections in an already-dedented docstring to
// Markdown: section headers are bolded, and Args entries become a bullet list.
// Anything unrecognized passes through verbatim, so a docstring can use plain
// Markdown throughout.
func FormatDoc(doc string) string {
	var out []string
	inArgs := false
	inFence := false
	afterHeading := false

	for line := range strings.SplitSeq(doc, "\n") {
		trimmed := strings.TrimLeft(line, " ")

		// Fenced blocks pass through untouched; an indented example inside an
		// Args section must not be mistaken for an argument entry.
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			out = append(out, line)
			continue
		}
		if inFence {
			out = append(out, line)
			continue
		}

		if m := sectionHeading.FindStringSubmatch(line); m != nil {
			if len(out) > 0 && out[len(out)-1] != "" {
				out = append(out, "")
			}
			out = append(out, "**"+m[1]+":**", "")
			inArgs = m[1] == "Args" || m[1] == "Arguments"
			afterHeading = true
			continue
		}

		// The blank line a header already emitted stands in for one written in
		// the source, so the two do not stack up.
		if afterHeading && strings.TrimSpace(line) == "" {
			afterHeading = false
			continue
		}
		afterHeading = false

		if !inArgs {
			out = append(out, line)
			continue
		}

		if trimmed == "" {
			out = append(out, "")
			continue
		}
		if line == trimmed {
			// An unindented line ends the Args block.
			inArgs = false
			out = append(out, line)
			continue
		}
		if name, desc, ok := strings.Cut(trimmed, ": "); ok && isParamName(name) {
			out = append(out, "- `"+name+"`: "+desc)
			continue
		}
		if len(out) > 0 && strings.HasPrefix(out[len(out)-1], "- ") {
			out[len(out)-1] += " " + trimmed // continuation of the previous entry
			continue
		}
		out = append(out, trimmed)
	}

	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

func isParamName(s string) bool {
	s = strings.TrimLeft(s, "*")
	if s == "" {
		return false
	}
	for i, r := range s {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r)) {
			continue
		}
		return false
	}
	return true
}
