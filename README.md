# Go-BankSet

Parse Indonesian bank statement PDFs and CSVs into structured Go data.

> Supported: Bank Jago (PDF) and BCA Personal/Bisnis (CSV).

## Install

```bash
go get github.com/kankburhan/go-bankset
```

## Usage

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kankburhan/go-bankset"
)

func main() {
	statement, err := bankstatement.ParseFile(context.Background(), "statement.pdf")
	if err != nil {
		log.Fatal(err)
	}

	output, _ := json.MarshalIndent(statement, "", "  ")
	fmt.Println(string(output))
}
```

## Output example

```json
{
  "bank": "JAGO",
  "account_name": "JOHN DOE",
  "account_no": "100000000001",
  "period": "April 2026",
  "currency": "IDR",
  "pockets": [
    {
      "name": "Main Pocket",
      "transactions": [
        {
          "date": "13 Apr 2026",
          "time": "11:33",
          "source_destination": "JOHN DOE — Jago 100000000002",
          "transaction_detail": "RDN Disbursement",
          "transaction_id": "260413-ABCD-EFGH01",
          "mutation_type": "CR",
          "notes": "WD jago",
          "amount": {
            "currency": "IDR",
            "display": "+Rp221.861",
            "value": 221861
          },
          "balance": {
            "currency": "IDR",
            "display": "Rp245.210",
            "value": 245210
          }
        }
      ]
    }
  ]
}
```

## Supported banks

| Bank | Format | Status |
|---|---|---|
| Bank Jago | Monthly Statement PDF | ✅ Supported |
| Bank Jago | Pockets Transactions History PDF | ✅ Supported |
| BCA Personal | CSV e-statement (KlikBCA) | ✅ Supported |
| BCA Bisnis | CSV e-statement (KlikBCA Bisnis) | ✅ Supported |
| Mandiri | PDF / e-statement | Planned |
| BRI | PDF / e-statement | Planned |
| BNI | PDF / e-statement | Planned |
| SeaBank | PDF / e-statement | Planned |
| Blu | PDF / e-statement | Planned |

## API

### Parse a file (PDF or CSV, auto-detected)

```go
statement, err := bankstatement.ParseFile(ctx, "statement.pdf")
statement, err := bankstatement.ParseFile(ctx, "statement.csv")
```

### Parse extracted text

```go
statement, err := bankstatement.ParseText(ctx, text)
```

### Use custom parsers

```go
import (
	"github.com/kankburhan/go-bankset"
	"github.com/kankburhan/go-bankset/banks/jago"
)

statement, err := bankstatement.ParseTextWithParsers(ctx, text, jago.NewJagoParser())
```

## Model Structure (`core` package)

```go
type Statement struct {
	Bank        BankCode      `json:"bank"`
	AccountName string        `json:"account_name,omitempty"`
	AccountNo   string        `json:"account_no,omitempty"`
	Period      string        `json:"period,omitempty"`
	Currency    string        `json:"currency,omitempty"`
	Pockets     []PocketGroup `json:"pockets"`
}

type PocketGroup struct {
	Name         string        `json:"name"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Date              string `json:"date"`
	Time              string `json:"time"`
	SourceDestination string `json:"source_destination"`
	TransactionDetail string `json:"transaction_detail"`
	TransactionID     string `json:"transaction_id,omitempty"`
	MutationType      string `json:"mutation_type"`
	Notes             string `json:"notes,omitempty"`
	Amount            Money  `json:"amount"`
	Balance           Money  `json:"balance"`
	Raw               string `json:"raw,omitempty"`
}
```

## Project structure

```
bankstatement.go       — Public facade API (ParseFile, ParseText, ParseCSVFile)
extractor.go           — PDF text extraction
core/              — Domain models and core logic
  model.go         — Statement, PocketGroup, Transaction, Money
  errors.go        — Error definitions
  parser.go        — Parser interface and Registry
  utils.go         — Shared utilities (IDR + BCA money parsing)
banks/             — Bank-specific parsers
  jago/
    jago.go        — Bank Jago parser
    jago_test.go   — Bank Jago tests
  bca/
    personal.go    — BCA Personal parser
    bisnis.go      — BCA Bisnis parser
    helpers.go     — Shared BCA CSV helpers
    bca_test.go    — BCA parser tests
examples/          — Usage examples
```

## Privacy

Do not commit real bank statements into this repository. Use anonymized fixtures for tests.

## License

MIT