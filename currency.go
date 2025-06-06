// Copyright (c) 2024, Frank Mueller / Tideland
// All rights reserved.

package bcd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Currency-specific errors.
var (
	ErrUnknownCurrency   = errors.New("unknown currency code")
	ErrCurrencyMismatch  = errors.New("currency mismatch")
	ErrInvalidAmount     = errors.New("invalid currency amount")
)

// CurrencyInfo contains information about a currency.
type CurrencyInfo struct {
	Code         string // ISO 4217 currency code
	NumericCode  string // ISO 4217 numeric code
	DecimalPlaces int   // Number of decimal places
	Symbol       string // Currency symbol
	Name         string // Currency name
}

// currencyData holds information about supported currencies.
var currencyData = map[string]CurrencyInfo{
	// Major currencies
	"USD": {"USD", "840", 2, "$", "US Dollar"},
	"EUR": {"EUR", "978", 2, "€", "Euro"},
	"GBP": {"GBP", "826", 2, "£", "British Pound"},
	"JPY": {"JPY", "392", 0, "¥", "Japanese Yen"},
	"CHF": {"CHF", "756", 2, "Fr", "Swiss Franc"},
	"CAD": {"CAD", "124", 2, "$", "Canadian Dollar"},
	"AUD": {"AUD", "036", 2, "$", "Australian Dollar"},
	"NZD": {"NZD", "554", 2, "$", "New Zealand Dollar"},
	"CNY": {"CNY", "156", 2, "¥", "Chinese Yuan"},
	"INR": {"INR", "356", 2, "₹", "Indian Rupee"},
	"KRW": {"KRW", "410", 0, "₩", "South Korean Won"},
	"MXN": {"MXN", "484", 2, "$", "Mexican Peso"},
	"BRL": {"BRL", "986", 2, "R$", "Brazilian Real"},
	"RUB": {"RUB", "643", 2, "₽", "Russian Ruble"},
	"ZAR": {"ZAR", "710", 2, "R", "South African Rand"},
	"SEK": {"SEK", "752", 2, "kr", "Swedish Krona"},
	"NOK": {"NOK", "578", 2, "kr", "Norwegian Krone"},
	"DKK": {"DKK", "208", 2, "kr", "Danish Krone"},
	"PLN": {"PLN", "985", 2, "zł", "Polish Złoty"},
	"THB": {"THB", "764", 2, "฿", "Thai Baht"},
	"SGD": {"SGD", "702", 2, "$", "Singapore Dollar"},
	"HKD": {"HKD", "344", 2, "$", "Hong Kong Dollar"},
	"ILS": {"ILS", "376", 2, "₪", "Israeli Shekel"},
	"PHP": {"PHP", "608", 2, "₱", "Philippine Peso"},
	"CZK": {"CZK", "203", 2, "Kč", "Czech Koruna"},
	"HUF": {"HUF", "348", 2, "Ft", "Hungarian Forint"},
	"AED": {"AED", "784", 2, "د.إ", "UAE Dirham"},
	"SAR": {"SAR", "682", 2, "﷼", "Saudi Riyal"},
	"MYR": {"MYR", "458", 2, "RM", "Malaysian Ringgit"},
	"IDR": {"IDR", "360", 2, "Rp", "Indonesian Rupiah"},
	"TRY": {"TRY", "949", 2, "₺", "Turkish Lira"},
	"TWD": {"TWD", "901", 2, "$", "Taiwan Dollar"},
	"VND": {"VND", "704", 0, "₫", "Vietnamese Dong"},
	"CLF": {"CLF", "990", 4, "UF", "Chilean Unit of Account"},
	
	// Cryptocurrencies (unofficial codes)
	"BTC": {"BTC", "---", 8, "₿", "Bitcoin"},
	"ETH": {"ETH", "---", 8, "Ξ", "Ethereum"},
	
	// Precious metals
	"XAU": {"XAU", "959", 2, "oz", "Gold (troy ounce)"},
	"XAG": {"XAG", "961", 2, "oz", "Silver (troy ounce)"},
	"XPT": {"XPT", "962", 2, "oz", "Platinum (troy ounce)"},
	"XPD": {"XPD", "964", 2, "oz", "Palladium (troy ounce)"},
}

// Currency represents a monetary amount in a specific currency.
type Currency struct {
	amount *BCD
	info   CurrencyInfo
}

// NewCurrency creates a new Currency from a string amount and currency code.
func NewCurrency(amount string, code string) (*Currency, error) {
	code = strings.ToUpper(code)
	info, ok := currencyData[code]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownCurrency, code)
	}

	bcd, err := New(amount)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAmount, err)
	}

	// Round to currency's decimal places
	bcd = bcd.Round(info.DecimalPlaces, RoundHalfEven)

	return &Currency{
		amount: bcd,
		info:   info,
	}, nil
}

// NewCurrencyFromInt creates a Currency from an integer amount in minor units.
func NewCurrencyFromInt(minorUnits int64, code string) (*Currency, error) {
	code = strings.ToUpper(code)
	info, ok := currencyData[code]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownCurrency, code)
	}

	bcd := NewFromInt(minorUnits)
	
	// Convert from minor units to major units
	if info.DecimalPlaces > 0 {
		divisor := NewFromInt(1)
		for range info.DecimalPlaces {
			divisor = divisor.Mul(NewFromInt(10))
		}
		var err error
		bcd, err = bcd.Div(divisor, info.DecimalPlaces, RoundHalfEven)
		if err != nil {
			return nil, err
		}
	}

	return &Currency{
		amount: bcd,
		info:   info,
	}, nil
}

// NewCurrencyFromFloat creates a Currency from a float amount.
func NewCurrencyFromFloat(amount float64, code string) (*Currency, error) {
	code = strings.ToUpper(code)
	info, ok := currencyData[code]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownCurrency, code)
	}

	bcd, err := NewFromFloat(amount, info.DecimalPlaces)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAmount, err)
	}

	return &Currency{
		amount: bcd,
		info:   info,
	}, nil
}

// ParseCurrency parses a formatted currency string like "$1,234.56" or "EUR 1.234,56".
func ParseCurrency(s string) (*Currency, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrInvalidAmount
	}

	// Try to find currency code or symbol
	var code string
	var amountStr string

	// Check for ISO code (3 uppercase letters)
	re := regexp.MustCompile(`\b([A-Z]{3})\b`)
	if matches := re.FindStringSubmatch(s); len(matches) > 1 {
		potentialCode := matches[1]
		if _, ok := currencyData[potentialCode]; ok {
			code = potentialCode
			amountStr = strings.Replace(s, potentialCode, "", 1)
		}
	}

	// If no code found, check for symbols - prioritize unique symbols
	if code == "" {
		// First pass: look for unique symbols
		uniqueSymbols := map[string]string{
			"€": "EUR", "£": "GBP", "₹": "INR", "₩": "KRW",
			"R$": "BRL", "₽": "RUB", "zł": "PLN", "฿": "THB", "₪": "ILS",
			"₱": "PHP", "Kč": "CZK", "Ft": "HUF", "₫": "VND", "₺": "TRY",
			"₿": "BTC", "Ξ": "ETH",
		}
		
		for symbol, currCode := range uniqueSymbols {
			if strings.Contains(s, symbol) {
				code = currCode
				amountStr = strings.Replace(s, symbol, "", 1)
				break
			}
		}
		
		// Second pass: check for ¥ which could be JPY or CNY
		if code == "" && strings.Contains(s, "¥") {
			// Default to JPY for ¥ symbol (more common in international usage)
			code = "JPY"
			amountStr = strings.Replace(s, "¥", "", 1)
		}
		
		// Third pass: check for $ which could be multiple currencies
		if code == "" && strings.Contains(s, "$") {
			// Default to USD for $ symbol
			code = "USD"
			amountStr = strings.Replace(s, "$", "", 1)
		}
		
		// Fourth pass: check for other ambiguous symbols
		if code == "" {
			ambiguousSymbols := map[string]string{
				"kr": "SEK", // Could be SEK, NOK, or DKK - default to SEK
				"Fr": "CHF",
			}
			
			for symbol, currCode := range ambiguousSymbols {
				if strings.Contains(s, symbol) {
					code = currCode
					amountStr = strings.Replace(s, symbol, "", 1)
					break
				}
			}
		}
	}

	if code == "" {
		return nil, fmt.Errorf("%w: no currency code or symbol found", ErrInvalidAmount)
	}

	// Clean amount string
	amountStr = strings.TrimSpace(amountStr)
	
	// Handle negative amounts in parentheses
	negative := false
	if strings.HasPrefix(amountStr, "(") && strings.HasSuffix(amountStr, ")") {
		negative = true
		amountStr = amountStr[1 : len(amountStr)-1]
	}

	// Remove thousands separators
	// First, handle special separators like apostrophes (Swiss format)
	amountStr = strings.ReplaceAll(amountStr, "'", "")
	
	// For currencies with no decimal places (like JPY), commas are always thousands separators
	info, hasInfo := currencyData[code]
	
	if hasInfo && info.DecimalPlaces == 0 {
		// No decimal places - all commas and dots are thousands separators
		amountStr = strings.ReplaceAll(amountStr, ",", "")
		amountStr = strings.ReplaceAll(amountStr, ".", "")
	} else {
		// Detect decimal separator (last occurrence of . or ,)
		lastDot := strings.LastIndex(amountStr, ".")
		lastComma := strings.LastIndex(amountStr, ",")

		if lastDot > lastComma {
			// Period is decimal separator
			amountStr = strings.ReplaceAll(amountStr, ",", "")
		} else if lastComma > lastDot {
			// Comma is decimal separator
			amountStr = strings.ReplaceAll(amountStr, ".", "")
			amountStr = strings.Replace(amountStr, ",", ".", 1)
		} else if lastDot == -1 && lastComma == -1 {
			// No decimal separator
		} else if lastDot >= 0 {
			// Only dots, check if it's a thousands separator
			parts := strings.Split(amountStr, ".")
			if len(parts) == 2 && len(parts[1]) != 2 && len(parts[1]) != 3 {
				// Dot is decimal separator
			} else if len(parts) > 2 {
				// Multiple dots, they are thousands separators
				amountStr = strings.ReplaceAll(amountStr, ".", "")
			}
		} else {
			// Only commas, check if it's a thousands separator
			parts := strings.Split(amountStr, ",")
			if len(parts) == 2 && (len(parts[1]) == 2 || len(parts[1]) == 3) {
				// Comma is decimal separator
				amountStr = strings.Replace(amountStr, ",", ".", 1)
			} else {
				// Comma is thousands separator
				amountStr = strings.ReplaceAll(amountStr, ",", "")
			}
		}
	}

	// Apply negative sign
	if negative {
		amountStr = "-" + amountStr
	}

	return NewCurrency(amountStr, code)
}

// Amount returns a copy of the underlying BCD amount.
func (c *Currency) Amount() *BCD {
	return c.amount.Copy()
}

// Code returns the ISO 4217 currency code.
func (c *Currency) Code() string {
	return c.info.Code
}

// Symbol returns the currency symbol.
func (c *Currency) Symbol() string {
	return c.info.Symbol
}

// Name returns the currency name.
func (c *Currency) Name() string {
	return c.info.Name
}

// DecimalPlaces returns the standard number of decimal places for this currency.
func (c *Currency) DecimalPlaces() int {
	return c.info.DecimalPlaces
}

// String returns the currency as a string with symbol.
func (c *Currency) String() string {
	return c.Format(true, false)
}

// Format formats the currency with various options.
func (c *Currency) Format(includeSymbol, includeCode bool) string {
	amountStr := c.amount.String()
	
	// Split into integer and decimal parts
	parts := strings.Split(amountStr, ".")
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) > 1 {
		decimalPart = parts[1]
	}

	// Handle negative
	negative := strings.HasPrefix(integerPart, "-")
	if negative {
		integerPart = integerPart[1:]
	}

	// Ensure proper decimal places
	if c.info.DecimalPlaces > 0 {
		if decimalPart == "" {
			decimalPart = strings.Repeat("0", c.info.DecimalPlaces)
		} else if len(decimalPart) < c.info.DecimalPlaces {
			decimalPart += strings.Repeat("0", c.info.DecimalPlaces-len(decimalPart))
		}
		amountStr = integerPart + "." + decimalPart
	} else {
		amountStr = integerPart
	}

	// Build final string
	var result strings.Builder
	
	if negative {
		result.WriteString("-")
	}
	
	if includeSymbol && c.info.Symbol != "" {
		result.WriteString(c.info.Symbol)
	}
	
	result.WriteString(amountStr)
	
	if includeCode {
		if includeSymbol {
			result.WriteString(" ")
		}
		result.WriteString(c.info.Code)
	}

	return result.String()
}

// FormatWithSeparators formats the currency with thousands separators.
func (c *Currency) FormatWithSeparators(includeSymbol, includeCode bool) string {
	amountStr := c.amount.String()
	
	// Split into integer and decimal parts
	parts := strings.Split(amountStr, ".")
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) > 1 {
		decimalPart = parts[1]
	}

	// Handle negative
	negative := strings.HasPrefix(integerPart, "-")
	if negative {
		integerPart = integerPart[1:]
	}

	// Add thousands separators
	if len(integerPart) > 3 {
		var result []string
		for i := len(integerPart); i > 0; i -= 3 {
			start := i - 3
			if start < 0 {
				start = 0
			}
			result = append([]string{integerPart[start:i]}, result...)
		}
		integerPart = strings.Join(result, ",")
	}

	// Ensure proper decimal places
	if c.info.DecimalPlaces > 0 {
		if decimalPart == "" {
			decimalPart = strings.Repeat("0", c.info.DecimalPlaces)
		} else if len(decimalPart) < c.info.DecimalPlaces {
			decimalPart += strings.Repeat("0", c.info.DecimalPlaces-len(decimalPart))
		}
		amountStr = integerPart + "." + decimalPart
	} else {
		amountStr = integerPart
	}

	// Build final string
	var result strings.Builder
	
	if negative {
		result.WriteString("-")
	}
	
	if includeSymbol && c.info.Symbol != "" {
		result.WriteString(c.info.Symbol)
	}
	
	result.WriteString(amountStr)
	
	if includeCode {
		if includeSymbol {
			result.WriteString(" ")
		}
		result.WriteString(c.info.Code)
	}

	return result.String()
}

// ToMinorUnits converts the currency to its minor units (e.g., cents).
func (c *Currency) ToMinorUnits() (int64, error) {
	if c.info.DecimalPlaces == 0 {
		return c.amount.ToInt64()
	}

	// Multiply by 10^decimalPlaces
	multiplier := NewFromInt(1)
	for range c.info.DecimalPlaces {
		multiplier = multiplier.Mul(NewFromInt(10))
	}

	result := c.amount.Mul(multiplier)
	return result.ToInt64()
}

// Add adds two currency amounts of the same currency.
func (c *Currency) Add(other *Currency) (*Currency, error) {
	if c.info.Code != other.info.Code {
		return nil, fmt.Errorf("%w: %s != %s", ErrCurrencyMismatch, c.info.Code, other.info.Code)
	}

	return &Currency{
		amount: c.amount.Add(other.amount),
		info:   c.info,
	}, nil
}

// Sub subtracts another currency amount from this one.
func (c *Currency) Sub(other *Currency) (*Currency, error) {
	if c.info.Code != other.info.Code {
		return nil, fmt.Errorf("%w: %s != %s", ErrCurrencyMismatch, c.info.Code, other.info.Code)
	}

	return &Currency{
		amount: c.amount.Sub(other.amount),
		info:   c.info,
	}, nil
}

// Mul multiplies the currency by a scalar value.
func (c *Currency) Mul(factor *BCD) *Currency {
	result := c.amount.Mul(factor)
	// Round to currency's decimal places
	result = result.Round(c.info.DecimalPlaces, RoundHalfEven)
	
	return &Currency{
		amount: result,
		info:   c.info,
	}
}

// MulInt64 multiplies the currency by an integer.
func (c *Currency) MulInt64(factor int64) *Currency {
	return c.Mul(NewFromInt(factor))
}

// MulFloat64 multiplies the currency by a float.
func (c *Currency) MulFloat64(factor float64) (*Currency, error) {
	bcd, err := NewFromFloat(factor, 10) // Use high precision for factors
	if err != nil {
		return nil, err
	}
	return c.Mul(bcd), nil
}

// Div divides the currency by a scalar value.
func (c *Currency) Div(divisor *BCD) (*Currency, error) {
	if divisor.IsZero() {
		return nil, ErrDivisionByZero
	}

	result, err := c.amount.Div(divisor, c.info.DecimalPlaces+4, RoundHalfEven)
	if err != nil {
		return nil, err
	}
	
	// Round to currency's decimal places
	result = result.Round(c.info.DecimalPlaces, RoundHalfEven)
	
	return &Currency{
		amount: result,
		info:   c.info,
	}, nil
}

// DivInt64 divides the currency by an integer.
func (c *Currency) DivInt64(divisor int64) (*Currency, error) {
	return c.Div(NewFromInt(divisor))
}

// DivFloat64 divides the currency by a float.
func (c *Currency) DivFloat64(divisor float64) (*Currency, error) {
	bcd, err := NewFromFloat(divisor, 10) // Use high precision for divisors
	if err != nil {
		return nil, err
	}
	return c.Div(bcd)
}

// Allocate distributes the currency amount according to the given ratios.
// The sum of all allocated amounts equals the original amount (no pennies lost).
func (c *Currency) Allocate(ratios []int) ([]*Currency, error) {
	if len(ratios) == 0 {
		return nil, errors.New("ratios cannot be empty")
	}

	// Calculate total ratio
	totalRatio := 0
	for _, r := range ratios {
		if r < 0 {
			return nil, errors.New("ratios must be non-negative")
		}
		totalRatio += r
	}

	if totalRatio == 0 {
		return nil, errors.New("sum of ratios must be positive")
	}

	// Special case: if all amounts are zero
	if c.amount.IsZero() {
		results := make([]*Currency, len(ratios))
		for i := range results {
			results[i] = &Currency{
				amount: Zero(),
				info:   c.info,
			}
		}
		return results, nil
	}

	// Calculate initial allocations
	results := make([]*Currency, len(ratios))
	allocated := Zero()
	
	for i, ratio := range ratios {
		if ratio == 0 {
			results[i] = &Currency{
				amount: Zero(),
				info:   c.info,
			}
			continue
		}

		// Calculate this allocation: amount * ratio / totalRatio
		ratioFactor := NewFromInt(int64(ratio))
		totalFactor := NewFromInt(int64(totalRatio))
		
		proportion, err := ratioFactor.Div(totalFactor, 10, RoundHalfEven)
		if err != nil {
			return nil, err
		}
		
		allocAmount := c.amount.Mul(proportion)
		allocAmount = allocAmount.Round(c.info.DecimalPlaces, RoundDown)
		
		results[i] = &Currency{
			amount: allocAmount,
			info:   c.info,
		}
		
		allocated = allocated.Add(allocAmount)
	}

	// Distribute any remainder due to rounding
	remainder := c.amount.Sub(allocated)
	
	// Add remainder to the largest allocation
	if !remainder.IsZero() {
		largestIdx := 0
		largestRatio := ratios[0]
		for i := 1; i < len(ratios); i++ {
			if ratios[i] > largestRatio {
				largestIdx = i
				largestRatio = ratios[i]
			}
		}
		
		results[largestIdx].amount = results[largestIdx].amount.Add(remainder)
	}

	return results, nil
}

// Split evenly divides the currency amount into n parts.
func (c *Currency) Split(n int) ([]*Currency, error) {
	if n <= 0 {
		return nil, errors.New("number of parts must be positive")
	}

	ratios := make([]int, n)
	for i := range ratios {
		ratios[i] = 1
	}

	return c.Allocate(ratios)
}

// IsZero returns true if the amount is zero.
func (c *Currency) IsZero() bool {
	return c.amount.IsZero()
}

// IsNegative returns true if the amount is negative.
func (c *Currency) IsNegative() bool {
	return c.amount.IsNegative()
}

// IsPositive returns true if the amount is positive.
func (c *Currency) IsPositive() bool {
	return c.amount.IsPositive()
}

// Abs returns the absolute value of the currency.
func (c *Currency) Abs() *Currency {
	return &Currency{
		amount: c.amount.Abs(),
		info:   c.info,
	}
}

// Neg returns the negation of the currency.
func (c *Currency) Neg() *Currency {
	return &Currency{
		amount: c.amount.Neg(),
		info:   c.info,
	}
}

// Cmp compares two currency amounts.
// Returns -1 if c < other, 0 if c == other, 1 if c > other.
func (c *Currency) Cmp(other *Currency) (int, error) {
	if c.info.Code != other.info.Code {
		return 0, fmt.Errorf("%w: %s != %s", ErrCurrencyMismatch, c.info.Code, other.info.Code)
	}
	return c.amount.Cmp(other.amount), nil
}

// Equal returns true if two currency amounts are equal.
func (c *Currency) Equal(other *Currency) bool {
	if c.info.Code != other.info.Code {
		return false
	}
	return c.amount.Equal(other.amount)
}

// GetCurrencyInfo returns information about a currency code.
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