// Copyright (c) 2024, Frank Mueller / Tideland
// All rights reserved.

package bcd

import (
	"math"
	"testing"
)

func TestNewBCD(t *testing.T) {
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
		{"invalid format", "12.34.56", "", true},
		{"invalid chars", "12a34", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.String() != tt.want {
				t.Errorf("New() = %v, want %v", got.String(), tt.want)
			}
		})
	}
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
			if err != nil {
				t.Fatalf("Failed to create BCD a: %v", err)
			}
			b, err := New(tt.b)
			if err != nil {
				t.Fatalf("Failed to create BCD b: %v", err)
			}

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
				if err != nil {
					t.Fatalf("Division error: %v", err)
				}
				// Simplify result for comparison
				result = result.Round(2, RoundHalfUp)
			}

			if result.String() != tt.want {
				t.Errorf("%s %s %s = %s, want %s", tt.a, tt.op, tt.b, result.String(), tt.want)
			}
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
			if got != tt.cmp {
				t.Errorf("Cmp(%s, %s) = %d, want %d", tt.a, tt.b, got, tt.cmp)
			}

			// Test comparison methods
			if tt.cmp < 0 && !a.LessThan(b) {
				t.Errorf("LessThan(%s, %s) = false, want true", tt.a, tt.b)
			}
			if tt.cmp == 0 && !a.Equal(b) {
				t.Errorf("Equal(%s, %s) = false, want true", tt.a, tt.b)
			}
			if tt.cmp > 0 && !a.GreaterThan(b) {
				t.Errorf("GreaterThan(%s, %s) = false, want true", tt.a, tt.b)
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
			if got.String() != tt.want {
				t.Errorf("Round(%s, %d, %v) = %s, want %s", 
					tt.value, tt.scale, tt.mode, got.String(), tt.want)
			}
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
			if err != nil {
				t.Errorf("ToInt64(%s) error = %v", tt.value, err)
				continue
			}
			if got != tt.want {
				t.Errorf("ToInt64(%s) = %d, want %d", tt.value, got, tt.want)
			}
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
			if math.Abs(got-tt.want) > 0.0000001 {
				t.Errorf("ToFloat64(%s) = %f, want %f", tt.value, got, tt.want)
			}
		}
	})
}

func TestCurrency(t *testing.T) {
	t.Run("NewCurrency", func(t *testing.T) {
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
			curr, err := NewCurrency(tt.amount, tt.code)
			if (err != nil) != tt.err {
				t.Errorf("NewCurrency(%s, %s) error = %v, wantErr %v", 
					tt.amount, tt.code, err, tt.err)
				continue
			}
			if err == nil && curr.String() != tt.want {
				t.Errorf("NewCurrency(%s, %s) = %s, want %s", 
					tt.amount, tt.code, curr.String(), tt.want)
			}
		}
	})

	t.Run("CurrencyArithmetic", func(t *testing.T) {
		usd100, _ := NewCurrency("100", "USD")
		usd50, _ := NewCurrency("50", "USD")
		eur100, _ := NewCurrency("100", "EUR")

		// Addition
		sum, err := usd100.Add(usd50)
		if err != nil {
			t.Errorf("Add error: %v", err)
		} else if sum.String() != "$150.00" {
			t.Errorf("100 USD + 50 USD = %s, want $150.00", sum.String())
		}

		// Currency mismatch
		_, err = usd100.Add(eur100)
		if err == nil {
			t.Error("Expected currency mismatch error")
		}

		// Multiplication
		double := usd100.MulInt64(2)
		if double.String() != "$200.00" {
			t.Errorf("100 USD * 2 = %s, want $200.00", double.String())
		}

		// Division
		half, err := usd100.DivInt64(2)
		if err != nil {
			t.Errorf("Div error: %v", err)
		} else if half.String() != "$50.00" {
			t.Errorf("100 USD / 2 = %s, want $50.00", half.String())
		}
	})

	t.Run("CurrencyAllocation", func(t *testing.T) {
		total, _ := NewCurrency("100", "USD")
		
		// Split evenly
		parts, err := total.Split(3)
		if err != nil {
			t.Errorf("Split error: %v", err)
		} else {
			if len(parts) != 3 {
				t.Errorf("Split returned %d parts, want 3", len(parts))
			}
			// Check that sum equals original
			sum := Zero()
			for _, part := range parts {
				sum = sum.Add(part.amount)
			}
			if !sum.Equal(total.amount) {
				t.Errorf("Split sum = %s, want %s", sum.String(), total.String())
			}
			// First part should get the extra penny
			if parts[0].String() != "$33.34" {
				t.Errorf("First part = %s, want $33.34", parts[0].String())
			}
			if parts[1].String() != "$33.33" {
				t.Errorf("Second part = %s, want $33.33", parts[1].String())
			}
		}

		// Allocate by ratios
		parts, err = total.Allocate([]int{1, 2, 2})
		if err != nil {
			t.Errorf("Allocate error: %v", err)
		} else {
			if parts[0].String() != "$20.00" {
				t.Errorf("First allocation = %s, want $20.00", parts[0].String())
			}
			if parts[1].String() != "$40.00" {
				t.Errorf("Second allocation = %s, want $40.00", parts[1].String())
			}
			if parts[2].String() != "$40.00" {
				t.Errorf("Third allocation = %s, want $40.00", parts[2].String())
			}
		}
	})

	t.Run("ParseCurrency", func(t *testing.T) {
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
			curr, err := ParseCurrency(tt.input)
			if err != nil {
				t.Errorf("ParseCurrency(%s) error: %v", tt.input, err)
				continue
			}
			if curr.String() != tt.want {
				t.Errorf("ParseCurrency(%s) = %s, want %s", tt.input, curr.String(), tt.want)
			}
			if curr.Code() != tt.code {
				t.Errorf("ParseCurrency(%s) code = %s, want %s", tt.input, curr.Code(), tt.code)
			}
		}
	})
}

func TestBCDPrecisionMaintenance(t *testing.T) {
	// Test that operations maintain precision correctly
	a, _ := New("0.1")
	b, _ := New("0.2")
	sum := a.Add(b)
	
	if sum.String() != "0.3" {
		t.Errorf("0.1 + 0.2 = %s, want 0.3", sum.String())
	}

	// Test repeated additions don't lose precision
	total := Zero()
	penny, _ := New("0.01")
	for range 100 {
		total = total.Add(penny)
	}
	
	if total.String() != "1" {
		t.Errorf("100 * 0.01 = %s, want 1", total.String())
	}

	// Test division precision
	one, _ := New("1")
	three, _ := New("3")
	third, _ := one.Div(three, 20, RoundHalfUp)
	
	// Multiply back
	result := third.Mul(three)
	result = result.Round(10, RoundHalfUp)
	
	if result.String() != "1" {
		t.Errorf("(1/3) * 3 = %s, want 1", result.String())
	}
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
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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