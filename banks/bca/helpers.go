package bca

import "strings"

// splitLines splits text into lines, normalizing line endings.
func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}

// extractMetaValue extracts the value part from "Key,=,Value" format.
func extractMetaValue(line string) string {
	parts := strings.SplitN(line, ",=,", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// cleanQuote removes leading single quotes used in BCA CSVs (e.g. '0154363374 → 0154363374).
func cleanQuote(s string) string {
	return strings.TrimPrefix(strings.TrimSpace(s), "'")
}

// stripAllQuotes removes surrounding double quotes from a string.
func stripAllQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// parseCSVQuoted parses a line of quoted CSV fields.
// e.g. "field1","field2","field3" → ["field1", "field2", "field3"]
func parseCSVQuoted(line string) []string {
	var fields []string
	line = strings.TrimSpace(line)

	for line != "" {
		if line[0] == '"' {

			end := strings.Index(line[1:], "\"")
			if end == -1 {
				fields = append(fields, line[1:])
				break
			}
			fields = append(fields, line[1:end+1])
			line = line[end+2:]

			if len(line) > 0 && line[0] == ',' {
				line = line[1:]
			}
		} else {

			end := strings.Index(line, ",")
			if end == -1 {
				fields = append(fields, strings.TrimSpace(line))
				break
			}
			fields = append(fields, strings.TrimSpace(line[:end]))
			line = line[end+1:]
		}
	}

	return fields
}

// formatBCADate converts BCA date format DD/MM/YYYY to a display string.
func formatBCADate(date string) string {

	date = cleanQuote(date)
	return date
}
