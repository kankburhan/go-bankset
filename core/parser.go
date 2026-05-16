package core

import "context"

// Parser defines the interface that bank-specific parsers must implement.
type Parser interface {
	// Bank returns the bank code this parser handles.
	Bank() BankCode

	// CanParse checks if the given text matches this bank's statement format.
	CanParse(text string) bool

	// Parse parses the extracted text into a structured Statement.
	Parse(ctx context.Context, text string) (*Statement, error)
}

// Registry holds registered parsers and dispatches parsing to the correct one.
type Registry struct {
	parsers []Parser
}

// NewRegistry creates a new Registry with the given parsers.
func NewRegistry(parsers ...Parser) *Registry {
	return &Registry{parsers: parsers}
}

// Parse iterates through registered parsers and uses the first one
// that recognizes the text format.
func (r *Registry) Parse(ctx context.Context, text string) (*Statement, error) {
	for _, parser := range r.parsers {
		if parser.CanParse(text) {
			return parser.Parse(ctx, text)
		}
	}

	return nil, ErrUnsupportedBank
}
