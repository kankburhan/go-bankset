package bca

import (
	"context"
	"testing"

	"github.com/kankburhan/go-bankset/core"
)

const bcaPersonalInput = `Account No.,=,'0111222333
Name,=,JOHN DOE
Currency,=,IDR

Date,Description,Branch,Amount,,Balance
'02/05/2026,TRSF E-BANKING CR 0205/FTSCY/WS95031         108400.00Transfer pembayaran invoice           ACME CORPORATION,'0000,108400.00,CR,1042939.96
Starting Balance,=,934539.96
Credit,=,108400.00
Debet,=,
Ending Balance,=,1042939.96
`

const bcaBisnisInput = `"Informasi Rekening - Mutasi Rekening"," "," "," "," ",

"No. rekening : 0350000001"
"Nama : PT ACME TEKNOLOGI"
"Periode : 13/05/2026 - 14/05/2026"
"Kode Mata Uang : Rp"
"Tanggal Transaksi","Keterangan","Cabang","Jumlah","Saldo"
"13/05/2026","KR OTOMATIS NTRF@0000000001X0Z 035@BCA26051800001  @IFP acme.co.id @AFR  PaymentGateway      ","0000","3,528,964.00 CR","25,441,156.00"
"Saldo Awal : 21,912,192.00"
"Mutasi Debet : 0.00","0"
"Mutasi Kredit : 3,528,964.00","1"
"Saldo Akhir : 25,441,156.00"
`

func TestBCAPersonalParser(t *testing.T) {
	parser := NewPersonalParser()

	if !parser.CanParse(bcaPersonalInput) {
		t.Fatal("CanParse should return true for BCA Personal CSV")
	}

	stmt, err := parser.Parse(context.Background(), bcaPersonalInput)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if stmt.Bank != core.BankBCA {
		t.Fatalf("bank = %s, want %s", stmt.Bank, core.BankBCA)
	}
	if stmt.AccountNo != "0111222333" {
		t.Errorf("account_no = %q, want %q", stmt.AccountNo, "0111222333")
	}
	if stmt.AccountName != "JOHN DOE" {
		t.Errorf("account_name = %q, want %q", stmt.AccountName, "JOHN DOE")
	}

	if len(stmt.Pockets) != 1 {
		t.Fatalf("pockets len = %d, want 1", len(stmt.Pockets))
	}

	pocket := stmt.Pockets[0]
	if pocket.Name != "Personal" {
		t.Errorf("pocket name = %q, want %q", pocket.Name, "Personal")
	}
	if len(pocket.Transactions) != 1 {
		t.Fatalf("transactions len = %d, want 1", len(pocket.Transactions))
	}

	tx := pocket.Transactions[0]
	if tx.Date != "02/05/2026" {
		t.Errorf("date = %q, want %q", tx.Date, "02/05/2026")
	}
	if tx.MutationType != "CR" {
		t.Errorf("mutation_type = %q, want %q", tx.MutationType, "CR")
	}
	if tx.Amount.Value != 108400 {
		t.Errorf("amount value = %d, want 108400", tx.Amount.Value)
	}
	if tx.Balance.Value != 1042939 {
		t.Errorf("balance value = %d, want 1042939", tx.Balance.Value)
	}
}

func TestBCABisnisParser(t *testing.T) {
	parser := NewBisnisParser()

	if !parser.CanParse(bcaBisnisInput) {
		t.Fatal("CanParse should return true for BCA Bisnis CSV")
	}

	stmt, err := parser.Parse(context.Background(), bcaBisnisInput)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if stmt.Bank != core.BankBCABisnis {
		t.Fatalf("bank = %s, want %s", stmt.Bank, core.BankBCABisnis)
	}
	if stmt.AccountNo != "0350000001" {
		t.Errorf("account_no = %q, want %q", stmt.AccountNo, "0350000001")
	}
	if stmt.AccountName != "PT ACME TEKNOLOGI" {
		t.Errorf("account_name = %q, want %q", stmt.AccountName, "PT ACME TEKNOLOGI")
	}
	if stmt.Period != "13/05/2026 - 14/05/2026" {
		t.Errorf("period = %q, want %q", stmt.Period, "13/05/2026 - 14/05/2026")
	}

	if len(stmt.Pockets) != 1 {
		t.Fatalf("pockets len = %d, want 1", len(stmt.Pockets))
	}

	pocket := stmt.Pockets[0]
	if pocket.Name != "Business" {
		t.Errorf("pocket name = %q, want %q", pocket.Name, "Business")
	}
	if len(pocket.Transactions) != 1 {
		t.Fatalf("transactions len = %d, want 1", len(pocket.Transactions))
	}

	tx := pocket.Transactions[0]
	if tx.Date != "13/05/2026" {
		t.Errorf("date = %q, want %q", tx.Date, "13/05/2026")
	}
	if tx.MutationType != "CR" {
		t.Errorf("mutation_type = %q, want %q", tx.MutationType, "CR")
	}
	if tx.Amount.Value != 3528964 {
		t.Errorf("amount value = %d, want 3528964", tx.Amount.Value)
	}
	if tx.Balance.Value != 25441156 {
		t.Errorf("balance value = %d, want 25441156", tx.Balance.Value)
	}
}

func TestBCACanParseNegative(t *testing.T) {
	parser := NewPersonalParser()
	if parser.CanParse("random text without bank identifiers") {
		t.Error("should not detect unrelated text")
	}

	bisnisParser := NewBisnisParser()
	if bisnisParser.CanParse("random text without bank identifiers") {
		t.Error("should not detect unrelated text")
	}
}
