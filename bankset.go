package bankset

import (
	"context"
	"os"
	"strings"

	"github.com/kankburhan/go-bankset/banks/bca"
	"github.com/kankburhan/go-bankset/banks/jago"
	"github.com/kankburhan/go-bankset/core"
)

// ParseText parses extracted text from a bank statement using the default registry.
func ParseText(ctx context.Context, text string) (*core.Statement, error) {
	return DefaultRegistry().Parse(ctx, text)
}

// ParseFile reads a file (PDF or CSV), extracts the text, and parses the bank statement.
func ParseFile(ctx context.Context, path string) (*core.Statement, error) {
	lower := strings.ToLower(path)

	if strings.HasSuffix(lower, ".csv") {
		return ParseCSVFile(ctx, path)
	}

	text, err := ExtractPDFText(path)
	if err != nil {
		return nil, err
	}

	return ParseText(ctx, text)
}

// ParseCSVFile reads a CSV file and parses the bank statement.
func ParseCSVFile(ctx context.Context, path string) (*core.Statement, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	text := string(data)
	if strings.TrimSpace(text) == "" {
		return nil, core.ErrEmptyPDF
	}

	return ParseText(ctx, text)
}

// ParseTextWithParsers parses extracted text using the provided parsers.
func ParseTextWithParsers(ctx context.Context, text string, parsers ...core.Parser) (*core.Statement, error) {
	return core.NewRegistry(parsers...).Parse(ctx, text)
}

// DefaultRegistry returns a registry with all built-in bank parsers.
func DefaultRegistry() *core.Registry {
	return core.NewRegistry(
		jago.NewJagoParser(),
		bca.NewPersonalParser(),
		bca.NewBisnisParser(),
	)
}
