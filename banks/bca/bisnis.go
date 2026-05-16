package bca

import (
	"context"
	"strings"

	"github.com/kankburhan/go-bankset/core"
)

// BisnisParser parses BCA Bisnis (Corporate) e-statement CSVs.
//
// CSV format (all values quoted):
//
//	"No. rekening : 0353520838"
//	"Nama : GASS MARKETING TEKNOLOGI"
//	"Periode : 13/05/2026 - 14/05/2026"
//	"Kode Mata Uang : Rp"
//	"Tanggal Transaksi","Keterangan","Cabang","Jumlah","Saldo"
//	"13/05/2026","description","0000","3,528,964.00 CR","25,441,156.00"
//	"Saldo Awal : 21,912,192.00"
//	...
type BisnisParser struct{}

// NewBisnisParser creates a new BCA Bisnis parser.
func NewBisnisParser() *BisnisParser {
	return &BisnisParser{}
}

func (p *BisnisParser) Bank() core.BankCode {
	return core.BankBCABisnis
}

func (p *BisnisParser) CanParse(text string) bool {
	return strings.Contains(text, "No. rekening :") &&
		(strings.Contains(text, "Saldo Awal") || strings.Contains(text, "Tanggal Transaksi"))
}

func (p *BisnisParser) Parse(ctx context.Context, text string) (*core.Statement, error) {
	lines := splitLines(text)
	if len(lines) == 0 {
		return nil, core.ErrEmptyPDF
	}

	stmt := &core.Statement{
		Bank:		core.BankBCABisnis,
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

		unquoted := stripAllQuotes(line)

		if strings.HasPrefix(unquoted, "No. rekening :") {
			stmt.AccountNo = strings.TrimSpace(strings.TrimPrefix(unquoted, "No. rekening :"))
			continue
		}
		if strings.HasPrefix(unquoted, "Nama :") {
			stmt.AccountName = strings.TrimSpace(strings.TrimPrefix(unquoted, "Nama :"))
			continue
		}
		if strings.HasPrefix(unquoted, "Periode :") {
			stmt.Period = strings.TrimSpace(strings.TrimPrefix(unquoted, "Periode :"))
			continue
		}
		if strings.HasPrefix(unquoted, "Informasi Rekening") ||
			strings.HasPrefix(unquoted, "Kode Mata Uang") {
			continue
		}

		if strings.Contains(line, "Tanggal Transaksi") {
			inData = true
			continue
		}

		if strings.HasPrefix(unquoted, "Saldo Awal") ||
			strings.HasPrefix(unquoted, "Mutasi Debet") ||
			strings.HasPrefix(unquoted, "Mutasi Kredit") ||
			strings.HasPrefix(unquoted, "Saldo Akhir") {
			inData = false
			continue
		}

		if !inData {
			continue
		}

		tx := p.parseBisnisRow(line)
		if tx.Date != "" {
			transactions = append(transactions, tx)
		}
	}

	if len(transactions) > 0 {
		stmt.Pockets = []core.PocketGroup{
			{
				Name:		"Business",
				Transactions:	transactions,
			},
		}
	}

	return stmt, nil
}

// parseBisnisRow parses a single BCA Bisnis CSV data row.
// Format: "date","description","branch","amount CR/DB","balance"
func (p *BisnisParser) parseBisnisRow(line string) core.Transaction {
	fields := parseCSVQuoted(line)
	if len(fields) < 5 {
		return core.Transaction{}
	}

	date := fields[0]
	description := core.CompactLine(fields[1])
	amount := fields[3]
	balance := fields[4]

	mutationType := "DB"
	if strings.HasSuffix(strings.TrimSpace(amount), "CR") {
		mutationType = "CR"
	}

	return core.Transaction{
		Date:			formatBCADate(date),
		Time:			"-",
		SourceDestination:	"-",
		TransactionDetail:	description,
		MutationType:		mutationType,
		Notes:			"-",
		Amount:			core.ParseBCAMoney(amount, false),
		Balance:		core.ParseBCABalance(balance),
		Raw:			line,
	}
}
