package core

// BankCode identifies the bank that issued a statement.
type BankCode string

const (
	BankUnknown	BankCode	= "UNKNOWN"
	BankJago	BankCode	= "JAGO"
	BankBCA		BankCode	= "BCA"
	BankBCABisnis	BankCode	= "BCA_BISNIS"
)

// Statement represents a parsed bank statement with metadata and transactions grouped by pocket.
type Statement struct {
	Bank		BankCode	`json:"bank"`
	AccountName	string		`json:"account_name,omitempty"`
	AccountNo	string		`json:"account_no,omitempty"`
	Period		string		`json:"period,omitempty"`
	Currency	string		`json:"currency,omitempty"`
	Pockets		[]PocketGroup	`json:"pockets"`
}

// PocketGroup groups transactions by their corresponding pocket.
type PocketGroup struct {
	Name		string		`json:"name"`
	Transactions	[]Transaction	`json:"transactions"`
}

// Transaction represents a single transaction entry in a bank statement.
// Date is always in "YYYY-MM-DD" format. Time is "HH:mm" or empty if unavailable.
type Transaction struct {
	Date			string	`json:"date"`
	Time			string	`json:"time"`
	SourceDestination	string	`json:"source_destination"`
	TransactionDetail	string	`json:"transaction_detail"`
	TransactionID		string	`json:"transaction_id,omitempty"`
	MutationType		string	`json:"mutation_type"`
	Notes			string	`json:"notes,omitempty"`
	Amount			Money	`json:"amount"`
	Balance			Money	`json:"balance"`
	Raw			string	`json:"raw,omitempty"`
}

// Money represents a monetary value with display formatting.
type Money struct {
	Currency	string	`json:"currency"`
	Display		string	`json:"display"`
	Value		int64	`json:"value"`
}
