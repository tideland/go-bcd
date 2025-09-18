// Tideland Go BCD
//
// Copyright (C) 2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package bcd

import (
	"math"
	"testing"

	"tideland.dev/go/asserts/verify"
)

func TestNewBCD_String(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"zero", "0", "0", false},
		{"positive integer", "123", "123", false},
		{"negative integer", "-456", "-456", false},
		{"positive decimal", "123.45", "123.45", false},
		{"negative decimal", "-123.45", "-123.45", false},
		{"leading zeros", "000123.45", "123.45", false},
		{"trailing zeros", "123.4500", "123.45", false},
		{"decimal only", ".5", "0.5", false},
		{"negative decimal only", "-.5", "-0.5", false},
		{"large number", "999999999999999999.99", "999999999999999999.99", false},
		{"very small", "0.000001", "0.000001", false},
		{"empty string", "", "0", false},
		{"scientific notation", "1.23e-4", "0.000123", false},
		{"invalid format", "12.34.56", "", true},
		{"invalid chars", "12a34", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.input)
			if tt.wantErr {
				verify.ErrorMatch(t, err, ".*")
			} else {
				verify.NoError(t, err)
				verify.Equal(t, got.String(), tt.want)
			}
		})
	}
}

func TestNewBCD_Generic(t *testing.T) {
	// Test generic API with different types
	t.Run("int types", func(t *testing.T) {
		tests := []struct {
			name  string
			value any
			want  string
		}{
			{"int", int(123), "123"},
			{"int8", int8(-128), "-128"},
			{"int16", int16(32767), "32767"},
			{"int32", int32(-2147483648), "-2147483648"},
			{"int64", int64(9223372036854775807), "9223372036854775807"},
			{"uint", uint(123), "123"},
			{"uint8", uint8(255), "255"},
			{"uint16", uint16(65535), "65535"},
			{"uint32", uint32(4294967295), "4294967295"},
			{"uint64", uint64(9223372036854775807), "9223372036854775807"}, // max int64
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var got *BCD

				switch v := tt.value.(type) {
				case int:
					got, _ = New(v)
				case int8:
					got, _ = New(v)
				case int16:
					got, _ = New(v)
				case int32:
					got, _ = New(v)
				case int64:
					got, _ = New(v)
				case uint:
					got, _ = New(v)
				case uint8:
					got, _ = New(v)
				case uint16:
					got, _ = New(v)
				case uint32:
					got, _ = New(v)
				case uint64:
					got, _ = New(v)
				}

				verify.NotNil(t, got)
				verify.Equal(t, got.String(), tt.want)
			})
		}
	})

	t.Run("float types", func(t *testing.T) {
		tests := []struct {
			name  string
			value any
			opts  []Option
			want  string
		}{
			{"float32", float32(123.45), nil, "123.449997"}, // float32 precision
			{"float64", float64(123.45), nil, "123.45"},
			{"float with scale", 123.456789, []Option{WithScale(3)}, "123.457"},
			{"float with rounding", 1.2345, []Option{WithScale(2), WithRounding(RoundDown)}, "1.23"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var got *BCD
				var err error

				switch v := tt.value.(type) {
				case float32:
					got, err = New(v, tt.opts...)
				case float64:
					got, err = New(v, tt.opts...)
				}

				verify.NoError(t, err)
				verify.Equal(t, got.String(), tt.want)
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		_, err := New(math.NaN())
		verify.ErrorMatch(t, err, ".*")

		_, err = New(math.Inf(1))
		verify.ErrorMatch(t, err, ".*")
	})
}

func TestMust(t *testing.T) {
	// Test that Must works correctly
	bcd := Must("123.45")
	verify.Equal(t, bcd.String(), "123.45")

	// Test that Must panics on error
	defer func() {
		r := recover()
		verify.NotNil(t, r)
	}()
	Must("invalid")
}

func TestBCDArithmetic(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		op   string
		want string
	}{
		// Addition
		{"add positive integers", "123", "456", "+", "579"},
		{"add positive and negative", "100", "-30", "+", "70"},
		{"add decimals", "12.34", "56.78", "+", "69.12"},
		{"add with carry", "999", "1", "+", "1000"},
		{"add zero", "123.45", "0", "+", "123.45"},

		// Subtraction
		{"subtract positive", "100", "30", "-", "70"},
		{"subtract larger", "30", "100", "-", "-70"},
		{"subtract decimals", "12.34", "5.67", "-", "6.67"},
		{"subtract same", "123.45", "123.45", "-", "0"},

		// Multiplication
		{"multiply integers", "12", "34", "*", "408"},
		{"multiply decimals", "1.2", "3.4", "*", "4.08"},
		{"multiply by zero", "123.45", "0", "*", "0"},
		{"multiply negative", "-5", "6", "*", "-30"},
		{"multiply two negatives", "-5", "-6", "*", "30"},

		// Division
		{"divide integers", "100", "5", "/", "20"},
		{"divide with decimal", "10", "4", "/", "2.5"},
		{"divide decimals", "7.5", "2.5", "/", "3"},
		{"divide by one", "123.45", "1", "/", "123.45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := New(tt.a)
			verify.NoError(t, err)

			b, err := New(tt.b)
			verify.NoError(t, err)

			var result *BCD
			switch tt.op {
			case "+":
				result = a.Add(b)
			case "-":
				result = a.Sub(b)
			case "*":
				result = a.Mul(b)
			case "/":
				result, err = a.Div(b, 10, RoundHalfUp)
				verify.NoError(t, err)
				// Simplify result for comparison
				result = result.Round(2, RoundHalfUp).Normalize()
			}

			verify.Equal(t, result.String(), tt.want)
		})
	}
}

func TestBCDComparison(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		cmp  int // -1: a < b, 0: a == b, 1: a > b
	}{
		{"equal integers", "123", "123", 0},
		{"equal decimals", "123.45", "123.45", 0},
		{"less than", "123", "456", -1},
		{"greater than", "456", "123", 1},
		{"negative less than positive", "-5", "5", -1},
		{"negative less than zero", "-5", "0", -1},
		{"positive greater than zero", "5", "0", 1},
		{"decimal comparison", "123.45", "123.46", -1},
		{"different scales equal", "123.4500", "123.45", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := New(tt.a)
			b, _ := New(tt.b)

			got := a.Cmp(b)
			verify.Equal(t, got, tt.cmp)

			// Test comparison methods
			if tt.cmp < 0 {
				verify.True(t, a.LessThan(b))
			}
			if tt.cmp == 0 {
				verify.True(t, a.Equal(b))
			}
			if tt.cmp > 0 {
				verify.True(t, a.GreaterThan(b))
			}
		})
	}
}

func TestBCDRounding(t *testing.T) {
	tests := []struct {
		name  string
		value string
		scale int
		mode  RoundingMode
		want  string
	}{
		// RoundHalfUp
		{"half up 1.25", "1.25", 1, RoundHalfUp, "1.3"},
		{"half up 1.24", "1.24", 1, RoundHalfUp, "1.2"},
		{"half up 1.26", "1.26", 1, RoundHalfUp, "1.3"},
		{"half up negative", "-1.25", 1, RoundHalfUp, "-1.3"},

		// RoundHalfDown
		{"half down 1.25", "1.25", 1, RoundHalfDown, "1.2"},
		{"half down 1.26", "1.26", 1, RoundHalfDown, "1.3"},

		// RoundHalfEven (Banker's rounding)
		{"half even 1.25", "1.25", 1, RoundHalfEven, "1.2"},
		{"half even 1.35", "1.35", 1, RoundHalfEven, "1.4"},
		{"half even 2.25", "2.25", 1, RoundHalfEven, "2.2"},
		{"half even 2.35", "2.35", 1, RoundHalfEven, "2.4"},

		// RoundUp (away from zero)
		{"round up positive", "1.21", 1, RoundUp, "1.3"},
		{"round up negative", "-1.21", 1, RoundUp, "-1.3"},

		// RoundDown (towards zero)
		{"round down positive", "1.29", 1, RoundDown, "1.2"},
		{"round down negative", "-1.29", 1, RoundDown, "-1.2"},

		// RoundCeiling
		{"ceiling positive", "1.21", 1, RoundCeiling, "1.3"},
		{"ceiling negative", "-1.29", 1, RoundCeiling, "-1.2"},

		// RoundFloor
		{"floor positive", "1.29", 1, RoundFloor, "1.2"},
		{"floor negative", "-1.21", 1, RoundFloor, "-1.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bcd, _ := New(tt.value)
			got := bcd.Round(tt.scale, tt.mode)
			verify.Equal(t, got.String(), tt.want)
		})
	}
}

func TestBCDConversion(t *testing.T) {
	t.Run("ToInt64", func(t *testing.T) {
		tests := []struct {
			value string
			want  int64
		}{
			{"123", 123},
			{"-456", -456},
			{"123.45", 123},
			{"123.99", 123},
			{"-123.99", -123},
		}

		for _, tt := range tests {
			bcd, _ := New(tt.value)
			got, err := bcd.ToInt64()
			verify.NoError(t, err)
			verify.Equal(t, got, tt.want)
		}
	})

	t.Run("ToFloat64", func(t *testing.T) {
		tests := []struct {
			value string
			want  float64
		}{
			{"123", 123.0},
			{"-456", -456.0},
			{"123.45", 123.45},
			{"0.000001", 0.000001},
		}

		for _, tt := range tests {
			bcd, _ := New(tt.value)
			got := bcd.ToFloat64()
			verify.About(t, got, tt.want, 0.0000001)
		}
	})
}

func TestAmount(t *testing.T) {
	t.Run("NewAmount", func(t *testing.T) {
		tests := []struct {
			amount string
			code   string
			want   string
			err    bool
		}{
			{"100", "USD", "$100.00", false},
			{"100.5", "USD", "$100.50", false},
			{"100.999", "USD", "$101.00", false}, // Rounds to 2 decimals
			{"1000", "JPY", "¥1000", false},      // No decimals for JPY
			{"100", "XXX", "", true},             // Unknown currency
		}

		for _, tt := range tests {
			curr, err := NewAmount(tt.amount, tt.code)
			if tt.err {
				verify.ErrorMatch(t, err, ".*")
			} else {
				verify.NoError(t, err)
				verify.Equal(t, curr.String(), tt.want)
			}
		}
	})

	t.Run("AmountAllocation", func(t *testing.T) {
		total, _ := NewAmount("100", "USD")

		// Split evenly
		parts, err := total.Split(3)
		verify.NoError(t, err)
		verify.Equal(t, len(parts), 3)

		// Check that sum equals original
		sum := Zero()
		for _, part := range parts {
			sum = sum.Add(part.Amount())
		}
		verify.True(t, sum.Equal(total.Amount()))

		// First part should get the extra penny
		verify.Equal(t, parts[0].String(), "$33.34")
		verify.Equal(t, parts[1].String(), "$33.33")
		verify.Equal(t, parts[2].String(), "$33.33")

		// Allocate by ratios
		parts, err = total.Allocate([]int{1, 2, 2})
		verify.NoError(t, err)
		verify.Equal(t, parts[0].String(), "$20.00")
		verify.Equal(t, parts[1].String(), "$40.00")
		verify.Equal(t, parts[2].String(), "$40.00")
	})

	t.Run("AmountArithmetic", func(t *testing.T) {
		usd100, _ := NewAmount("100", "USD")
		usd50, _ := NewAmount("50", "USD")
		eur100, _ := NewAmount("100", "EUR")

		// Add same currency
		sum, err := usd100.Add(usd50)
		verify.NoError(t, err)
		verify.Equal(t, sum.String(), "$150.00")

		// Add different currencies should error
		_, err = usd100.Add(eur100)
		verify.ErrorMatch(t, err, ".*currency mismatch.*")

		// Multiplication
		double := usd100.MulInt64(2)
		verify.Equal(t, double.String(), "$200.00")

		// Division
		half, err := usd100.DivInt64(2)
		verify.NoError(t, err)
		verify.Equal(t, half.String(), "$50.00")
	})

	t.Run("ParseAmount", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
			code  string
		}{
			{"$1,234.56", "$1234.56", "USD"},
			{"€1.234,56", "€1234.56", "EUR"},
			{"USD 1234.56", "$1234.56", "USD"},
			{"1,234.56 USD", "$1234.56", "USD"},
			{"¥1,234", "¥1234", "JPY"},
			{"($100.50)", "-$100.50", "USD"},
		}

		for _, tt := range tests {
			curr, err := ParseAmount(tt.input)
			verify.NoError(t, err)
			verify.Equal(t, curr.String(), tt.want)
			verify.Equal(t, curr.Code(), tt.code)
		}
	})
}

func TestBCDPrecisionMaintenance(t *testing.T) {
	// Test that floating point errors don't occur
	a, _ := New("0.1")
	b, _ := New("0.2")
	sum := a.Add(b)

	verify.Equal(t, sum.String(), "0.3")

	// Test repeated additions don't lose precision
	penny, _ := New("0.01")
	total := Zero()
	for range 100 {
		total = total.Add(penny)
	}

	verify.Equal(t, total.Normalize().String(), "1")

	// Test division precision
	t.Run("DivisionPrecision", func(t *testing.T) {
		one, _ := New("1")
		three, _ := New("3")
		third, _ := one.Div(three, 20, RoundHalfUp)

		// Multiply back
		result := third.Mul(three)
		result = result.Round(10, RoundHalfUp).Normalize()

		verify.Equal(t, result.String(), "1")
	})
}

func BenchmarkBCDAddition(b *testing.B) {
	x, _ := New("123.45")
	y, _ := New("678.90")

	for b.Loop() {
		_ = x.Add(y)
	}
}

func BenchmarkBCDMultiplication(b *testing.B) {
	x, _ := New("123.45")
	y, _ := New("678.90")

	for b.Loop() {
		_ = x.Mul(y)
	}
}

func BenchmarkBCDDivision(b *testing.B) {
	x, _ := New("123.45")
	y, _ := New("678.90")

	for b.Loop() {
		_, _ = x.Div(y, 10, RoundHalfUp)
	}
}
