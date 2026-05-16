package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// jagoMonthMap maps abbreviated month names (used by Bank Jago) to month numbers.
var jagoMonthMap = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "mei": 5,
	"jun": 6, "jul": 7, "aug": 8, "agu": 8, "sep": 9,
	"oct": 10, "okt": 10, "nov": 11, "dec": 12, "des": 12,
}

// ParseJagoDate converts Bank Jago date format "d MMM YYYY" → "YYYY-MM-DD".
// Returns the original string unchanged if parsing fails.
func ParseJagoDate(date string) string {
	parts := strings.Fields(strings.TrimSpace(date))
	if len(parts) != 3 {
		return date
	}

	month, ok := jagoMonthMap[strings.ToLower(parts[1])]
	if !ok {
		return date
	}

	t, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%02d-%02d",
		parts[2], month, mustAtoi(parts[0])))
	if err != nil {
		return date
	}

	return t.Format("2006-01-02")
}

// ParseBCADate converts BCA date format "DD/MM/YYYY" → "YYYY-MM-DD".
// Returns the original string unchanged if parsing fails.
func ParseBCADate(date string) string {
	t, err := time.Parse("02/01/2006", strings.TrimSpace(date))
	if err != nil {
		return date
	}
	return t.Format("2006-01-02")
}

func mustAtoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// CompactLine trims space and collapses multiple spaces into one.
func CompactLine(line string) string {
	line = strings.TrimSpace(line)
	return strings.Join(strings.Fields(line), " ")
}

// ParseIDR parses an IDR currency string into a Money struct.
func ParseIDR(value string) Money {
	value = strings.TrimSpace(value)

	return Money{
		Currency:	"IDR",
		Display:	formatIDR(value),
		Value:		parseIDRToInt(value),
	}
}

func formatIDR(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}

	if strings.HasPrefix(value, "+Rp") ||
		strings.HasPrefix(value, "-Rp") ||
		strings.HasPrefix(value, "Rp") {
		return value
	}

	if strings.HasPrefix(value, "+") {
		return "+Rp" + strings.TrimPrefix(value, "+")
	}

	if strings.HasPrefix(value, "-") {
		return "-Rp" + strings.TrimPrefix(value, "-")
	}

	return "Rp" + value
}

func parseIDRToInt(value string) int64 {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "Rp", "")
	value = strings.ReplaceAll(value, ".", "")

	sign := int64(1)

	if after, ok := strings.CutPrefix(value, "+"); ok {
		value = after
	}

	if strings.HasPrefix(value, "-") {
		sign = -1
		value = strings.TrimPrefix(value, "-")
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return parsed * sign
}

// ParseBCAMoney parses BCA-style money values like "3,528,964.00 CR" or "108400.00".
// BCA uses comma as thousands separator and dot as decimal separator.
// Returns a Money struct. The sign is determined by the CR/DB suffix if present,
// or explicitly by the isCredit parameter.
func ParseBCAMoney(value string, isCredit bool) Money {
	value = strings.TrimSpace(value)
	if value == "" {
		return Money{Currency: "IDR"}
	}

	if strings.HasSuffix(value, " CR") {
		isCredit = true
		value = strings.TrimSuffix(value, " CR")
	} else if strings.HasSuffix(value, " DB") {
		isCredit = false
		value = strings.TrimSuffix(value, " DB")
	}

	value = strings.TrimSpace(value)

	cleaned := strings.ReplaceAll(value, ",", "")

	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return Money{Currency: "IDR"}
	}

	intVal := int64(f)
	if !isCredit {
		intVal = -intVal
	}

	display := formatIDR(strconv.FormatInt(intVal, 10))

	return Money{
		Currency:	"IDR",
		Display:	display,
		Value:		intVal,
	}
}

// ParseBCABalance parses BCA-style balance values (always positive).
func ParseBCABalance(value string) Money {
	value = strings.TrimSpace(value)
	if value == "" {
		return Money{Currency: "IDR"}
	}

	cleaned := strings.ReplaceAll(value, ",", "")

	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return Money{Currency: "IDR"}
	}

	intVal := int64(f)
	display := "Rp" + value

	return Money{
		Currency:	"IDR",
		Display:	display,
		Value:		intVal,
	}
}
