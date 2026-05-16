package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/kankburhan/go-bankset"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./examples/parse_pdf <statement.pdf>")
	}

	statement, err := bankset.ParseFile(context.Background(), os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	output, err := json.MarshalIndent(statement, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}
