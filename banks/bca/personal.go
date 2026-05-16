package bca

import (
	"context"
	"strings"

	"github.com/kankburhan/go-bankset/core"
)

// PersonalParser parses BCA Personal (KlikBCA Individual) e-statement CSVs.
//
// CSV format:
//
//	Account No.,=,'0154363374
//	Name,=,ALLEIN MARTIN YERE
//	Currency,=,IDR
//
//	Date,Description,Branch,Amount,,Balance
//	'02/05/2026,TRSF E-BANKING CR ...,0000,108400.00,CR,1042939.96
//	Starting Balance,=,934539.96
//	...
type PersonalParser struct{}

// NewPersonalParser creates a new BCA Personal parser.
func NewPersonalParser() *PersonalParser {
	return &PersonalParser{}
}

func (p *PersonalParser) Bank() core.BankCode {
	return core.BankBCA
}

func (p *PersonalParser) CanParse(text string) bool {
	return strings.Contains(text, "Account No.,=,") &&
		strings.Contains(text, "Starting Balance")
}

func (p *PersonalParser) Parse(ctx context.Context, text string) (*core.Statement, error) {
	lines := splitLines(text)
	if len(lines) == 0 {
		return nil, core.ErrEmptyPDF
	}

	stmt := &core.Statement{
		Bank:		core.BankBCA,
		Currency:	"IDR",
	}

	var transactions []core.Transaction
	inData := false

	for _, line := range lines {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Account No.,=,") {
			stmt.AccountNo = cleanQuote(extractMetaValue(line))
			continue
		}
		if strings.HasPrefix(line, "Name,=,") {
			stmt.AccountName = extractMetaValue(line)
			continue
		}

		if strings.HasPrefix(line, "Date,Description,") {
			inData = true
			continue
		}

		if strings.HasPrefix(line, "Starting Balance") ||
			strings.HasPrefix(line, "Credit,") ||
			strings.HasPrefix(line, "Debet,") ||
			strings.HasPrefix(line, "Ending Balance") {
			inData = false
			continue
		}

		if !inData {
			continue
		}

		tx := p.parsePersonalRow(line)
		if tx.Date != "" {
			transactions = append(transactions, tx)
		}
	}

	if len(transactions) > 0 {
		stmt.Pockets = []core.PocketGroup{
			{
				Name:		"Personal",
				Transactions:	transactions,
			},
		}
	}

	return stmt, nil
}

// parsePersonalRow parses a single BCA Personal CSV data row.
// Format: 'DD/MM/YYYY,Description,Branch,Amount,CR/DB,Balance
func (p *PersonalParser) parsePersonalRow(line string) core.Transaction {

	parts := strings.Split(line, ",")
	if len(parts) < 6 {
		return core.Transaction{}
	}

	balance := strings.TrimSpace(parts[len(parts)-1])
	crdb := strings.TrimSpace(parts[len(parts)-2])
	amount := strings.TrimSpace(parts[len(parts)-3])
	_ = strings.TrimSpace(parts[len(parts)-4])

	date := cleanQuote(strings.TrimSpace(parts[0]))

	descParts := parts[1 : len(parts)-4]
	description := core.CompactLine(strings.Join(descParts, ","))

	isCredit := strings.ToUpper(crdb) == "CR"

	mutationType := "DB"
	if isCredit {
		mutationType = "CR"
	}

	return core.Transaction{
		Date:			formatBCADate(date),
		Time:			"-",
		SourceDestination:	"-",
		TransactionDetail:	description,
		MutationType:		mutationType,
		Notes:			"-",
		Amount:			core.ParseBCAMoney(amount, isCredit),
		Balance:		core.ParseBCABalance(balance),
		Raw:			line,
	}
}
