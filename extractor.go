package bankset

import (
	"strings"

	"github.com/kankburhan/go-bankset/core"
	"github.com/ledongthuc/pdf"
)

// ExtractPDFText extracts plain text from a PDF file.
func ExtractPDFText(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var builder strings.Builder

	for pageIndex := 1; pageIndex <= reader.NumPage(); pageIndex++ {
		page := reader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}

		builder.WriteString(pageText)
		builder.WriteString("\n")
	}

	result := strings.TrimSpace(builder.String())
	if result == "" {
		return "", core.ErrEmptyPDF
	}

	return result, nil
}
