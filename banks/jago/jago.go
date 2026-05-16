package jago

import (
	"context"
	"regexp"
	"strings"

	"github.com/kankburhan/go-bankset/core"
)

// JagoParser parses Bank Jago monthly statement PDFs and History PDFs.
type JagoParser struct {
	dateRegex	*regexp.Regexp
	timeRegex	*regexp.Regexp
	idRegex		*regexp.Regexp
	moneyRegex	*regexp.Regexp
	accountRegex	*regexp.Regexp
	accountIDRegex	*regexp.Regexp
	monthYearRegex	*regexp.Regexp
	knownPocketName	map[string]bool
}

// NewJagoParser creates a new parser for Bank Jago statements.
func NewJagoParser() *JagoParser {
	return &JagoParser{
		dateRegex:	regexp.MustCompile(`^\d{1,2} [A-Za-z]{3} \d{4}$`),
		timeRegex:	regexp.MustCompile(`^\d{2}:\d{2}$`),
		idRegex:	regexp.MustCompile(`ID#\s*([A-Za-z0-9\-]+)`),
		moneyRegex:	regexp.MustCompile(`^[+-]?\d{1,3}(\.\d{3})*$`),
		accountRegex:	regexp.MustCompile(`(?i)(jago|mandiri|bca|bni|bri|cimb|permata|danamon|ocbc|seabank|blu|line bank|bank|wallet|\d{6,})`),
		accountIDRegex:	regexp.MustCompile(`^(.+)\s*/\s*(\d+)$`),
		monthYearRegex:	regexp.MustCompile(`^[A-Za-z]+ \d{4}$`),
		knownPocketName: map[string]bool{
			"Main Pocket":			true,
			"Stockbit Sekuritas RDN":	true,
			"GoPay Tabungan":		true,
			"GoPay Simpanan":		true,
			"Euros":			true,
		},
	}
}

func (p *JagoParser) Bank() core.BankCode {
	return core.BankJago
}

func (p *JagoParser) CanParse(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "pt bank jago tbk") ||
		strings.Contains(lower, "www.jago.com") ||
		strings.Contains(lower, "main pocket")
}

// Parse parses Bank Jago statement text into a structured Statement.
func (p *JagoParser) Parse(ctx context.Context, text string) (*core.Statement, error) {
	lines := p.normalizeLines(text)
	if len(lines) == 0 {
		return nil, core.ErrEmptyPDF
	}

	startIndex := p.findTransactionStartIndex(lines)
	if startIndex == -1 {
		return nil, core.ErrInvalidFormat
	}

	statement := &core.Statement{
		Bank:		core.BankJago,
		Currency:	"IDR",
	}

	isHistory, globalPocket := p.extractMetadata(lines[:startIndex], statement)

	lines = lines[startIndex:]

	var currentPocket string
	if isHistory {
		currentPocket = globalPocket
	}

	groups := make(map[string]*core.PocketGroup)
	var orderedPockets []string

	for i := 0; i < len(lines); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := lines[i]

		if !isHistory && p.isPocketHeader(line) {
			currentPocket = line
			continue
		}

		if currentPocket == "" || !p.dateRegex.MatchString(line) {
			continue
		}

		tx, nextIndex := p.parseTransactionBlock(lines, i, currentPocket)
		if tx.Date != "" && tx.Time != "" && tx.TransactionDetail != "" {
			if _, exists := groups[currentPocket]; !exists {
				groups[currentPocket] = &core.PocketGroup{
					Name:		currentPocket,
					Transactions:	[]core.Transaction{},
				}
				orderedPockets = append(orderedPockets, currentPocket)
			}
			groups[currentPocket].Transactions = append(groups[currentPocket].Transactions, tx)
		}
		i = nextIndex
	}

	for _, name := range orderedPockets {
		statement.Pockets = append(statement.Pockets, *groups[name])
	}

	return statement, nil
}

// extractMetadata parses the header section.
// Returns true if this is a History PDF, along with the global pocket name.
func (p *JagoParser) extractMetadata(headerLines []string, stmt *core.Statement) (bool, string) {
	isHistory := false
	globalPocket := ""

	for i, line := range headerLines {
		if strings.Contains(line, "Pockets Transactions") || strings.Contains(line, "History") {
			isHistory = true
		}

		if match := p.accountIDRegex.FindStringSubmatch(line); len(match) == 3 {

			if isHistory {
				globalPocket = strings.TrimSpace(match[1])

				if i > 0 {
					stmt.AccountName = headerLines[i-1]
				}
			} else {
				stmt.AccountName = strings.TrimSpace(match[1])
				stmt.AccountNo = strings.TrimSpace(match[2])
			}
			continue
		}

		if p.looksLikePeriod(line) {
			stmt.Period = line
		}
	}

	if isHistory && globalPocket == "" {
		globalPocket = "Unknown Pocket"
	}

	return isHistory, globalPocket
}

func (p *JagoParser) looksLikePeriod(line string) bool {

	if strings.Contains(line, "-") && regexp.MustCompile(`\d{4}.*\-.*\d{4}`).MatchString(line) {
		return true
	}

	months := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
	for _, m := range months {
		if strings.Contains(line, m) && regexp.MustCompile(`\d{4}`).MatchString(line) {
			return true
		}
	}
	return false
}

func (p *JagoParser) parseTransactionBlock(lines []string, dateIndex int, pocket string) (core.Transaction, int) {
	tx := core.Transaction{
		Date:	core.ParseJagoDate(lines[dateIndex]),
		Notes:	"-",
	}

	i := dateIndex + 1
	if i < len(lines) && p.timeRegex.MatchString(lines[i]) {
		tx.Time = lines[i]
		i++
	}

	var block []string
	seenMoney := false
	for i < len(lines) {
		line := lines[i]

		if p.dateRegex.MatchString(line) || p.isHardStop(line) {
			i--
			break
		}

		if p.monthYearRegex.MatchString(line) {
			i++
			continue
		}

		if seenMoney && p.isPocketHeader(line) {

			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "Pocket ID") {
				i--
				break
			}
		}

		if p.moneyRegex.MatchString(line) {
			seenMoney = true
		}

		block = append(block, line)
		i++
	}

	p.fillTransaction(&tx, pocket, block)
	return tx, i
}

func (p *JagoParser) fillTransaction(tx *core.Transaction, pocket string, block []string) {
	block = p.removeNoise(block)

	var amountCandidates []string
	var textParts []string
	var rawParts []string

	for _, part := range block {
		part = core.CompactLine(part)
		if part == "" {
			continue
		}

		rawParts = append(rawParts, part)

		if p.moneyRegex.MatchString(part) {
			amountCandidates = append(amountCandidates, part)
			continue
		}

		if match := p.idRegex.FindStringSubmatch(part); len(match) == 2 {
			tx.TransactionID = match[1]

			cleaned := core.CompactLine(p.idRegex.ReplaceAllString(part, ""))
			if cleaned != "" {
				textParts = append(textParts, cleaned)
			}
			continue
		}

		textParts = append(textParts, part)
	}

	tx.Raw = strings.Join(rawParts, " ")

	if len(amountCandidates) >= 2 {
		tx.Amount = core.ParseIDR(amountCandidates[len(amountCandidates)-2])
		tx.Balance = core.ParseIDR(amountCandidates[len(amountCandidates)-1])
	} else if len(amountCandidates) == 1 {
		tx.Amount = core.ParseIDR(amountCandidates[0])
	}

	if tx.Amount.Value >= 0 {
		tx.MutationType = "CR"
	} else {
		tx.MutationType = "DB"
	}

	p.parseTextParts(tx, pocket, textParts)
}

func (p *JagoParser) parseTextParts(tx *core.Transaction, pocket string, parts []string) {
	if len(parts) == 0 {
		return
	}

	switch pocket {
	case "Main Pocket":
		p.parseMainPocketTextParts(tx, parts)
	case "Stockbit Sekuritas RDN":
		p.parseStockbitTextParts(tx, parts)
	default:
		p.parseGenericTextParts(tx, parts)
	}
}

func (p *JagoParser) parseMainPocketTextParts(tx *core.Transaction, parts []string) {
	accountIndex := -1
	for i, part := range parts {
		if p.looksLikeAccountInfo(part) {
			accountIndex = i
			break
		}
	}

	if accountIndex > 0 {
		name := strings.Join(parts[:accountIndex], " ")
		account := parts[accountIndex]

		tx.SourceDestination = strings.TrimSpace(name + " — " + account)

		if accountIndex+1 < len(parts) {
			tx.TransactionDetail = parts[accountIndex+1]
		}

		if accountIndex+2 < len(parts) {
			tx.Notes = strings.Join(parts[accountIndex+2:], " ")
		}
		return
	}

	if len(parts) >= 3 && isJagoInterestLike(parts[0]) {
		tx.SourceDestination = parts[0] + " — " + parts[1]
		tx.TransactionDetail = parts[2]
		if len(parts) > 3 {
			tx.Notes = strings.Join(parts[3:], " ")
		}
		return
	}

	p.parseGenericTextParts(tx, parts)
}

func (p *JagoParser) parseStockbitTextParts(tx *core.Transaction, parts []string) {
	if len(parts) >= 3 && isJagoInterestLike(parts[0]) {
		tx.SourceDestination = parts[0] + " — " + parts[1]
		tx.TransactionDetail = parts[2]
		if len(parts) > 3 {
			tx.Notes = strings.Join(parts[3:], " ")
		}
		return
	}

	tx.SourceDestination = parts[0]
	if len(parts) >= 2 {
		tx.TransactionDetail = parts[1]
	}
	if len(parts) >= 3 {
		tx.Notes = strings.Join(parts[2:], " ")
	}
}

func (p *JagoParser) parseGenericTextParts(tx *core.Transaction, parts []string) {
	if len(parts) >= 1 {
		tx.SourceDestination = parts[0]
	}
	if len(parts) >= 2 {
		tx.TransactionDetail = parts[1]
	}
	if len(parts) >= 3 {
		tx.Notes = strings.Join(parts[2:], " ")
	}
}

func (p *JagoParser) normalizeLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	rawLines := strings.Split(text, "\n")
	lines := make([]string, 0, len(rawLines))

	for _, line := range rawLines {
		line = core.CompactLine(line)
		if line == "" || p.shouldSkipLine(line) {
			continue
		}

		lines = append(lines, line)
	}

	return lines
}

func (p *JagoParser) shouldSkipLine(line string) bool {
	skipContains := []string{
		"Monthly Statements",
		"PT Bank Jago Tbk",
		"licensed and supervised",
		"Financial Services Authority",
		"Bank Indonesia",
		"Indonesia Deposit Insurance",
		"www.jago.com",
		"Date & Time",
		"Source/Destination",
		"Transaction Details",
		"Notes",
		"Amount",
		"Balance",
		"Currency In IDR",
		"Currency In EUR",
		"Rates against the Indonesian IDR",
		"BALANCE SUMMARY",
		"Total Personal Balance",
		"Total Shared Balance",
		"Ending balance",
		"HIGHLIGHTS",
		"MONEY IN",
		"MONEY OUT",
		"PERSONAL POCKETS",
		"SHARED POCKETS",
		"Pocket Name",
		"Currency",
		"From last month",
		"Page ",
		"Showing IDR transaction from",
	}

	for _, keyword := range skipContains {
		if strings.Contains(line, keyword) {
			return true
		}
	}

	return false
}

func (p *JagoParser) removeNoise(parts []string) []string {
	clean := make([]string, 0, len(parts))

	for _, part := range parts {
		part = core.CompactLine(part)
		if part == "" || p.shouldSkipLine(part) {
			continue
		}

		switch part {
		case "Pocket ID",
			"Pocket is created on",
			"Previous Balance",
			"Total Incoming",
			"Total Outgoing",
			"Closing Balance":
			continue
		}

		if strings.HasPrefix(part, "Latest Balance per") {
			continue
		}

		clean = append(clean, part)
	}

	return clean
}

// findTransactionStartIndex finds where transactions actually begin.
// Looks for "Amount" followed by "Balance" or the actual start of transactions.
func (p *JagoParser) findTransactionStartIndex(lines []string) int {
	for i := 0; i < len(lines)-1; i++ {

		if lines[i] == "Main Pocket" && strings.HasPrefix(lines[i+1], "Pocket ID") {
			return i
		}

		if p.dateRegex.MatchString(lines[i]) {

			return i
		}
	}
	return -1
}

func (p *JagoParser) isPocketHeader(line string) bool {
	return p.knownPocketName[line]
}

func (p *JagoParser) isHardStop(line string) bool {
	return strings.HasPrefix(line, "CURRENCY EXCHANGE RATE") ||
		strings.HasPrefix(line, "DISCLAIMER")
}

func (p *JagoParser) looksLikeAccountInfo(value string) bool {
	return p.accountRegex.MatchString(value)
}

func isJagoInterestLike(value string) bool {
	return value == "Interest" || value == "Tax on Interest"
}
