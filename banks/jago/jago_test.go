package jago

import (
	"context"
	"testing"

	"github.com/kankburhan/go-bankset/core"
)

const jagoTestInput = `
Monthly Statements April 2026 Page 2 of 3
PT Bank Jago Tbk is licensed and supervised by Financial Services Authority (OJK), Bank Indonesia,
and also a member of Indonesia Deposit Insurance Corporation (LPS) deposit insurance program. www.jago.com

JOHN DOE / 100000000001

Main Pocket
Pocket ID 100000000001
Pocket is created on 17 Dec 2023
Previous Balance
Total Incoming
Total Outgoing
Closing Balance
23.349,76
+221.876,48
-245.003,10
223,15
Date & Time Source/Destination Transaction Details Notes Amount Balance
13 Apr 2026
11:33
JOHN
DOE
Jago 100000000002
RDN Disbursement
ID# 260413-ABCD-EFGH01
WD jago
+221.861
245.210
13 Apr 2026
11:55
JOHN
DOE
Mandiri 1390000000001
Outgoing Transfer
ID# 260413-WXYZ-123456
-245.000
210
28 Apr 2026
06:41
Interest
Main Pocket
Interest
ID# 260428-INTX-000001
+15
226
28 Apr 2026
06:41
Tax on Interest
Main Pocket
Tax on Interest
ID# 260428-TAXR-000001
-3
223

Stockbit Sekuritas RDN
Pocket ID 100000000002
Pocket is created on22 Dec 2025
Previous Balance
Total Incoming
Total Outgoing
Closing Balance
221.861,60
+359.949,58
-221.951,12
359.860,06
Date & Time Source/Destination Transaction Details Notes Amount Balance
13 Apr 2026
11:33
RDN Withdrawal
RDN Withdrawal
ID# 3500000001
WD jago
-221.861
0
15 Apr 2026
10:34
Incoming Fund
Incoming Fund
ID# 3500000002
{botd} P-000001
Payment to: john
doe
+359.498
359.499
28 Apr 2026
01:04
Interest
Stockbit Sekuritas RDN
Interest
ID# 3500000003
+450
359.950
28 Apr 2026
01:04
Tax on Interest
Stockbit Sekuritas RDN
Tax on Interest
ID# 3500000004
-90
359.860

GoPay Tabungan
Pocket ID 100000000003
Previous Balance
Total Incoming
Total Outgoing
Closing Balance
0,00
+0,00
-0,00
0,00
`

const jagoHistoryInput = `
PT Bank Jago Tbk is licensed and supervised by Financial Services Authority (OJK), Bank Indonesia, and
also a member of Indonesia Deposit Insurance Corporation (LPS) deposit insurance program.
www.jago.com
Pockets Transactions
History
Page 1 of 47
JANE SMITH
savings / 100000000099
Showing IDR transaction from
Latest Balance per 14 May 2026
28 Mar 2022 - 14 May 2026
IDR 16
Date & Time
Source/Destination
Transaction Details
Notes
Amount
Balance
March 2022
28 Mar 2022
13:29
Main Pocket
Movement between Pockets
Pocket Money In
ID# 42000001
+200.000
200.000
28 Mar 2022
13:32
acme
Movement between Pockets
Movement between Pockets
ID# 42000002
+930.602
1.130.602
`

func TestJagoParser_MonthlyStatement(t *testing.T) {
	parser := NewJagoParser()
	statement, err := parser.Parse(context.Background(), jagoTestInput)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if statement.Bank != core.BankJago {
		t.Fatalf("bank = %s, want %s", statement.Bank, core.BankJago)
	}

	if statement.AccountName != "JOHN DOE" {
		t.Errorf("account_name = %q, want %q", statement.AccountName, "JOHN DOE")
	}
	if statement.AccountNo != "100000000001" {
		t.Errorf("account_no = %q, want %q", statement.AccountNo, "100000000001")
	}

	if got, want := len(statement.Pockets), 2; got != want {
		t.Fatalf("pockets len = %d, want %d", got, want)
	}

	mainPocket := statement.Pockets[0]
	assertEqual(t, "pocket name", mainPocket.Name, "Main Pocket")
	if got, want := len(mainPocket.Transactions), 4; got != want {
		t.Fatalf("Main Pocket transactions len = %d, want %d", got, want)
	}

	stockbitPocket := statement.Pockets[1]
	assertEqual(t, "pocket name", stockbitPocket.Name, "Stockbit Sekuritas RDN")
	if got, want := len(stockbitPocket.Transactions), 4; got != want {
		t.Fatalf("Stockbit transactions len = %d, want %d", got, want)
	}
}

func TestJagoParser_HistoryPDF(t *testing.T) {
	parser := NewJagoParser()
	statement, err := parser.Parse(context.Background(), jagoHistoryInput)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if statement.Bank != core.BankJago {
		t.Fatalf("bank = %s, want %s", statement.Bank, core.BankJago)
	}

	if statement.AccountName != "JANE SMITH" {
		t.Errorf("account_name = %q, want %q", statement.AccountName, "JANE SMITH")
	}

	if got, want := len(statement.Pockets), 1; got != want {
		t.Fatalf("pockets len = %d, want %d", got, want)
	}

	pocket := statement.Pockets[0]
	assertEqual(t, "pocket name", pocket.Name, "savings")
	if got, want := len(pocket.Transactions), 2; got != want {
		t.Fatalf("Transactions len = %d, want %d", got, want)
	}

	firstTx := pocket.Transactions[0]
	assertEqual(t, "date", firstTx.Date, "2022-03-28")
	assertEqual(t, "time", firstTx.Time, "13:29")
	assertEqual(t, "source_destination", firstTx.SourceDestination, "Main Pocket")
	if firstTx.Amount.Value != 200000 {
		t.Errorf("firstTx amount = %d", firstTx.Amount.Value)
	}

	secondTx := pocket.Transactions[1]
	assertEqual(t, "date", secondTx.Date, "2022-03-28")
	assertEqual(t, "time", secondTx.Time, "13:32")
	assertEqual(t, "source_destination", secondTx.SourceDestination, "acme")
	if secondTx.Amount.Value != 930602 {
		t.Errorf("secondTx amount = %d", secondTx.Amount.Value)
	}
}

func TestJagoParser_CanParse(t *testing.T) {
	parser := NewJagoParser()

	if !parser.CanParse("something PT Bank Jago Tbk something") {
		t.Error("should detect PT Bank Jago Tbk")
	}
	if !parser.CanParse("something www.jago.com something") {
		t.Error("should detect www.jago.com")
	}
	if parser.CanParse("random text without bank identifiers") {
		t.Error("should not detect unrelated text")
	}
}

func TestJagoParser_EmptyInput(t *testing.T) {
	parser := NewJagoParser()
	_, err := parser.Parse(context.Background(), "")
	if err != core.ErrEmptyPDF {
		t.Fatalf("expected ErrEmptyPDF, got %v", err)
	}
}

func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %q, want %q", field, got, want)
	}
}
