// Tideland Go BCD
//
// Copyright (C) 2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package bcd provides a Binary Coded Decimal (BCD) implementation
// for precise decimal arithmetic, particularly useful for financial
// and currency calculations where floating-point errors are unacceptable.
package bcd

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Common errors returned by the package.
var (
	ErrDivisionByZero = fmt.Errorf("division by zero")
	ErrInvalidFormat  = fmt.Errorf("invalid decimal format")
	ErrOverflow       = fmt.Errorf("numeric overflow")
	ErrPrecisionLoss  = fmt.Errorf("precision loss")
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

// BCD represents a decimal number using binary coded decimal encoding.
type BCD struct {
	// digits stores the decimal digits in little-endian order.
	// For example, 123.45 is stored as [5, 4, 3, 2, 1].
	digits []uint8
	// scale is the number of digits after the decimal point.
	// For 123.45, scale is 2.
	scale int
	// negative indicates if the number is negative.
	negative bool
}

// Numeric represents types that can be converted to BCD.
type Numeric interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Option represents optional parameters for BCD creation.
type Option func(*options)

type options struct {
	scale        int
	roundingMode RoundingMode
}

// WithScale sets the scale for float conversions.
func WithScale(scale int) Option {
	return func(o *options) {
		o.scale = scale
	}
}

// WithRounding sets the rounding mode.
func WithRounding(mode RoundingMode) Option {
	return func(o *options) {
		o.roundingMode = mode
	}
}

// New creates a BCD from any numeric type.
func New[T Numeric](value T, opts ...Option) (*BCD, error) {
	options := &options{
		scale:        6, // default scale for floats
		roundingMode: RoundHalfEven,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Use type switch on the underlying value
	var v any = value
	switch val := v.(type) {
	case string:
		return parseString(val)
	case int:
		return fromInt64(int64(val)), nil
	case int8:
		return fromInt64(int64(val)), nil
	case int16:
		return fromInt64(int64(val)), nil
	case int32:
		return fromInt64(int64(val)), nil
	case int64:
		return fromInt64(val), nil
	case uint:
		if val > math.MaxInt64 {
			return nil, ErrOverflow
		}
		return fromInt64(int64(val)), nil
	case uint8:
		return fromInt64(int64(val)), nil
	case uint16:
		return fromInt64(int64(val)), nil
	case uint32:
		return fromInt64(int64(val)), nil
	case uint64:
		if val > math.MaxInt64 {
			return nil, ErrOverflow
		}
		return fromInt64(int64(val)), nil
	case float32:
		return fromFloat64(float64(val), options.scale)
	case float64:
		return fromFloat64(val, options.scale)
	default:
		return nil, fmt.Errorf("%w: unsupported type %T", ErrInvalidFormat, value)
	}
}

// Must creates a BCD and panics on error.
// Useful for package-level variables and constants.
func Must[T Numeric](value T, opts ...Option) *BCD {
	bcd, err := New(value, opts...)
	if err != nil {
		panic(fmt.Sprintf("bcd.Must: %v", err))
	}
	return bcd
}

// Zero returns a BCD representing zero.
func Zero() *BCD {
	return &BCD{digits: []uint8{0}, scale: 0, negative: false}
}

// parseString parses a decimal string into a BCD.
func parseString(s string) (*BCD, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Zero(), nil
	}

	// Handle scientific notation
	if strings.ContainsAny(s, "eE") {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidFormat, err)
		}

		// Determine appropriate scale
		// For very small numbers, we need to ensure enough decimal places
		absF := math.Abs(f)
		scale := 6 // default

		if absF < 1 && absF > 0 {
			// For small numbers, determine scale based on the first significant digit
			scale = 0
			temp := absF
			for temp < 1 {
				temp *= 10
				scale++
			}
			// Add a few more digits for precision
			scale += 6
		}

		// Also check if the original string had decimal places specified
		if dotIdx := strings.Index(s, "."); dotIdx >= 0 {
			eIdx := strings.IndexAny(s, "eE")
			if eIdx < 0 {
				eIdx = len(s)
			}
			specifiedScale := eIdx - dotIdx - 1
			if specifiedScale > scale {
				scale = specifiedScale
			}
		}

		return fromFloat64(f, scale)
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
	if allDigits == "" || allDigits == "0" {
		return Zero(), nil
	}

	digits := make([]uint8, len(allDigits))
	for i := range len(allDigits) {
		digits[i] = uint8(allDigits[len(allDigits)-1-i] - '0')
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

// fromInt64 creates a BCD from an int64.
func fromInt64(n int64) *BCD {
	if n == 0 {
		return Zero()
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

// fromFloat64 creates a BCD from a float64 with specified scale.
func fromFloat64(f float64, scale int) (*BCD, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, ErrInvalidFormat
	}

	// Convert to string with proper precision
	format := fmt.Sprintf("%%.%df", scale)
	s := fmt.Sprintf(format, f)
	return parseString(s)
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
	if len(b.digits) == 0 || (len(b.digits) == 1 && b.digits[0] == 0) {
		return "0"
	}

	// Convert digits to string (remember they're in little-endian)
	var sb strings.Builder

	// Pre-allocate capacity
	capacity := len(b.digits) + 1 // digits + possible decimal point
	if b.negative {
		capacity++
	}
	sb.Grow(capacity)

	if b.negative {
		sb.WriteByte('-')
	}

	// Determine where to place the decimal point
	intDigits := len(b.digits) - b.scale

	if intDigits <= 0 {
		// Number is less than 1
		sb.WriteString("0.")
		// Add leading zeros
		for range -intDigits {
			sb.WriteByte('0')
		}
		// Add the digits
		for i := len(b.digits) - 1; i >= 0; i-- {
			sb.WriteByte(b.digits[i] + '0')
		}
	} else {
		// Add integer part
		for i := len(b.digits) - 1; i >= b.scale; i-- {
			sb.WriteByte(b.digits[i] + '0')
		}

		if b.scale > 0 {
			// Add decimal point and fractional part
			sb.WriteByte('.')
			for i := b.scale - 1; i >= 0; i-- {
				sb.WriteByte(b.digits[i] + '0')
			}
		}
	}

	return sb.String()
}

// IsZero returns true if the BCD is zero.
func (b *BCD) IsZero() bool {
	return isZero(b.digits)
}

// IsNegative returns true if the BCD is negative.
func (b *BCD) IsNegative() bool {
	return b.negative && !b.IsZero()
}

// IsPositive returns true if the BCD is positive.
func (b *BCD) IsPositive() bool {
	return !b.negative && !b.IsZero()
}

// Scale returns the scale (number of decimal places) of the BCD.
func (b *BCD) Scale() int {
	return b.scale
}

// Precision returns the total number of significant digits.
func (b *BCD) Precision() int {
	// Remove trailing zeros from fractional part for precision calculation
	digits := b.digits
	scale := b.scale

	for len(digits) > 0 && scale > 0 && digits[0] == 0 {
		digits = digits[1:]
		scale--
	}

	return len(digits)
}

// Abs returns the absolute value of the BCD.
func (b *BCD) Abs() *BCD {
	c := b.Copy()
	c.negative = false
	return c
}

// Neg returns the negation of the BCD.
func (b *BCD) Neg() *BCD {
	if b.IsZero() {
		return b.Copy()
	}
	c := b.Copy()
	c.negative = !c.negative
	return c
}

// Normalize removes trailing zeros after the decimal point.
func (b *BCD) Normalize() *BCD {
	if b.scale == 0 || b.IsZero() {
		return b.Copy()
	}

	c := b.Copy()

	// Count trailing zeros
	trailingZeros := 0
	for i := 0; i < c.scale && i < len(c.digits) && c.digits[i] == 0; i++ {
		trailingZeros++
	}

	if trailingZeros > 0 {
		// Remove trailing zeros
		c.digits = c.digits[trailingZeros:]
		c.scale -= trailingZeros

		// If all fractional digits were zeros, set scale to 0
		if c.scale == 0 || len(c.digits) == 0 {
			c.scale = 0
			if len(c.digits) == 0 {
				c.digits = []uint8{0}
			}
		}
	}

	return c
}

// Cmp compares two BCDs and returns:
//
//	-1 if b < other
//	 0 if b == other
//	 1 if b > other
func (b *BCD) Cmp(other *BCD) int {
	// Handle signs
	if b.negative && !other.negative {
		return -1
	}
	if !b.negative && other.negative {
		return 1
	}

	// Same sign, compare magnitudes
	result := compareMagnitudes(b, other)

	// If both negative, reverse the result
	if b.negative {
		return -result
	}
	return result
}

// Equal returns true if b equals other.
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

// Add returns b + other.
func (b *BCD) Add(other *BCD) *BCD {
	// Handle same sign
	if b.negative == other.negative {
		aligned1, aligned2 := alignDecimals(b, other)
		sum := addMagnitudes(aligned1, aligned2)
		sum.negative = b.negative
		return sum
	}

	// Different signs: this is subtraction
	cmp := compareMagnitudes(b, other)
	if cmp == 0 {
		return Zero()
	}

	aligned1, aligned2 := alignDecimals(b, other)
	if cmp > 0 {
		// |b| > |other|
		diff := subtractMagnitudes(aligned1, aligned2)
		diff.negative = b.negative
		return diff
	} else {
		// |b| < |other|
		diff := subtractMagnitudes(aligned2, aligned1)
		diff.negative = other.negative
		return diff
	}
}

// Sub returns b - other.
func (b *BCD) Sub(other *BCD) *BCD {
	return b.Add(other.Neg())
}

// Mul returns b * other.
func (b *BCD) Mul(other *BCD) *BCD {
	if b.IsZero() || other.IsZero() {
		return Zero()
	}

	// Multiply digits
	resultDigits := make([]uint8, len(b.digits)+len(other.digits))

	for i := range b.digits {
		carry := uint8(0)
		for j := range other.digits {
			prod := b.digits[i]*other.digits[j] + resultDigits[i+j] + carry
			resultDigits[i+j] = prod % 10
			carry = prod / 10
		}
		if carry > 0 {
			resultDigits[i+len(other.digits)] = carry
		}
	}

	// Remove leading zeros
	for len(resultDigits) > 1 && resultDigits[len(resultDigits)-1] == 0 {
		resultDigits = resultDigits[:len(resultDigits)-1]
	}

	return &BCD{
		digits:   resultDigits,
		scale:    b.scale + other.scale,
		negative: b.negative != other.negative,
	}
}

// Div returns b / other with the specified scale and rounding mode.
func (b *BCD) Div(other *BCD, scale int, mode RoundingMode) (*BCD, error) {
	if other.IsZero() {
		return nil, ErrDivisionByZero
	}

	if b.IsZero() {
		return Zero(), nil
	}

	// Perform division with extra precision for rounding
	quotient, _ := divideWithRemainder(b, other, scale+1)

	// Apply rounding
	quotient.negative = b.negative != other.negative
	quotient = quotient.Round(scale, mode)

	return quotient, nil
}

// DivInt returns the integer quotient b / other.
func (b *BCD) DivInt(other *BCD) (*BCD, error) {
	if other.IsZero() {
		return nil, ErrDivisionByZero
	}

	if b.IsZero() {
		return Zero(), nil
	}

	quotient, _ := divideWithRemainder(b, other, 0)
	quotient.negative = b.negative != other.negative

	// Truncate to integer
	if quotient.scale > 0 {
		// Remove fractional digits
		if quotient.scale >= len(quotient.digits) {
			return Zero(), nil
		}
		quotient.digits = quotient.digits[quotient.scale:]
		quotient.scale = 0
	}

	return quotient, nil
}

// Mod returns b % other.
func (b *BCD) Mod(other *BCD) (*BCD, error) {
	if other.IsZero() {
		return nil, ErrDivisionByZero
	}

	if b.IsZero() {
		return Zero(), nil
	}

	_, remainder := divideWithRemainder(b, other, 0)
	remainder.negative = b.negative

	return remainder, nil
}

// Round rounds the BCD to the specified number of decimal places using the given mode.
func (b *BCD) Round(places int, mode RoundingMode) *BCD {
	if places < 0 {
		places = 0
	}

	// Already at desired precision
	if b.scale <= places {
		return b.Copy()
	}

	// Calculate how many digits to remove
	removeCount := b.scale - places
	if removeCount >= len(b.digits) {
		// All digits would be removed
		if b.IsZero() {
			return Zero()
		}
		// Round 0.00...xyz to 0 or ±1
		if shouldRoundUp(b.digits[len(b.digits)-1], 0, false, mode, b.negative) {
			result := &BCD{
				digits:   []uint8{1},
				scale:    places,
				negative: b.negative,
			}
			if places == 0 {
				result.scale = 0
			} else {
				// Need to add zeros
				zeros := make([]uint8, places-1)
				result.digits = append(zeros, 1)
			}
			return result
		}
		return Zero()
	}

	// Copy digits we're keeping
	newDigits := make([]uint8, len(b.digits)-removeCount)
	copy(newDigits, b.digits[removeCount:])

	// Check if we need to round up
	roundDigit := b.digits[removeCount-1]
	var nextDigit uint8
	if removeCount >= 2 {
		nextDigit = b.digits[removeCount-2]
	}
	isEven := newDigits[0]%2 == 0

	if shouldRoundUp(roundDigit, nextDigit, isEven, mode, b.negative) {
		// Add 1 to the result
		carry := uint8(1)
		for i := 0; i < len(newDigits) && carry > 0; i++ {
			sum := newDigits[i] + carry
			newDigits[i] = sum % 10
			carry = sum / 10
		}
		if carry > 0 {
			newDigits = append(newDigits, carry)
		}
	}

	// Remove trailing zeros if scale becomes 0
	if places == 0 {
		for len(newDigits) > 1 && newDigits[0] == 0 {
			newDigits = newDigits[1:]
		}
	}

	return &BCD{
		digits:   newDigits,
		scale:    places,
		negative: b.negative,
	}
}

// ToInt64 converts the BCD to int64, returning an error if the value doesn't fit.
func (b *BCD) ToInt64() (int64, error) {
	// Truncate to integer (round towards zero)
	rounded := b.Round(0, RoundDown)

	// Check if it's zero
	if rounded.IsZero() {
		return 0, nil
	}

	// Convert digits to int64
	var result int64
	multiplier := int64(1)

	for _, digit := range rounded.digits {
		digitValue := int64(digit) * multiplier

		// Check for overflow
		if multiplier > math.MaxInt64/10 {
			return 0, ErrOverflow
		}

		newResult := result + digitValue
		if (result > 0 && newResult < result) || (result < 0 && newResult > result) {
			return 0, ErrOverflow
		}

		result = newResult
		multiplier *= 10
	}

	if b.negative {
		result = -result
		if result > 0 {
			return 0, ErrOverflow
		}
	}

	return result, nil
}

// ToFloat64 converts the BCD to float64.
func (b *BCD) ToFloat64() float64 {
	f, _ := strconv.ParseFloat(b.String(), 64)
	return f
}

// Helper functions

// isZero checks if all digits are zero.
func isZero(digits []uint8) bool {
	for _, d := range digits {
		if d != 0 {
			return false
		}
	}
	return true
}

// compareMagnitudes compares the absolute values of two BCDs.
func compareMagnitudes(a, b *BCD) int {
	// Align decimals for comparison
	aligned1, aligned2 := alignDecimals(a, b)

	// Compare lengths
	if len(aligned1.digits) > len(aligned2.digits) {
		return 1
	}
	if len(aligned1.digits) < len(aligned2.digits) {
		return -1
	}

	// Same length, compare digits from most significant
	for i := len(aligned1.digits) - 1; i >= 0; i-- {
		if aligned1.digits[i] > aligned2.digits[i] {
			return 1
		}
		if aligned1.digits[i] < aligned2.digits[i] {
			return -1
		}
	}

	return 0
}

// alignDecimals aligns two BCDs to have the same scale.
func alignDecimals(a, b *BCD) (*BCD, *BCD) {
	if a.scale == b.scale {
		return a.Copy(), b.Copy()
	}

	aCopy := a.Copy()
	bCopy := b.Copy()

	if a.scale < b.scale {
		// Add zeros to a
		diff := b.scale - a.scale
		zeros := make([]uint8, diff)
		aCopy.digits = append(zeros, aCopy.digits...)
		aCopy.scale = b.scale
	} else {
		// Add zeros to b
		diff := a.scale - b.scale
		zeros := make([]uint8, diff)
		bCopy.digits = append(zeros, bCopy.digits...)
		bCopy.scale = a.scale
	}

	return aCopy, bCopy
}

// addMagnitudes adds two positive BCDs with the same scale.
func addMagnitudes(a, b *BCD) *BCD {
	maxLen := max(len(b.digits), len(a.digits))

	result := make([]uint8, maxLen+1)
	carry := uint8(0)

	for i := 0; i < maxLen || carry > 0; i++ {
		sum := carry
		if i < len(a.digits) {
			sum += a.digits[i]
		}
		if i < len(b.digits) {
			sum += b.digits[i]
		}

		result[i] = sum % 10
		carry = sum / 10
	}

	// Remove leading zeros
	for len(result) > 1 && result[len(result)-1] == 0 {
		result = result[:len(result)-1]
	}

	return &BCD{
		digits: result,
		scale:  a.scale,
	}
}

// subtractMagnitudes subtracts b from a (assumes a >= b).
func subtractMagnitudes(a, b *BCD) *BCD {
	result := make([]uint8, len(a.digits))
	borrow := uint8(0)

	for i := range a.digits {
		diff := int8(a.digits[i]) - int8(borrow)
		if i < len(b.digits) {
			diff -= int8(b.digits[i])
		}

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
		digits: result,
		scale:  a.scale,
	}
}

// shouldRoundUp determines if rounding should increase the magnitude.
func shouldRoundUp(digit, nextDigit uint8, isEven bool, mode RoundingMode, negative bool) bool {
	switch mode {
	case RoundDown:
		return false
	case RoundUp:
		return digit > 0 || nextDigit > 0
	case RoundHalfUp:
		return digit >= 5
	case RoundHalfDown:
		return digit > 5 || (digit == 5 && nextDigit > 0)
	case RoundHalfEven:
		if digit > 5 || (digit == 5 && nextDigit > 0) {
			return true
		}
		if digit == 5 && nextDigit == 0 {
			return !isEven
		}
		return false
	case RoundCeiling:
		if negative {
			return false
		}
		return digit > 0 || nextDigit > 0
	case RoundFloor:
		if !negative {
			return false
		}
		return digit > 0 || nextDigit > 0
	default:
		return false
	}
}

// divideWithRemainder performs long division of a by b.
func divideWithRemainder(a, b *BCD, targetScale int) (*BCD, *BCD) {
	if b.IsZero() {
		panic("division by zero")
	}

	// Convert to integers by removing decimal points
	// a has 'a.scale' decimal places, b has 'b.scale' decimal places
	// We want the result to have 'targetScale' decimal places

	// To get targetScale decimal places in the result, we need:
	// (a * 10^(targetScale + b.scale)) / b
	// This gives us targetScale decimal places after accounting for b's scale

	dividend := a.Copy()
	divisor := b.Copy()

	// Add extra zeros to dividend to get desired precision
	extraZeros := targetScale + b.scale
	if extraZeros > 0 {
		zeros := make([]uint8, extraZeros)
		dividend.digits = append(zeros, dividend.digits...)
	}

	// Now both are treated as integers
	dividend.scale = 0
	divisor.scale = 0

	// Perform integer division
	quotient, remainder := divideIntegers(dividend, divisor)

	// Set the correct scale
	quotient.scale = targetScale + a.scale

	return quotient, remainder
}

// divideIntegers performs integer division of two BCDs treated as integers
func divideIntegers(a, b *BCD) (*BCD, *BCD) {
	if len(b.digits) == 1 && b.digits[0] == 0 {
		panic("division by zero")
	}

	// Handle simple case where divisor is single digit
	if len(b.digits) == 1 {
		return divideBySmallInt(a, b.digits[0])
	}

	// For larger divisors, use long division
	return longDivision(a, b)
}

// divideBySmallInt divides by a single digit
func divideBySmallInt(a *BCD, divisor uint8) (*BCD, *BCD) {
	if divisor == 0 {
		panic("division by zero")
	}

	quotientDigits := make([]uint8, len(a.digits))
	remainder := uint16(0)

	// Process from most significant digit
	for i := len(a.digits) - 1; i >= 0; i-- {
		dividend := remainder*10 + uint16(a.digits[i])
		quotientDigits[i] = uint8(dividend / uint16(divisor))
		remainder = dividend % uint16(divisor)
	}

	// Remove leading zeros from the most significant end (the end of the array)
	// But keep at least one digit
	for len(quotientDigits) > 1 && quotientDigits[len(quotientDigits)-1] == 0 {
		quotientDigits = quotientDigits[:len(quotientDigits)-1]
	}

	return &BCD{digits: quotientDigits}, &BCD{digits: []uint8{uint8(remainder)}}
}

// longDivision performs long division for multi-digit divisors
func longDivision(dividend, divisor *BCD) (*BCD, *BCD) {
	if compareMagnitudes(dividend, divisor) < 0 {
		return Zero(), dividend.Copy()
	}

	// Convert divisor to uint64 if possible for faster division
	if len(divisor.digits) <= 15 { // Fits in uint64
		divisorValue := uint64(0)
		multiplier := uint64(1)
		for i := range divisor.digits {
			divisorValue += uint64(divisor.digits[i]) * multiplier
			multiplier *= 10
		}

		// Convert dividend to uint64 if possible
		if len(dividend.digits) <= 15 {
			dividendValue := uint64(0)
			multiplier = 1
			for i := range dividend.digits {
				dividendValue += uint64(dividend.digits[i]) * multiplier
				multiplier *= 10
			}

			quotientValue := dividendValue / divisorValue
			remainderValue := dividendValue % divisorValue

			// Convert back to BCD
			quotient := &BCD{digits: []uint8{0}}
			if quotientValue > 0 {
				var qdigits []uint8
				for quotientValue > 0 {
					qdigits = append(qdigits, uint8(quotientValue%10))
					quotientValue /= 10
				}
				quotient.digits = qdigits
			}

			remainder := &BCD{digits: []uint8{0}}
			if remainderValue > 0 {
				var rdigits []uint8
				for remainderValue > 0 {
					rdigits = append(rdigits, uint8(remainderValue%10))
					remainderValue /= 10
				}
				remainder.digits = rdigits
			}

			return quotient, remainder
		}
	}

	// Fallback to digit-by-digit division for large numbers
	// This is still slow but at least correct
	quotientDigits := make([]uint8, 0)
	remainder := dividend.Copy()

	// Estimate quotient by repeated subtraction with multipliers
	for compareMagnitudes(remainder, divisor) >= 0 {
		// Find the largest multiplier where divisor * multiplier <= remainder
		multiplier := uint8(1)
		testProduct := divisor.Copy()

		for multiplier < 9 {
			nextProduct := multiplyByDigit(divisor, multiplier+1)
			if compareMagnitudes(nextProduct, remainder) > 0 {
				break
			}
			multiplier++
			testProduct = nextProduct
		}

		// Subtract divisor * multiplier from remainder
		remainder = subtractMagnitudes(remainder, testProduct)
		quotientDigits = append([]uint8{multiplier}, quotientDigits...)
	}

	// Reverse quotient digits
	for i, j := 0, len(quotientDigits)-1; i < j; i, j = i+1, j-1 {
		quotientDigits[i], quotientDigits[j] = quotientDigits[j], quotientDigits[i]
	}

	if len(quotientDigits) == 0 {
		quotientDigits = []uint8{0}
	}

	return &BCD{digits: quotientDigits}, remainder
}

// multiplyByDigit multiplies a BCD by a single digit
func multiplyByDigit(b *BCD, digit uint8) *BCD {
	if digit == 0 {
		return Zero()
	}

	result := make([]uint8, len(b.digits)+1)
	carry := uint8(0)

	for i := range b.digits {
		prod := b.digits[i]*digit + carry
		result[i] = prod % 10
		carry = prod / 10
	}

	if carry > 0 {
		result[len(b.digits)] = carry
	} else {
		result = result[:len(b.digits)]
	}

	return &BCD{
		digits: result,
		scale:  b.scale,
	}
}
