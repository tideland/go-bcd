// Copyright (c) 2024, Frank Mueller / Tideland
// All rights reserved.

// Package bcd provides a Binary Coded Decimal (BCD) implementation
// for precise decimal arithmetic, particularly useful for financial
// and currency calculations where floating-point errors are unacceptable.
package bcd

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Common errors for BCD operations.
var (
	ErrDivisionByZero = errors.New("division by zero")
	ErrInvalidFormat  = errors.New("invalid decimal format")
	ErrOverflow       = errors.New("arithmetic overflow")
	ErrPrecisionLoss  = errors.New("precision loss")
)

// RoundingMode defines how to round decimal numbers.
type RoundingMode int

const (
	// RoundDown rounds towards zero (truncate).
	RoundDown RoundingMode = iota
	// RoundUp rounds away from zero.
	RoundUp
	// RoundHalfUp rounds to nearest, ties away from zero.
	RoundHalfUp
	// RoundHalfDown rounds to nearest, ties towards zero.
	RoundHalfDown
	// RoundHalfEven rounds to nearest, ties to even (banker's rounding).
	RoundHalfEven
	// RoundCeiling rounds towards positive infinity.
	RoundCeiling
	// RoundFloor rounds towards negative infinity.
	RoundFloor
)

// BCD represents a decimal number using Binary Coded Decimal encoding.
// Each decimal digit is stored separately to avoid floating-point errors.
type BCD struct {
	// digits stores the decimal digits in little-endian order (least significant first).
	// For example, 123.45 would be stored as [5, 4, 3, 2, 1].
	digits []uint8
	// scale is the number of digits after the decimal point.
	// For 123.45, scale would be 2.
	scale int
	// negative indicates if the number is negative.
	negative bool
}

// New creates a new BCD from a string representation.
func New(s string) (*BCD, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return &BCD{digits: []uint8{0}, scale: 0, negative: false}, nil
	}

	// Handle sign
	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}

	// Validate format
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return nil, ErrInvalidFormat
	}

	// Process integer part
	intPart := strings.TrimLeft(parts[0], "0")
	if intPart == "" {
		intPart = "0"
	}

	// Process decimal part
	var decPart string
	scale := 0
	if len(parts) == 2 {
		decPart = strings.TrimRight(parts[1], "0")
		scale = len(decPart)
	}

	// Validate digits
	allDigits := intPart + decPart
	for _, r := range allDigits {
		if r < '0' || r > '9' {
			return nil, ErrInvalidFormat
		}
	}

	// Create digit array (little-endian)
	digits := make([]uint8, len(allDigits))
	for i := range len(allDigits) {
		digits[i] = uint8(allDigits[len(allDigits)-1-i] - '0')
	}

	// Handle zero
	if len(digits) == 0 || (len(digits) == 1 && digits[0] == 0) {
		return &BCD{digits: []uint8{0}, scale: 0, negative: false}, nil
	}

	// Remove leading zeros
	for len(digits) > 1 && digits[len(digits)-1] == 0 {
		digits = digits[:len(digits)-1]
	}

	return &BCD{
		digits:   digits,
		scale:    scale,
		negative: negative && !isZero(digits),
	}, nil
}

// NewFromInt creates a BCD from an integer.
func NewFromInt(n int64) *BCD {
	if n == 0 {
		return &BCD{digits: []uint8{0}, scale: 0, negative: false}
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []uint8
	for n > 0 {
		digits = append(digits, uint8(n%10))
		n /= 10
	}

	return &BCD{
		digits:   digits,
		scale:    0,
		negative: negative,
	}
}

// NewFromFloat creates a BCD from a float with specified scale.
func NewFromFloat(f float64, scale int) (*BCD, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, ErrInvalidFormat
	}

	// Convert to string with proper precision
	format := fmt.Sprintf("%%.%df", scale)
	s := fmt.Sprintf(format, f)
	return New(s)
}

// Zero returns a BCD representing zero.
func Zero() *BCD {
	return &BCD{digits: []uint8{0}, scale: 0, negative: false}
}

// Copy creates a deep copy of the BCD.
func (b *BCD) Copy() *BCD {
	digits := make([]uint8, len(b.digits))
	copy(digits, b.digits)
	return &BCD{
		digits:   digits,
		scale:    b.scale,
		negative: b.negative,
	}
}

// String returns the string representation of the BCD.
func (b *BCD) String() string {
	if isZero(b.digits) {
		return "0"
	}

	var result strings.Builder

	if b.negative {
		result.WriteByte('-')
	}

	// Determine decimal point position
	intDigits := len(b.digits) - b.scale
	if intDigits <= 0 {
		// Number is less than 1
		result.WriteString("0.")
		for range -intDigits {
			result.WriteByte('0')
		}
		for i := len(b.digits) - 1; i >= 0; i-- {
			result.WriteByte('0' + b.digits[i])
		}
	} else {
		// Write integer part
		for i := len(b.digits) - 1; i >= b.scale; i-- {
			result.WriteByte('0' + b.digits[i])
		}
		// Write decimal part if exists
		if b.scale > 0 {
			result.WriteByte('.')
			for i := b.scale - 1; i >= 0; i-- {
				result.WriteByte('0' + b.digits[i])
			}
		}
	}

	return result.String()
}

// IsZero returns true if the BCD represents zero.
func (b *BCD) IsZero() bool {
	return isZero(b.digits)
}

// IsNegative returns true if the BCD is negative.
func (b *BCD) IsNegative() bool {
	return b.negative
}

// IsPositive returns true if the BCD is positive (greater than zero).
func (b *BCD) IsPositive() bool {
	return !b.negative && !b.IsZero()
}

// Scale returns the number of decimal places.
func (b *BCD) Scale() int {
	return b.scale
}

// Precision returns the total number of significant digits.
func (b *BCD) Precision() int {
	// Remove trailing zeros after decimal point for precision count
	digits := b.digits
	scale := b.scale
	for scale > 0 && len(digits) > 0 && digits[0] == 0 {
		digits = digits[1:]
		scale--
	}
	return len(digits)
}

// Abs returns the absolute value of the BCD.
func (b *BCD) Abs() *BCD {
	result := b.Copy()
	result.negative = false
	return result
}

// Neg returns the negation of the BCD.
func (b *BCD) Neg() *BCD {
	if b.IsZero() {
		return b.Copy()
	}
	result := b.Copy()
	result.negative = !result.negative
	return result
}

// Cmp compares two BCDs.
// Returns -1 if b < other, 0 if b == other, 1 if b > other.
func (b *BCD) Cmp(other *BCD) int {
	// Handle different signs
	if b.negative && !other.negative {
		return -1
	}
	if !b.negative && other.negative {
		return 1
	}

	// Same sign - compare magnitudes
	cmp := compareMagnitudes(b, other)

	// If both negative, reverse the comparison
	if b.negative {
		return -cmp
	}
	return cmp
}

// Equal returns true if two BCDs are equal.
func (b *BCD) Equal(other *BCD) bool {
	return b.Cmp(other) == 0
}

// LessThan returns true if b < other.
func (b *BCD) LessThan(other *BCD) bool {
	return b.Cmp(other) < 0
}

// LessOrEqual returns true if b <= other.
func (b *BCD) LessOrEqual(other *BCD) bool {
	return b.Cmp(other) <= 0
}

// GreaterThan returns true if b > other.
func (b *BCD) GreaterThan(other *BCD) bool {
	return b.Cmp(other) > 0
}

// GreaterOrEqual returns true if b >= other.
func (b *BCD) GreaterOrEqual(other *BCD) bool {
	return b.Cmp(other) >= 0
}

// Add adds two BCDs.
func (b *BCD) Add(other *BCD) *BCD {
	// Handle signs
	if b.negative == other.negative {
		// Same sign - add magnitudes
		result := addMagnitudes(b, other)
		result.negative = b.negative
		return result
	}

	// Different signs - subtract magnitudes
	cmp := compareMagnitudes(b, other)
	if cmp >= 0 {
		result := subtractMagnitudes(b, other)
		result.negative = b.negative
		return result
	}
	result := subtractMagnitudes(other, b)
	result.negative = other.negative
	return result
}

// Sub subtracts other from b.
func (b *BCD) Sub(other *BCD) *BCD {
	return b.Add(other.Neg())
}

// Mul multiplies two BCDs.
func (b *BCD) Mul(other *BCD) *BCD {
	if b.IsZero() || other.IsZero() {
		return Zero()
	}

	// Multiply digits
	resultLen := len(b.digits) + len(other.digits)
	result := make([]uint8, resultLen)

	for i := range b.digits {
		carry := uint8(0)
		for j := range other.digits {
			prod := b.digits[i]*other.digits[j] + result[i+j] + carry
			result[i+j] = prod % 10
			carry = prod / 10
		}
		if carry > 0 {
			result[i+len(other.digits)] += carry
		}
	}

	// Remove leading zeros
	for len(result) > 1 && result[len(result)-1] == 0 {
		result = result[:len(result)-1]
	}

	return &BCD{
		digits:   result,
		scale:    b.scale + other.scale,
		negative: b.negative != other.negative,
	}
}

// Div divides b by other with specified scale and rounding mode.
func (b *BCD) Div(other *BCD, scale int, mode RoundingMode) (*BCD, error) {
	if other.IsZero() {
		return nil, ErrDivisionByZero
	}

	if b.IsZero() {
		return Zero(), nil
	}

	// Scale up dividend for precision
	extraScale := scale + 10 // Extra digits for accurate rounding
	scaledDividend := b.Copy()
	scaledDividend.digits = append(make([]uint8, extraScale), scaledDividend.digits...)
	scaledDividend.scale += extraScale

	// Perform long division
	quotient, _ := divideWithRemainder(scaledDividend, other)
	
	// Adjust scale
	quotient.scale = scaledDividend.scale - other.scale
	quotient.negative = b.negative != other.negative

	// Round to requested scale
	return quotient.Round(scale, mode), nil
}

// DivInt performs integer division.
func (b *BCD) DivInt(other *BCD) (*BCD, error) {
	if other.IsZero() {
		return nil, ErrDivisionByZero
	}

	if b.IsZero() {
		return Zero(), nil
	}

	quotient, _ := divideWithRemainder(b, other)
	quotient.negative = b.negative != other.negative

	// Truncate to integer
	if quotient.scale > 0 {
		quotient.digits = quotient.digits[quotient.scale:]
		quotient.scale = 0
	}

	if len(quotient.digits) == 0 {
		return Zero(), nil
	}

	return quotient, nil
}

// Mod returns the remainder of b divided by other.
func (b *BCD) Mod(other *BCD) (*BCD, error) {
	if other.IsZero() {
		return nil, ErrDivisionByZero
	}

	if b.IsZero() {
		return Zero(), nil
	}

	_, remainder := divideWithRemainder(b, other)
	remainder.negative = b.negative
	return remainder, nil
}

// Round rounds the BCD to the specified scale using the given rounding mode.
func (b *BCD) Round(scale int, mode RoundingMode) *BCD {
	if scale < 0 {
		scale = 0
	}

	if b.scale <= scale {
		return b.Copy()
	}

	result := b.Copy()
	digitsToRemove := result.scale - scale

	if digitsToRemove >= len(result.digits) {
		return Zero()
	}

	// Get rounding digit
	roundingDigit := result.digits[digitsToRemove-1]

	// Remove excess digits
	result.digits = result.digits[digitsToRemove:]
	result.scale = scale

	// Apply rounding
	shouldRoundUp := false
	switch mode {
	case RoundUp:
		shouldRoundUp = roundingDigit > 0
	case RoundDown:
		shouldRoundUp = false
	case RoundHalfUp:
		shouldRoundUp = roundingDigit >= 5
	case RoundHalfDown:
		shouldRoundUp = roundingDigit > 5
	case RoundHalfEven:
		if roundingDigit > 5 {
			shouldRoundUp = true
		} else if roundingDigit == 5 {
			// Round to even
			if len(result.digits) > 0 {
				shouldRoundUp = result.digits[0]%2 == 1
			}
		}
	case RoundCeiling:
		shouldRoundUp = !result.negative && roundingDigit > 0
	case RoundFloor:
		shouldRoundUp = result.negative && roundingDigit > 0
	}

	if shouldRoundUp {
		addOne(result.digits)
	}

	// Handle zero
	if isZero(result.digits) {
		result.negative = false
	}

	return result
}

// ToInt64 converts the BCD to int64, truncating any decimal places.
func (b *BCD) ToInt64() (int64, error) {
	// Start from the decimal point
	startIdx := b.scale
	if startIdx >= len(b.digits) {
		return 0, nil
	}

	var result int64
	multiplier := int64(1)

	for i := startIdx; i < len(b.digits); i++ {
		digit := int64(b.digits[i])
		if result > (math.MaxInt64-digit)/10 {
			return 0, ErrOverflow
		}
		result = result + digit*multiplier
		multiplier *= 10
	}

	if b.negative {
		result = -result
	}

	return result, nil
}

// ToFloat64 converts the BCD to float64 (may lose precision).
func (b *BCD) ToFloat64() float64 {
	f, _ := strconv.ParseFloat(b.String(), 64)
	return f
}

// Internal helper functions

func isZero(digits []uint8) bool {
	for _, d := range digits {
		if d != 0 {
			return false
		}
	}
	return true
}

func compareMagnitudes(a, b *BCD) int {
	// Align decimal places
	aDigits, bDigits, _ := alignDecimals(a, b)

	// Compare lengths
	if len(aDigits) != len(bDigits) {
		if len(aDigits) > len(bDigits) {
			return 1
		}
		return -1
	}

	// Compare digit by digit (most significant first)
	for i := len(aDigits) - 1; i >= 0; i-- {
		if aDigits[i] > bDigits[i] {
			return 1
		} else if aDigits[i] < bDigits[i] {
			return -1
		}
	}

	return 0
}

func alignDecimals(a, b *BCD) ([]uint8, []uint8, int) {
	maxScale := max(b.scale, a.scale)

	aDigits := make([]uint8, len(a.digits))
	copy(aDigits, a.digits)
	bDigits := make([]uint8, len(b.digits))
	copy(bDigits, b.digits)

	// Pad with zeros to align decimal places
	if a.scale < maxScale {
		padding := make([]uint8, maxScale-a.scale)
		aDigits = append(padding, aDigits...)
	}
	if b.scale < maxScale {
		padding := make([]uint8, maxScale-b.scale)
		bDigits = append(padding, bDigits...)
	}

	// Ensure same length
	maxLen := max(len(bDigits), len(aDigits))

	if len(aDigits) < maxLen {
		aDigits = append(aDigits, make([]uint8, maxLen-len(aDigits))...)
	}
	if len(bDigits) < maxLen {
		bDigits = append(bDigits, make([]uint8, maxLen-len(bDigits))...)
	}

	return aDigits, bDigits, maxScale
}

func addMagnitudes(a, b *BCD) *BCD {
	aDigits, bDigits, scale := alignDecimals(a, b)

	result := make([]uint8, len(aDigits)+1)
	carry := uint8(0)

	for i := range aDigits {
		sum := aDigits[i] + bDigits[i] + carry
		result[i] = sum % 10
		carry = sum / 10
	}

	if carry > 0 {
		result[len(aDigits)] = carry
	} else {
		result = result[:len(aDigits)]
	}

	// Remove leading zeros
	for len(result) > 1 && result[len(result)-1] == 0 {
		result = result[:len(result)-1]
	}

	return &BCD{
		digits:   result,
		scale:    scale,
		negative: false,
	}
}

func subtractMagnitudes(a, b *BCD) *BCD {
	aDigits, bDigits, scale := alignDecimals(a, b)

	result := make([]uint8, len(aDigits))
	borrow := uint8(0)

	for i := range aDigits {
		diff := int8(aDigits[i]) - int8(bDigits[i]) - int8(borrow)
		if diff < 0 {
			diff += 10
			borrow = 1
		} else {
			borrow = 0
		}
		result[i] = uint8(diff)
	}

	// Remove leading zeros
	for len(result) > 1 && result[len(result)-1] == 0 {
		result = result[:len(result)-1]
	}

	return &BCD{
		digits:   result,
		scale:    scale,
		negative: false,
	}
}

func addOne(digits []uint8) {
	carry := uint8(1)
	for i := 0; i < len(digits) && carry > 0; i++ {
		sum := digits[i] + carry
		digits[i] = sum % 10
		carry = sum / 10
	}
}

func divideWithRemainder(dividend, divisor *BCD) (*BCD, *BCD) {
	// Simple long division implementation
	// This is a basic implementation - production code might use more efficient algorithms

	// Normalize operands to integers
	dividendInt := dividend.Copy()
	divisorInt := divisor.Copy()

	// Scale to remove decimals
	scaleAdjust := 0
	if dividendInt.scale > divisorInt.scale {
		scaleAdjust = dividendInt.scale - divisorInt.scale
		padding := make([]uint8, scaleAdjust)
		divisorInt.digits = append(padding, divisorInt.digits...)
	} else if divisorInt.scale > dividendInt.scale {
		scaleAdjust = divisorInt.scale - dividendInt.scale
		padding := make([]uint8, scaleAdjust)
		dividendInt.digits = append(padding, dividendInt.digits...)
	}

	dividendInt.scale = 0
	divisorInt.scale = 0

	// Perform integer division
	if compareMagnitudes(dividendInt, divisorInt) < 0 {
		return Zero(), dividendInt
	}

	quotientDigits := make([]uint8, 0)
	remainder := Zero()

	// Process dividend digits from most significant to least
	for i := len(dividendInt.digits) - 1; i >= 0; i-- {
		// Shift remainder left and add next digit
		remainder.digits = append([]uint8{dividendInt.digits[i]}, remainder.digits...)
		
		// Find how many times divisor fits into current remainder
		count := uint8(0)
		for compareMagnitudes(remainder, divisorInt) >= 0 {
			remainder = subtractMagnitudes(remainder, divisorInt)
			count++
		}

		quotientDigits = append([]uint8{count}, quotientDigits...)
	}

	// Remove leading zeros from quotient
	for len(quotientDigits) > 1 && quotientDigits[len(quotientDigits)-1] == 0 {
		quotientDigits = quotientDigits[:len(quotientDigits)-1]
	}

	quotient := &BCD{
		digits:   quotientDigits,
		scale:    0,
		negative: false,
	}

	return quotient, remainder
}