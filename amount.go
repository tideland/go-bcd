// Tideland Go BCD
//
// Copyright (C) 2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package bcd

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// Amount errors.
var (
	ErrUnknownCurrency  = fmt.Errorf("unknown currency code")
	ErrCurrencyMismatch = fmt.Errorf("currency mismatch")
	ErrInvalidAmount    = fmt.Errorf("invalid amount")
)

// CurrencyInfo contains information about a currency.
type CurrencyInfo struct {
	Code          string
	NumericCode   string
	DecimalPlaces int
	Symbol        string
	Name          string
}

// currencyData maps ISO 4217 currency codes to their information.
var currencyData = map[string]CurrencyInfo{
	// Major currencies
	"USD": {"USD", "840", 2, "$", "US Dollar"},
	"EUR": {"EUR", "978", 2, "€", "Euro"},
	"GBP": {"GBP", "826", 2, "£", "British Pound"},
	"JPY": {"JPY", "392", 0, "¥", "Japanese Yen"},
	"CHF": {"CHF", "756", 2, "Fr", "Swiss Franc"},
	"CAD": {"CAD", "124", 2, "C$", "Canadian Dollar"},
	"AUD": {"AUD", "036", 2, "A$", "Australian Dollar"},
	"NZD": {"NZD", "554", 2, "NZ$", "New Zealand Dollar"},
	"CNY": {"CNY", "156", 2, "¥", "Chinese Yuan"},
	"INR": {"INR", "356", 2, "₹", "Indian Rupee"},
	"KRW": {"KRW", "410", 0, "₩", "South Korean Won"},
	"SGD": {"SGD", "702", 2, "S$", "Singapore Dollar"},
	"HKD": {"HKD", "344", 2, "HK$", "Hong Kong Dollar"},
	"SEK": {"SEK", "752", 2, "kr", "Swedish Krona"},
	"NOK": {"NOK", "578", 2, "kr", "Norwegian Krone"},
	"DKK": {"DKK", "208", 2, "kr", "Danish Krone"},
	"PLN": {"PLN", "985", 2, "zł", "Polish Zloty"},
	"CZK": {"CZK", "203", 2, "Kč", "Czech Koruna"},
	"HUF": {"HUF", "348", 2, "Ft", "Hungarian Forint"},
	"RUB": {"RUB", "643", 2, "₽", "Russian Ruble"},
	"TRY": {"TRY", "949", 2, "₺", "Turkish Lira"},
	"BRL": {"BRL", "986", 2, "R$", "Brazilian Real"},
	"MXN": {"MXN", "484", 2, "Mex$", "Mexican Peso"},
	"ZAR": {"ZAR", "710", 2, "R", "South African Rand"},
	"AED": {"AED", "784", 2, "د.إ", "UAE Dirham"},
	"SAR": {"SAR", "682", 2, "﷼", "Saudi Riyal"},
	"THB": {"THB", "764", 2, "฿", "Thai Baht"},
	"MYR": {"MYR", "458", 2, "RM", "Malaysian Ringgit"},
	"IDR": {"IDR", "360", 2, "Rp", "Indonesian Rupiah"},
	"PHP": {"PHP", "608", 2, "₱", "Philippine Peso"},
	"VND": {"VND", "704", 0, "₫", "Vietnamese Dong"},
	"ILS": {"ILS", "376", 2, "₪", "Israeli Shekel"},

	// Cryptocurrencies
	"BTC": {"BTC", "XBT", 8, "₿", "Bitcoin"},
	"ETH": {"ETH", "ETH", 8, "Ξ", "Ethereum"},

	// Precious metals
	"XAU": {"XAU", "959", 2, "Au", "Gold (ounce)"},
	"XAG": {"XAG", "961", 2, "Ag", "Silver (ounce)"},
	"XPT": {"XPT", "962", 2, "Pt", "Platinum (ounce)"},
	"XPD": {"XPD", "964", 2, "Pd", "Palladium (ounce)"},
}

// Amount represents a monetary amount in a specific currency.
type Amount struct {
	amount *BCD
	info   CurrencyInfo
}

// NewAmount creates an Amount from any numeric type.
func NewAmount[T any](value T, code string, opts ...Option) (*Amount, error) {
	code = strings.ToUpper(code)
	info, ok := currencyData[code]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownCurrency, code)
	}

	var amount *BCD
	var err error

	// Handle *BCD type specially
	if v, ok := any(value).(*BCD); ok {
		amount = v.Copy()
	} else {
		// For all other types, try to create BCD with currency's decimal places
		currencyOpts := append([]Option{WithScale(info.DecimalPlaces)}, opts...)

		// Use reflection to check if type can be handled by New
		switch any(value).(type) {
		case string, int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			// These types are supported by New
			amount, err = newFromAny(value, currencyOpts...)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrInvalidAmount, err)
			}
		default:
			return nil, fmt.Errorf("%w: unsupported type %T", ErrInvalidAmount, value)
		}
	}

	// Round to currency's decimal places
	amount = amount.Round(info.DecimalPlaces, RoundHalfEven)

	return &Amount{
		amount: amount,
		info:   info,
	}, nil
}

// newFromAny is a helper that calls New with the correct type parameter
func newFromAny(value any, opts ...Option) (*BCD, error) {
	switch v := value.(type) {
	case string:
		return New(v, opts...)
	case int:
		return New(v, opts...)
	case int8:
		return New(v, opts...)
	case int16:
		return New(v, opts...)
	case int32:
		return New(v, opts...)
	case int64:
		return New(v, opts...)
	case uint:
		return New(v, opts...)
	case uint8:
		return New(v, opts...)
	case uint16:
		return New(v, opts...)
	case uint32:
		return New(v, opts...)
	case uint64:
		return New(v, opts...)
	case float32:
		return New(v, opts...)
	case float64:
		return New(v, opts...)
	default:
		return nil, fmt.Errorf("unsupported type: %T", value)
	}
}

// MustNewAmount creates an Amount and panics on error.
func MustNewAmount[T any](value T, code string, opts ...Option) *Amount {
	c, err := NewAmount(value, code, opts...)
	if err != nil {
		panic(fmt.Sprintf("bcd.MustNewAmount: %v", err))
	}
	return c
}

// NewAmountMinor creates an Amount from minor units (cents, pence, etc.).
// This function only accepts integer types.
type IntegerType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func NewAmountMinor[T IntegerType](minorUnits T, code string) (*Amount, error) {
	code = strings.ToUpper(code)
	info, ok := currencyData[code]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownCurrency, code)
	}

	// Convert to int64 for processing
	var units int64
	switch v := any(minorUnits).(type) {
	case int:
		units = int64(v)
	case int8:
		units = int64(v)
	case int16:
		units = int64(v)
	case int32:
		units = int64(v)
	case int64:
		units = v
	case uint:
		if v > math.MaxInt64 {
			return nil, ErrOverflow
		}
		units = int64(v)
	case uint8:
		units = int64(v)
	case uint16:
		units = int64(v)
	case uint32:
		units = int64(v)
	case uint64:
		if v > math.MaxInt64 {
			return nil, ErrOverflow
		}
		units = int64(v)
	}

	amount := fromInt64(units)

	// Convert from minor units to major units
	if info.DecimalPlaces > 0 {
		divisor := fromInt64(1)
		for range info.DecimalPlaces {
			divisor = divisor.Mul(fromInt64(10))
		}
		var err error
		amount, err = amount.Div(divisor, info.DecimalPlaces, RoundHalfEven)
		if err != nil {
			return nil, err
		}
	}

	return &Amount{
		amount: amount,
		info:   info,
	}, nil
}

// MustNewAmountMinor creates an Amount from minor units and panics on error.
func MustNewAmountMinor[T IntegerType](minorUnits T, code string) *Amount {
	c, err := NewAmountMinor(minorUnits, code)
	if err != nil {
		panic(fmt.Sprintf("bcd.MustNewAmountMinor: %v", err))
	}
	return c
}

// ParseAmount parses a formatted currency string.
func ParseAmount(s string) (*Amount, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrInvalidAmount
	}

	// Regular expressions for different formats
	// Use slice of structs to ensure deterministic order
	symbolPatterns := []struct {
		code    string
		pattern string
	}{
		{"USD", `^\$`},
		{"EUR", `^€`},
		{"GBP", `^£`},
		{"JPY", `^¥|^￥`}, // Check JPY before CNY for deterministic behavior
		{"INR", `^₹`},
		{"KRW", `^₩`},
		{"BTC", `^₿`},
		{"CHF", `^Fr\.?`},
		{"SEK", `^kr`},
		{"NOK", `^kr`},
		{"DKK", `^kr`},
		{"CNY", `^¥|^￥`}, // CNY after JPY
	}

	// Try to identify currency by symbol
	var identifiedCode string
	var amountStr string

	for _, sp := range symbolPatterns {
		re := regexp.MustCompile(sp.pattern)
		if re.MatchString(s) {
			identifiedCode = sp.code
			amountStr = re.ReplaceAllString(s, "")
			break
		}
	}

	// If no symbol found, try to find currency code
	if identifiedCode == "" {
		// Check for 3-letter currency code at the beginning or end
		codePattern := regexp.MustCompile(`^([A-Z]{3})\s*(.+)$|^(.+?)\s*([A-Z]{3})$`)
		matches := codePattern.FindStringSubmatch(s)

		if len(matches) > 0 {
			if matches[1] != "" {
				identifiedCode = matches[1]
				amountStr = matches[2]
			} else if matches[4] != "" {
				identifiedCode = matches[4]
				amountStr = matches[3]
			}
		}
	}

	if identifiedCode == "" {
		// Check for accounting format with parentheses
		if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
			// Try to extract currency from within parentheses
			inner := strings.TrimPrefix(s, "(")
			inner = strings.TrimSuffix(inner, ")")

			// Check if it starts with a common symbol
			if strings.HasPrefix(inner, "$") {
				identifiedCode = "USD"
				amountStr = strings.TrimPrefix(inner, "$")
				// Mark as negative (will be handled below)
			} else {
				return nil, fmt.Errorf("%w: cannot identify currency", ErrInvalidAmount)
			}
		} else if strings.HasPrefix(s, "$") {
			// Default to USD if no currency identified and starts with $
			identifiedCode = "USD"
			amountStr = strings.TrimPrefix(s, "$")
		} else {
			return nil, fmt.Errorf("%w: cannot identify currency", ErrInvalidAmount)
		}
	}

	// Clean up amount string
	amountStr = strings.TrimSpace(amountStr)

	// Handle negative amounts in parentheses (accounting format)
	// Also check if the original string had parentheses
	negative := false
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		negative = true
	}
	if strings.HasPrefix(amountStr, "(") && strings.HasSuffix(amountStr, ")") {
		negative = true
		amountStr = strings.TrimPrefix(amountStr, "(")
		amountStr = strings.TrimSuffix(amountStr, ")")
	}

	// Determine format and handle separators
	if strings.Contains(amountStr, ".") && strings.Contains(amountStr, ",") {
		// Both present, determine which is decimal separator
		lastDot := strings.LastIndex(amountStr, ".")
		lastComma := strings.LastIndex(amountStr, ",")

		if lastComma > lastDot {
			// Comma is decimal separator (European format)
			// Remove dots (thousand separators) and convert comma to dot
			amountStr = strings.ReplaceAll(amountStr, ".", "")
			amountStr = strings.Replace(amountStr, ",", ".", 1)
		} else {
			// Dot is decimal separator (US format)
			// Remove commas (thousand separators)
			amountStr = strings.ReplaceAll(amountStr, ",", "")
		}
	} else if strings.Count(amountStr, ",") == 1 && !strings.Contains(amountStr, ".") {
		// Only comma present, check if it's decimal separator
		lastComma := strings.LastIndex(amountStr, ",")
		if lastComma > 0 && len(amountStr)-lastComma-1 <= 2 {
			// Comma is likely decimal separator (e.g., "1234,56")
			amountStr = strings.Replace(amountStr, ",", ".", 1)
		} else {
			// Comma is thousand separator (e.g., "1,234")
			amountStr = strings.ReplaceAll(amountStr, ",", "")
		}
	} else {
		// Remove thousand separators
		amountStr = strings.ReplaceAll(amountStr, ",", "")
		amountStr = strings.ReplaceAll(amountStr, "'", "")
		amountStr = strings.ReplaceAll(amountStr, " ", "")
	}

	// Parse the amount
	amount, err := parseString(amountStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAmount, err)
	}

	if negative {
		amount = amount.Neg()
	}

	// Create currency with identified code
	return NewAmount(amount, identifiedCode)
}

// Amount returns the BCD amount.
func (c *Amount) Amount() *BCD {
	return c.amount.Copy()
}

// Code returns the ISO 4217 currency code.
func (c *Amount) Code() string {
	return c.info.Code
}

// Symbol returns the currency symbol.
func (c *Amount) Symbol() string {
	return c.info.Symbol
}

// Name returns the currency name.
func (c *Amount) Name() string {
	return c.info.Name
}

// DecimalPlaces returns the number of decimal places for the currency.
func (c *Amount) DecimalPlaces() int {
	return c.info.DecimalPlaces
}

// String returns the default string representation with symbol.
func (c *Amount) String() string {
	return c.Format(true, false)
}

// Format formats the currency with various options.
func Format(c *Amount, includeSymbol, includeCode bool) string {
	var sb strings.Builder

	// Handle negative amounts
	isNegative := c.amount.IsNegative()
	absAmount := c.amount.Abs()

	if isNegative {
		sb.WriteByte('-')
	}

	// Add symbol if requested
	if includeSymbol {
		sb.WriteString(c.info.Symbol)
	}

	// Format the amount
	amountStr := absAmount.String()

	// Handle currencies with no decimal places
	if c.info.DecimalPlaces == 0 {
		// Remove decimal point if present
		if idx := strings.Index(amountStr, "."); idx >= 0 {
			amountStr = amountStr[:idx]
		}
	} else {
		// Ensure correct decimal places
		if idx := strings.Index(amountStr, "."); idx >= 0 {
			decimalPart := amountStr[idx+1:]
			if len(decimalPart) < c.info.DecimalPlaces {
				// Pad with zeros
				amountStr += strings.Repeat("0", c.info.DecimalPlaces-len(decimalPart))
			}
		} else {
			// No decimal point, add it
			amountStr += "." + strings.Repeat("0", c.info.DecimalPlaces)
		}
	}

	sb.WriteString(amountStr)

	// Add code if requested
	if includeCode {
		sb.WriteByte(' ')
		sb.WriteString(c.info.Code)
	}

	return sb.String()
}

// Format formats the currency with various options.
func (c *Amount) Format(includeSymbol, includeCode bool) string {
	return Format(c, includeSymbol, includeCode)
}

// FormatWithSeparators formats the currency with custom separators.
func (c *Amount) FormatWithSeparators(separator string, includeSymbol, includeCode bool) string {
	var sb strings.Builder

	// Handle negative amounts
	isNegative := c.amount.IsNegative()
	absAmount := c.amount.Abs()

	if isNegative {
		sb.WriteByte('-')
	}

	// Add symbol if requested
	if includeSymbol {
		sb.WriteString(c.info.Symbol)
	}

	// Format the amount
	amountStr := absAmount.String()

	// Split into integer and decimal parts
	parts := strings.Split(amountStr, ".")
	integerPart := parts[0]

	// Add thousand separators to integer part
	if len(integerPart) > 3 {
		var result []string
		chars := []rune(integerPart)

		for i := len(chars); i > 0; i -= 3 {
			start := i - 3
			if start < 0 {
				start = 0
			}
			result = append([]string{string(chars[start:i])}, result...)
		}

		integerPart = strings.Join(result, separator)
	}

	sb.WriteString(integerPart)

	// Add decimal part if currency has decimal places
	if c.info.DecimalPlaces > 0 {
		sb.WriteByte('.')

		if len(parts) > 1 {
			decimalPart := parts[1]
			// Ensure correct decimal places
			if len(decimalPart) < c.info.DecimalPlaces {
				decimalPart += strings.Repeat("0", c.info.DecimalPlaces-len(decimalPart))
			} else if len(decimalPart) > c.info.DecimalPlaces {
				decimalPart = decimalPart[:c.info.DecimalPlaces]
			}
			sb.WriteString(decimalPart)
		} else {
			sb.WriteString(strings.Repeat("0", c.info.DecimalPlaces))
		}
	}

	// Add code if requested
	if includeCode {
		sb.WriteByte(' ')
		sb.WriteString(c.info.Code)
	}

	return sb.String()
}

// ToMinorUnits converts the currency to its minor units (e.g., cents).
func (c *Amount) ToMinorUnits() (int64, error) {
	if c.info.DecimalPlaces == 0 {
		return c.amount.ToInt64()
	}

	// Multiply by 10^decimalPlaces
	multiplier := fromInt64(1)
	for range c.info.DecimalPlaces {
		multiplier = multiplier.Mul(fromInt64(10))
	}

	minorAmount := c.amount.Mul(multiplier)
	return minorAmount.ToInt64()
}

// Arithmetic operations

// Add adds two currency values of the same currency.
func (c *Amount) Add(other *Amount) (*Amount, error) {
	if c.info.Code != other.info.Code {
		return nil, fmt.Errorf("%w: %s != %s", ErrCurrencyMismatch, c.info.Code, other.info.Code)
	}

	return &Amount{
		amount: c.amount.Add(other.amount),
		info:   c.info,
	}, nil
}

// Sub subtracts two currency amounts of the same currency.
func (c *Amount) Sub(other *Amount) (*Amount, error) {
	if c.info.Code != other.info.Code {
		return nil, fmt.Errorf("%w: %s != %s", ErrCurrencyMismatch, c.info.Code, other.info.Code)
	}

	return &Amount{
		amount: c.amount.Sub(other.amount),
		info:   c.info,
	}, nil
}

// Mul multiplies currency by a BCD factor.
func (c *Amount) Mul(factor *BCD) *Amount {
	result := c.amount.Mul(factor)
	// Round to currency's decimal places
	result = result.Round(c.info.DecimalPlaces, RoundHalfEven)

	return &Amount{
		amount: result,
		info:   c.info,
	}
}

// MulInt64 multiplies the currency by an integer.
func (c *Amount) MulInt64(n int64) *Amount {
	return c.Mul(fromInt64(n))
}

// MulFloat64 multiplies currency by a float.
func (c *Amount) MulFloat64(f float64) (*Amount, error) {
	factor, err := New(f, WithScale(4)) // Use 4 decimal places for factors
	if err != nil {
		return nil, err
	}
	return c.Mul(factor), nil
}

// Div divides currency by a BCD divisor.
func (c *Amount) Div(divisor *BCD) (*Amount, error) {
	if divisor.IsZero() {
		return nil, ErrDivisionByZero
	}

	// Perform division with extra precision
	result, err := c.amount.Div(divisor, c.info.DecimalPlaces+2, RoundHalfEven)
	if err != nil {
		return nil, err
	}

	// Round to currency's decimal places
	result = result.Round(c.info.DecimalPlaces, RoundHalfEven)

	return &Amount{
		amount: result,
		info:   c.info,
	}, nil
}

// DivInt64 divides the currency by an integer.
func (c *Amount) DivInt64(n int64) (*Amount, error) {
	return c.Div(fromInt64(n))
}

// DivFloat64 divides currency by a float.
func (c *Amount) DivFloat64(f float64) (*Amount, error) {
	divisor, err := New(f, WithScale(4))
	if err != nil {
		return nil, err
	}
	return c.Div(divisor)
}

// Allocate distributes the currency amount according to the given ratios.
// The sum of the allocated amounts equals the original amount (no pennies lost).
func (c *Amount) Allocate(ratios []int) ([]*Amount, error) {
	if len(ratios) == 0 {
		return nil, fmt.Errorf("%w: no ratios provided", ErrInvalidAmount)
	}

	// Calculate total ratio
	totalRatio := 0
	for _, ratio := range ratios {
		if ratio < 0 {
			return nil, fmt.Errorf("%w: negative ratio", ErrInvalidAmount)
		}
		totalRatio += ratio
	}

	if totalRatio == 0 {
		return nil, fmt.Errorf("%w: total ratio is zero", ErrInvalidAmount)
	}

	// Convert to minor units for precise allocation
	minorUnits, err := c.ToMinorUnits()
	if err != nil {
		return nil, err
	}

	// Allocate based on ratios
	allocated := make([]int64, len(ratios))
	allocatedSum := int64(0)

	for i, ratio := range ratios {
		share := (minorUnits * int64(ratio)) / int64(totalRatio)
		allocated[i] = share
		allocatedSum += share
	}

	// Distribute remainder
	remainder := minorUnits - allocatedSum

	// Distribute remainder to largest shares first
	if remainder != 0 {
		// Create index array for sorting
		indices := make([]int, len(ratios))
		for i := range indices {
			indices[i] = i
		}

		// Sort indices by ratio (descending)
		for i := 0; i < len(indices)-1; i++ {
			for j := i + 1; j < len(indices); j++ {
				if ratios[indices[j]] > ratios[indices[i]] {
					indices[i], indices[j] = indices[j], indices[i]
				}
			}
		}

		// Distribute remainder
		direction := int64(1)
		if remainder < 0 {
			direction = -1
			remainder = -remainder
		}

		for i := 0; i < int(remainder); i++ {
			idx := indices[i%len(indices)]
			allocated[idx] += direction
		}
	}

	// Convert back to currency
	result := make([]*Amount, len(allocated))
	for i, minorAmount := range allocated {
		result[i], err = NewAmountMinor(minorAmount, c.info.Code)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Split divides the currency amount evenly among n parts.
func (c *Amount) Split(n int) ([]*Amount, error) {
	if n <= 0 {
		return nil, fmt.Errorf("%w: invalid split count", ErrInvalidAmount)
	}

	// Create equal ratios
	ratios := make([]int, n)
	for i := range ratios {
		ratios[i] = 1
	}

	return c.Allocate(ratios)
}

// Comparison operations

// IsZero returns true if the currency amount is zero.
func (c *Amount) IsZero() bool {
	return c.amount.IsZero()
}

// IsNegative returns true if the currency amount is negative.
func (c *Amount) IsNegative() bool {
	return c.amount.IsNegative()
}

// IsPositive returns true if the currency amount is positive.
func (c *Amount) IsPositive() bool {
	return c.amount.IsPositive()
}

// Abs returns the absolute value of the currency.
func (c *Amount) Abs() *Amount {
	return &Amount{
		amount: c.amount.Abs(),
		info:   c.info,
	}
}

// Neg returns the negation of the currency.
func (c *Amount) Neg() *Amount {
	return &Amount{
		amount: c.amount.Neg(),
		info:   c.info,
	}
}

// Cmp compares two currency values.
// Returns -1 if c < other, 0 if c == other, 1 if c > other.
func (c *Amount) Cmp(other *Amount) (int, error) {
	if c.info.Code != other.info.Code {
		return 0, fmt.Errorf("%w: %s != %s", ErrCurrencyMismatch, c.info.Code, other.info.Code)
	}
	return c.amount.Cmp(other.amount), nil
}

// Equal returns true if both currency amounts are equal.
// Returns false if currencies don't match.
func (c *Amount) Equal(other *Amount) bool {
	if c.info.Code != other.info.Code {
		return false
	}
	return c.amount.Equal(other.amount)
}

// GetCurrencyInfo returns the CurrencyInfo for the given code.
func GetCurrencyInfo(code string) (CurrencyInfo, bool) {
	info, ok := currencyData[strings.ToUpper(code)]
	return info, ok
}

// SupportedCurrencies returns a list of all supported currency codes.
func SupportedCurrencies() []string {
	codes := make([]string, 0, len(currencyData))
	for code := range currencyData {
		codes = append(codes, code)
	}
	return codes
}
